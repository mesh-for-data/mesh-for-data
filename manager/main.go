// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

//nolint:revive
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/fsnotify/fsnotify"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	fapp "fybrik.io/fybrik/manager/apis/app/v1"
	"fybrik.io/fybrik/manager/controllers"
	"fybrik.io/fybrik/manager/controllers/app"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/adminconfig"
	dcclient "fybrik.io/fybrik/pkg/connectors/datacatalog/clients"
	pmclient "fybrik.io/fybrik/pkg/connectors/policymanager/clients"
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/helm"
	"fybrik.io/fybrik/pkg/infrastructure"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/monitor"
	"fybrik.io/fybrik/pkg/multicluster"
	"fybrik.io/fybrik/pkg/multicluster/local"
	"fybrik.io/fybrik/pkg/multicluster/razee"
	"fybrik.io/fybrik/pkg/storage"
)

const certSubDir = "/k8s-webhook-server"

var (
	gitCommit string
	gitTag    string
	scheme    = kruntime.NewScheme()
	setupLog  = logging.LogInit(logging.SETUP, "main")
)

func init() {
	_ = fapp.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = coordinationv1.AddToScheme(scheme)
}

//nolint:funlen,gocyclo
func run(namespace string, metricsAddr string, enableLeaderElection bool,
	enableApplicationController, enableBlueprintController, enablePlotterController bool) int {
	setupLog.Info().Msg("creating manager. based on: gitTag=" + gitTag + ", latest gitCommit=" + gitCommit)
	environment.LogEnvVariables(&setupLog)

	var applicationNamespaceSelector fields.Selector
	applicationNamespace := environment.GetApplicationNamespace()
	if len(applicationNamespace) > 0 {
		applicationNamespaceSelector = fields.SelectorFromSet(fields.Set{"metadata.namespace": applicationNamespace})
	}
	setupLog.Info().Msg("Application namespace: " + applicationNamespace)

	systemNamespaceSelector := fields.SelectorFromSet(fields.Set{"metadata.namespace": environment.GetSystemNamespace()})
	selectorsByObject := cache.SelectorsByObject{
		&fapp.FybrikApplication{}:    {Field: applicationNamespaceSelector},
		&fapp.Plotter{}:              {Field: systemNamespaceSelector},
		&fapp.FybrikModule{}:         {Field: systemNamespaceSelector},
		&fapp.FybrikStorageAccount{}: {Field: systemNamespaceSelector},
		&corev1.ConfigMap{}:          {Field: systemNamespaceSelector},
		&fapp.Blueprint{}:            {Field: systemNamespaceSelector},
		&corev1.Secret{}:             {Field: systemNamespaceSelector},
	}

	client := ctrl.GetConfigOrDie()
	client.QPS = environment.GetEnvAsFloat32(controllers.KubernetesClientQPSConfiguration, controllers.DefaultKubernetesClientQPS)
	client.Burst = environment.GetEnvAsInt(controllers.KubernetesClientBurstConfiguration, controllers.DefaultKubernetesClientBurst)

	setupLog.Info().Msg("Manager client rate limits: qps = " + fmt.Sprint(client.QPS) + " burst=" + fmt.Sprint(client.Burst))

	mgr, err := ctrl.NewManager(client, ctrl.Options{
		CertDir:            environment.GetDataDir() + certSubDir,
		Scheme:             scheme,
		Namespace:          namespace,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "fybrik-operator-leader-election",
		Port:               controllers.ManagerPort,
		NewCache:           cache.BuilderWithOptions(cache.Options{SelectorsByObject: selectorsByObject}),
	})
	if err != nil {
		setupLog.Error().Err(err).Msg("unable to start manager")
		return 1
	}

	// Initialize ClusterManager
	setupLog.Trace().Msg("creating cluster manager")
	var clusterManager multicluster.ClusterManager
	if enableApplicationController || enablePlotterController {
		clusterManager, err = newClusterManager(mgr)
		if err != nil {
			setupLog.Error().Err(err).Msg("unable to initialize cluster manager")
			return 1
		}
	}

	if enableApplicationController {
		setupLog.Trace().Msg("creating FybrikApplication controller")

		// Initialize PolicyManager interface
		policyManager, err := newPolicyManager(scheme)
		if err != nil {
			setupLog.Error().Err(err).Str(logging.CONTROLLER, "FybrikApplication").Msg("unable to create policy manager facade")
			return 1
		}
		defer func() {
			if err = policyManager.Close(); err != nil {
				setupLog.Error().Err(err).Str(logging.CONTROLLER, "FybrikApplication").Msg("unable to close policy manager facade")
			}
		}()

		// Initialize DataCatalog interface
		catalog, err := newDataCatalog(scheme)
		if err != nil {
			setupLog.Error().Err(err).Str(logging.CONTROLLER, "FybrikApplication").Msg("unable to create data catalog facade")
			return 1
		}
		defer func() {
			if err = catalog.Close(); err != nil {
				setupLog.Error().Err(err).Str(logging.CONTROLLER, "FybrikApplication").Msg("unable to close data catalog facade")
			}
		}()

		evaluator, err := adminconfig.NewRegoPolicyEvaluator()
		if err != nil {
			setupLog.Error().Err(err).Str(logging.CONTROLLER, "FybrikApplication").Msg("unable to compile configuration policies")
			return 1
		}
		infrastructureManager, err := infrastructure.NewAttributeManager()
		if err != nil {
			setupLog.Error().Err(err).Str(logging.CONTROLLER, "FybrikApplication").Msg("unable to get infrastructure attributes")
			return 1
		}

		// Initiate the FybrikApplication Controller
		applicationController := app.NewFybrikApplicationReconciler(
			mgr,
			"FybrikApplication",
			policyManager,
			catalog,
			clusterManager,
			storage.NewProvisionImpl(mgr.GetClient()),
			evaluator,
			infrastructureManager,
		)
		if err = applicationController.SetupWithManager(mgr); err != nil {
			setupLog.Error().Err(err).Str(logging.CONTROLLER, "FybrikApplication").Msg("unable to create controller")
			return 1
		}
		if os.Getenv("ENABLE_WEBHOOKS") != "false" {
			if err = (&fapp.FybrikApplication{}).SetupWebhookWithManager(mgr); err != nil {
				setupLog.Error().Err(err).Str(logging.WEBHOOK, "FybrikApplication").Msg("unable to create webhook")
				return 1
			}
			if err = (&fapp.FybrikModule{}).SetupWebhookWithManager(mgr); err != nil {
				setupLog.Error().Err(err).Str(logging.WEBHOOK, "FybrikModule").Msg("unable to create webhook")
				return 1
			}
		}

		// monitor changes in config policies and attributes
		fileMonitor := &monitor.FileMonitor{Subsciptions: []monitor.Subscription{}, Log: setupLog}
		if err = fileMonitor.Subscribe(evaluator); err != nil {
			setupLog.Error().Err(err).Str(logging.CONTROLLER, "FybrikApplication").Msg("unable to monitor policy changes")
		}
		if err = fileMonitor.Subscribe(infrastructureManager); err != nil {
			setupLog.Error().Err(err).Str(logging.CONTROLLER, "FybrikApplication").Msg("unable to monitor attribute changes")
		}
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			setupLog.Err(err).Msg("error creating a file system watcher")
			return 1
		}
		defer watcher.Close()
		// watch $DATA_DIR/adminconfig directory for changes
		err = watcher.Add(adminconfig.RegoPolicyDirectory)
		if err != nil {
			setupLog.Err(err).Msg("error adding a directory to monitor")
			return 1
		}

		fileMonitor.Run(watcher)
		// Initiate the FybrikModule Controller
		moduleController := app.NewFybrikModuleReconciler(
			mgr,
			"FybrikModule",
		)
		if err := moduleController.SetupWithManager(mgr); err != nil {
			setupLog.Error().Err(err).Str(logging.CONTROLLER, "FybrikModule").Msg("unable to create controller")
			return 1
		}
	}

	if enablePlotterController {
		// Initiate the Plotter Controller
		setupLog.Trace().Msg("creating Plotter controller")
		plotterController := app.NewPlotterReconciler(mgr, "Plotter", clusterManager)
		if err := plotterController.SetupWithManager(mgr); err != nil {
			setupLog.Error().Err(err).Str(logging.CONTROLLER, "Plotter").Msg("unable to create controller " + plotterController.Name)
			return 1
		}
	}

	if enableBlueprintController {
		// Initiate the Blueprint Controller
		localMountPath := os.Getenv("LOCAL_CHARTS_DIR")
		setupLog.Trace().Str("local charts dir", localMountPath).Msg("creating Blueprint controller")
		blueprintController := app.NewBlueprintReconciler(mgr, "Blueprint", helm.NewHelmerImpl(localMountPath))
		if err := blueprintController.SetupWithManager(mgr); err != nil {
			setupLog.Error().Err(err).Str(logging.CONTROLLER, "Blueprint").Msg("unable to create controller " + blueprintController.Name)
			return 1
		}
	}

	setupLog.Trace().Msg("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error().Err(err).Str(logging.SETUP, "main").Msg("problem running manager")
		return 1
	}

	return 0
}

