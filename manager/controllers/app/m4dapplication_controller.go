// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"strings"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/hashicorp/vault/api"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/manager/controllers/app/modules"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	"github.com/ibm/the-mesh-for-data/pkg/multicluster"
	"github.com/ibm/the-mesh-for-data/pkg/storage"

	pc "github.com/ibm/the-mesh-for-data/pkg/policy-compiler/policy-compiler"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// OwnerLabelKey is a key to Labels map.
	// All owned resources should be labeled using this key.
	OwnerLabelKey string = "m4d.ibm.com/owner"
)

// M4DApplicationReconciler reconciles a M4DApplication object
type M4DApplicationReconciler struct {
	client.Client
	Name              string
	Log               logr.Logger
	Scheme            *runtime.Scheme
	VaultClient       *api.Client
	PolicyCompiler    pc.IPolicyCompiler
	ResourceInterface ContextInterface
	ClusterManager    multicluster.ClusterLister
	Provision         storage.ProvisionInterface
}

// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dapplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dapplications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=blueprints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=plotters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dmodules,verbs=get;list;watch

// +kubebuilder:rbac:groups=*,resources=*,verbs=*

// Reconcile reconciles M4DApplication CRD
// It receives M4DApplication CRD and selects the appropriate modules that will run
// The outcome is either a single Blueprint running on the same cluster or a Plotter containing multiple Blueprints that may run on different clusters
func (r *M4DApplicationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("m4dapplication", req.NamespacedName)
	// obtain M4DApplication resource
	applicationContext := &app.M4DApplication{}
	if err := r.Get(ctx, req.NamespacedName, applicationContext); err != nil {
		log.V(0).Info("The reconciled object was not found")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if err := r.reconcileFinalizers(applicationContext); err != nil {
		log.V(0).Info("Could not reconcile finalizers " + err.Error())
		return ctrl.Result{}, err
	}

	// If the object has a scheduled deletion time, update status and return
	if !applicationContext.DeletionTimestamp.IsZero() {
		// The object is being deleted
		return ctrl.Result{}, nil
	}

	observedStatus := applicationContext.Status.DeepCopy()

	// check if reconcile is required
	// reconcile is required if the spec has been changed, or the previous reconcile has failed to allocate a Blueprint or a Plotter resource
	generationComplete := r.ResourceInterface.ResourceExists(applicationContext.Status.Generated)
	if !generationComplete || observedStatus.ObservedGeneration != applicationContext.GetGeneration() {
		if result, err := r.reconcile(applicationContext); err != nil {
			// another attempt will be done
			// users should be informed in case of errors
			_ = r.deleteExternalResources(applicationContext)
			if !equality.Semantic.DeepEqual(&applicationContext.Status, observedStatus) {
				// ignore an update error, a new reconcile will be made in any case
				_ = r.Client.Status().Update(ctx, applicationContext)
			}
			return result, err
		}
		applicationContext.Status.ObservedGeneration = applicationContext.GetGeneration()
	} else {
		resourceStatus, err := r.ResourceInterface.GetResourceStatus(applicationContext.Status.Generated)
		if err != nil {
			return ctrl.Result{}, err
		}
		if err = r.checkReadiness(applicationContext, resourceStatus); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Update CRD status in case of change (other than deletion, which was handled separately)
	if !equality.Semantic.DeepEqual(&applicationContext.Status, observedStatus) && applicationContext.DeletionTimestamp.IsZero() {
		log.V(0).Info("Reconcile: Updating status for desired generation " + fmt.Sprint(applicationContext.GetGeneration()))
		if err := r.Client.Status().Update(ctx, applicationContext); err != nil {
			return ctrl.Result{}, err
		}
	}
	if hasError(applicationContext) {
		log.Info("Reconciled with errors: " + getErrorMessages(applicationContext))
	}
	return ctrl.Result{}, nil
}

func getBucketResourceRef(bucketName string) *types.NamespacedName {
	return &types.NamespacedName{Name: bucketName, Namespace: utils.GetSystemNamespace()}
}

func (r *M4DApplicationReconciler) checkReadiness(applicationContext *app.M4DApplication, status app.ObservedState) error {
	applicationContext.Status.DataAccessInstructions = ""
	applicationContext.Status.Ready = false
	if hasError(applicationContext) {
		return nil
	}
	if status.Error != "" {
		setCondition(applicationContext, "", status.Error, true)
		return nil
	}
	if !status.Ready {
		return nil
	}
	// Plotter is ready - update the M4DApplication status
	if applicationContext.Status.Ready {
		// nothing to be done
		return nil
	}
	// register assets if necessary if the ready state has been received
	for _, dataCtx := range applicationContext.Spec.Data {
		if dataCtx.Requirements.Copy.Catalog.CatalogID != "" {
			// TODO(shlomitk1) register the asset in the catalog
			// mark the bucket as persistent
			bucketName, found := applicationContext.Status.ProvisionedStorage[dataCtx.DataSetID]
			if !found {
				message := "No copy has been created for the asset " + dataCtx.DataSetID + " required to be registered"
				r.Log.V(0).Info(message)
				return errors.New(message)
			}
			if err := r.Provision.SetPersistent(getBucketResourceRef(bucketName), true); err != nil {
				return err
			}
		}
	}
	applicationContext.Status.Ready = true
	applicationContext.Status.DataAccessInstructions = status.DataAccessInstructions
	return nil
}

// reconcileFinalizers reconciles finalizers for M4DApplication
func (r *M4DApplicationReconciler) reconcileFinalizers(applicationContext *app.M4DApplication) error {
	// finalizer
	finalizerName := r.Name + ".finalizer"
	hasFinalizer := ctrlutil.ContainsFinalizer(applicationContext, finalizerName)

	// If the object has a scheduled deletion time, delete it and all resources it has created
	if !applicationContext.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if hasFinalizer { // Finalizer was created when the object was created
			// the finalizer is present - delete the allocated resources
			if err := r.deleteExternalResources(applicationContext); err != nil {
				return err
			}

			// remove the finalizer from the list and update it, because it needs to be deleted together with the object
			ctrlutil.RemoveFinalizer(applicationContext, finalizerName)

			if err := r.Update(context.Background(), applicationContext); err != nil {
				return err
			}
		}
		return nil
	}
	// Make sure this CRD instance has a finalizer
	if !hasFinalizer {
		ctrlutil.AddFinalizer(applicationContext, finalizerName)
		if err := r.Update(context.Background(), applicationContext); err != nil {
			return err
		}
	}
	return nil
}

func (r *M4DApplicationReconciler) deleteExternalResources(applicationContext *app.M4DApplication) error {
	// clear provisioned storage
	// References to buckets (Dataset resources) are deleted. Buckets that are persistent will not be removed upon Dataset deletion.
	var deletedBuckets []string
	var errMsgs []string
	for _, bucketName := range applicationContext.Status.ProvisionedStorage {
		if err := r.Provision.DeleteDataset(getBucketResourceRef(bucketName)); err != nil {
			errMsgs = append(errMsgs, err.Error())
		} else {
			deletedBuckets = append(deletedBuckets, bucketName)
		}
	}
	for _, bucket := range deletedBuckets {
		delete(applicationContext.Status.ProvisionedStorage, bucket)
	}
	if len(errMsgs) != 0 {
		return errors.New(strings.Join(errMsgs, ";"))
	}
	// delete the generated resource
	if applicationContext.Status.Generated == nil {
		return nil
	}

	r.Log.V(0).Info("Reconcile: M4DApplication is deleting the generated " + applicationContext.Status.Generated.Kind)
	if err := r.ResourceInterface.DeleteResource(applicationContext.Status.Generated); err != nil {
		return err
	}
	applicationContext.Status.Generated = nil
	return nil
}

// reconcile receives either M4DApplication CRD
// or a status update from the generated resource
func (r *M4DApplicationReconciler) reconcile(applicationContext *app.M4DApplication) (ctrl.Result, error) {
	utils.PrintStructure(applicationContext.Spec, r.Log, "M4DApplication")

	// Data User created or updated the M4DApplication

	// clear status
	resetConditions(applicationContext)
	applicationContext.Status.DataAccessInstructions = ""
	applicationContext.Status.Ready = false
	if applicationContext.Status.ProvisionedStorage == nil {
		applicationContext.Status.ProvisionedStorage = make(map[string]string, 0)
	}

	clusters, err := r.ClusterManager.GetClusters()
	if err != nil {
		return ctrl.Result{}, err
	}
	// create a list of requirements for creating a data flow (actions, interface to app, data format) per a single data set
	var requirements []modules.DataInfo
	for _, dataset := range applicationContext.Spec.Data {
		req := modules.DataInfo{
			Context: &dataset,
		}
		if err := r.constructDataInfo(&req, applicationContext, clusters); err != nil {
			return ctrl.Result{}, err
		}
		requirements = append(requirements, req)
	}
	// check for errors
	if hasError(applicationContext) {
		return ctrl.Result{}, nil
	}

	// create a module manager that will select modules to be orchestrated based on user requirements and module capabilities
	moduleMap, err := r.GetAllModules()
	if err != nil {
		return ctrl.Result{}, err
	}
	objectKey, _ := client.ObjectKeyFromObject(applicationContext)
	moduleManager := &ModuleManager{
		Client:             r.Client,
		Log:                r.Log,
		Modules:            moduleMap,
		Clusters:           clusters,
		Owner:              objectKey,
		PolicyCompiler:     r.PolicyCompiler,
		Provision:          r.Provision,
		VaultClient:        r.VaultClient,
		ProvisionedStorage: make(map[string]*storage.ProvisionedBucket, 0),
	}
	instances := make([]modules.ModuleInstanceSpec, 0)
	for _, item := range requirements {
		instancesPerDataset, err := moduleManager.SelectModuleInstances(item, applicationContext)
		if err != nil {
			setCondition(applicationContext, item.Context.DataSetID, err.Error(), true)
		}
		instances = append(instances, instancesPerDataset...)
	}
	// check for errors
	if hasError(applicationContext) {
		return ctrl.Result{}, nil
	}
	// update allocated storage in the status
	// clean irrelevant buckets
	for datasetID, bucketName := range applicationContext.Status.ProvisionedStorage {
		if _, found := moduleManager.ProvisionedStorage[datasetID]; !found {
			r.Provision.DeleteDataset(getBucketResourceRef(bucketName))
			delete(applicationContext.Status.ProvisionedStorage, datasetID)
		}
	}
	// add or update new buckets
	for datasetID, bucket := range moduleManager.ProvisionedStorage {
		applicationContext.Status.ProvisionedStorage[datasetID] = bucket.Name
	}
	ready := true
	var allocErr error
	// check that the buckets have been created successfully using Dataset status
	for id, bucketName := range applicationContext.Status.ProvisionedStorage {
		res, err := r.Provision.GetDatasetStatus(getBucketResourceRef(bucketName))
		if err != nil {
			ready = false
			break
		}
		if !res.Provisioned {
			ready = false
			r.Log.V(0).Info("No bucket has been provisioned for " + id)
			if res.ErrorMsg != "" {
				allocErr = errors.New(res.ErrorMsg)
			}
			break
		}
	}
	if !ready {
		return ctrl.Result{}, allocErr
	}
	// generate blueprint specifications (per cluster)
	blueprintPerClusterMap := r.GenerateBlueprints(instances, applicationContext)
	resourceRef := r.ResourceInterface.CreateResourceReference(applicationContext.Name, applicationContext.Namespace)
	ownerRef := &app.ResourceReference{Name: applicationContext.Name, Namespace: applicationContext.Namespace}
	if err := r.ResourceInterface.CreateOrUpdateResource(ownerRef, resourceRef, blueprintPerClusterMap); err != nil {
		r.Log.V(0).Info("Error creating " + resourceRef.Kind + " : " + err.Error())
		if err.Error() == app.InvalidClusterConfiguration {
			setCondition(applicationContext, "", app.InvalidClusterConfiguration, true)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	applicationContext.Status.Generated = resourceRef
	r.Log.V(0).Info("Created " + resourceRef.Kind + " successfully!")
	return ctrl.Result{}, nil
}

func (r *M4DApplicationReconciler) constructDataInfo(req *modules.DataInfo, input *app.M4DApplication, clusters []multicluster.Cluster) error {
	datasetID := req.Context.DataSetID
	var err error
	// Call the DataCatalog service to get info about the dataset
	if err = GetConnectionDetails(req, input); err != nil {
		return AnalyzeError(input, r.Log, datasetID, err)
	}
	// Call the CredentialsManager service to get info about the dataset
	if err = GetCredentials(req, input); err != nil {
		return AnalyzeError(input, r.Log, datasetID, err)
	}
	// The received credentials are stored in vault
	if err = r.RegisterCredentials(req); err != nil {
		return AnalyzeError(input, r.Log, datasetID, err)
	}
	return nil
}

// NewM4DApplicationReconciler creates a new reconciler for M4DApplications
func NewM4DApplicationReconciler(mgr ctrl.Manager, name string, vaultClient *api.Client,
	policyCompiler pc.IPolicyCompiler, cm multicluster.ClusterLister, provision storage.ProvisionInterface) *M4DApplicationReconciler {
	return &M4DApplicationReconciler{
		Client:            mgr.GetClient(),
		Name:              name,
		Log:               ctrl.Log.WithName("controllers").WithName(name),
		Scheme:            mgr.GetScheme(),
		VaultClient:       vaultClient,
		PolicyCompiler:    policyCompiler,
		ResourceInterface: NewPlotterInterface(mgr.GetClient()),
		ClusterManager:    cm,
		Provision:         provision,
	}
}

// SetupWithManager registers M4DApplication controller
func (r *M4DApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	mapFn := handler.ToRequestsFunc(
		func(a handler.MapObject) []reconcile.Request {
			labels := a.Meta.GetLabels()
			if labels == nil {
				return []reconcile.Request{}
			}
			label, ok := labels[OwnerLabelKey]
			namespaced := strings.Split(label, ".")
			if !ok || len(namespaced) != 2 {
				return []reconcile.Request{}
			}
			return []reconcile.Request{
				{NamespacedName: types.NamespacedName{
					Name:      namespaced[1],
					Namespace: namespaced[0],
				}},
			}
		})
	return ctrl.NewControllerManagedBy(mgr).
		For(&app.M4DApplication{}).
		Watches(&source.Kind{Type: r.ResourceInterface.GetManagedObject()},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: mapFn,
			}).Complete(r)
}

// AnalyzeError analyzes whether the given error is fatal, or a retrial attempt can be made.
// Reasons for retrial can be either communication problems with external services, or kubernetes problems to perform some action on a resource.
// A retrial is achieved by returning an error to the reconcile method
func AnalyzeError(app *app.M4DApplication, log logr.Logger, assetID string, err error) error {
	errStatus, _ := status.FromError(err)
	log.V(0).Info(errStatus.Message())
	if errStatus.Code() == codes.InvalidArgument {
		setCondition(app, assetID, errStatus.Message(), true)
		return nil
	}
	setCondition(app, assetID, errStatus.Message(), false)
	return err
}

func ownerLabels(id types.NamespacedName) map[string]string {
	return map[string]string{OwnerLabelKey: id.Namespace + "." + id.Name}
}

// GetAllModules returns all CRDs of the kind M4DModule mapped by their name
func (r *M4DApplicationReconciler) GetAllModules() (map[string]*app.M4DModule, error) {
	ctx := context.Background()

	moduleMap := make(map[string]*app.M4DModule)
	var moduleList app.M4DModuleList
	if err := r.List(ctx, &moduleList); err != nil {
		r.Log.V(0).Info("Error while listing modules: " + err.Error())
		return moduleMap, err
	}
	r.Log.Info("Listing all modules")
	for _, module := range moduleList.Items {
		r.Log.Info(module.GetName())
		moduleMap[module.Name] = module.DeepCopy()
	}
	return moduleMap, nil
}