// Main entry point starts manager and controllers
func main() {
	var namespace string
	var metricsAddr string
	var enableLeaderElection bool
	var enableApplicationController bool
	var enableBlueprintController bool
	var enablePlotterController bool
	var enableAllControllers bool
	address := utils.ListeningAddress(controllers.ListeningPortAddress)

	flag.StringVar(&metricsAddr, "metrics-bind-addr", address, "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&enableApplicationController, "enable-application-controller", false,
		"Enable application controller of the manager. This manages CRDs of type FybrikApplication.")
	flag.BoolVar(&enableBlueprintController, "enable-blueprint-controller", false,
		"Enable blueprint controller of the manager. This manages CRDs of type Blueprint.")
	flag.BoolVar(&enablePlotterController, "enable-plotter-controller", false,
		"Enable plotter controller of the manager. This manages CRDs of type Plotter.")
	flag.BoolVar(&enableAllControllers, "enable-all-controllers", false,
		"Enables all controllers.")
	flag.StringVar(&namespace, "namespace", "", "The namespace to which this controller manager is limited.")
	flag.Parse()

	if enableAllControllers {
		enableApplicationController = true
		enableBlueprintController = true
		enablePlotterController = true
	}

	if !enableApplicationController && !enablePlotterController && !enableBlueprintController {
		setupLog.Debug().Msg("At least one controller flag must be set!")
		os.Exit(1)
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	os.Exit(run(namespace, metricsAddr, enableLeaderElection,
		enableApplicationController, enableBlueprintController, enablePlotterController))
}

func newDataCatalog(schema *kruntime.Scheme) (dcclient.DataCatalog, error) {
	connectionTimeout, err := getConnectionTimeout()
	if err != nil {
		return nil, err
	}
	providerName := os.Getenv("CATALOG_PROVIDER_NAME")
	connectorURL := os.Getenv("CATALOG_CONNECTOR_URL")
	setupLog.Info().Str("Name", providerName).Str("URL", connectorURL).
		Str("Timeout", connectionTimeout.String()).Msg("setting data catalog client")
	return dcclient.NewDataCatalog(
		providerName,
		connectorURL,
		connectionTimeout,
		schema,
	)
}

func newPolicyManager(schema *kruntime.Scheme) (pmclient.PolicyManager, error) {
	connectionTimeout, err := getConnectionTimeout()
	if err != nil {
		return nil, err
	}

	mainPolicyManagerName := os.Getenv("MAIN_POLICY_MANAGER_NAME")
	mainPolicyManagerURL := os.Getenv("MAIN_POLICY_MANAGER_CONNECTOR_URL")
	setupLog.Info().Str("Name", mainPolicyManagerName).Str("URL", mainPolicyManagerURL).
		Str("Timeout", connectionTimeout.String()).Msg("setting main policy manager client")

	var policyManager pmclient.PolicyManager
	if strings.HasPrefix(mainPolicyManagerURL, "http") {
		policyManager, err = pmclient.NewOpenAPIPolicyManager(
			mainPolicyManagerName,
			mainPolicyManagerURL,
			connectionTimeout,
			schema,
		)
	} else {
		policyManager, err = pmclient.NewGrpcPolicyManager(mainPolicyManagerName, mainPolicyManagerURL, connectionTimeout)
	}

	return policyManager, err
}

// newClusterManager decides based on the environment variables that are set which
// cluster manager instance should be initiated.
func newClusterManager(mgr manager.Manager) (multicluster.ClusterManager, error) {
	multiClusterGroup := os.Getenv("MULTICLUSTER_GROUP")
	if user, razeeLocal := os.LookupEnv("RAZEE_USER"); razeeLocal {
		razeeURL := strings.TrimSpace(os.Getenv("RAZEE_URL"))
		password := strings.TrimSpace(os.Getenv("RAZEE_PASSWORD"))

		setupLog.Info().Msg("Using razee local at " + razeeURL)
		return razee.NewRazeeLocalClusterManager(strings.TrimSpace(razeeURL), strings.TrimSpace(user), password, multiClusterGroup)
	} else if apiKey, satConf := os.LookupEnv("IAM_API_KEY"); satConf {
		setupLog.Info().Msg("Using IBM Satellite config")
		return razee.NewSatConfClusterManager(strings.TrimSpace(apiKey), multiClusterGroup)
	} else if apiKey, razeeOauth := os.LookupEnv("API_KEY"); razeeOauth {
		setupLog.Info().Msg("Using Razee oauth")

		razeeURL := strings.TrimSpace(os.Getenv("RAZEE_URL"))
		return razee.NewRazeeOAuthClusterManager(strings.TrimSpace(razeeURL), strings.TrimSpace(apiKey), multiClusterGroup)
	} else {
		setupLog.Info().Msg("Using local cluster manager")
		return local.NewClusterManager(mgr.GetClient(), environment.GetSystemNamespace())
	}
}

func getConnectionTimeout() (time.Duration, error) {
	connectionTimeout := os.Getenv("CONNECTION_TIMEOUT")
	timeOutInSeconds, err := strconv.Atoi(connectionTimeout)
	if err != nil {
		return 0, errors.Wrap(err, "Atoi conversion of CONNECTION_TIMEOUT failed")
	}
	return time.Duration(timeOutInSeconds) * time.Second, nil
}
