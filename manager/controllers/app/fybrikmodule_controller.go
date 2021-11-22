// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"

	"encoding/json"

	validate "fybrik.io/fybrik/pkg/taxonomy/validate"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "fybrik.io/fybrik/manager/apis/app/v1alpha1"
)

// FybrikModuleReconciler reconciles a FybrikModule object
type FybrikModuleReconciler struct {
	client.Client
	Name   string
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	ModuleTaxonomy = "/tmp/taxonomy/fybrik_module.json"
)

// Reconcile validates FybrikModule CRD
func (r *FybrikModuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("FybrikModule", req.NamespacedName)
	// obtain FybrikModule resource
	moduleContext := &api.FybrikModule{}
	if err := r.Get(ctx, req.NamespacedName, moduleContext); err != nil {
		log.V(0).Info("The reconciled object was not found")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// If the object has a scheduled deletion time, update status and return
	if !moduleContext.DeletionTimestamp.IsZero() {
		// The object is being deleted
		return ctrl.Result{}, nil
	}

	observedStatus := moduleContext.Status.DeepCopy()
	moduleVersion := moduleContext.GetGeneration()
	if len(moduleContext.Status.Conditions) == 0 {
		moduleContext.Status.Conditions = []api.Condition{{Type: api.ReadyCondition, Status: corev1.ConditionUnknown, ObservedGeneration: 0}}
	}

	// check if module has been validated before or if validated module is outdated
	condition := moduleContext.Status.Conditions[0]
	if condition.ObservedGeneration != moduleVersion || condition.Status == corev1.ConditionUnknown {
		// do validation on moduleContext
		err := ValidateFybrikModule(moduleContext, ModuleTaxonomy)
		condition.ObservedGeneration = moduleVersion
		// if validation fails
		if err != nil {
			// set error message
			log.V(0).Info("Fybrik module validation failed " + err.Error())
			condition.Message = err.Error()
			condition.Status = v1.ConditionFalse
		} else {
			condition.Status = v1.ConditionTrue
			condition.Message = ""
		}
		moduleContext.Status.Conditions[0] = condition
	}

	// Update CRD status in case of change (other than deletion, which was handled separately)
	if !equality.Semantic.DeepEqual(&moduleContext.Status, observedStatus) && moduleContext.DeletionTimestamp.IsZero() {
		log.V(0).Info("Reconcile: Updating status for desired generation " + fmt.Sprint(moduleContext.GetGeneration()))
		if err := r.Client.Status().Update(ctx, moduleContext); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// NewFybrikModuleReconciler creates a new reconciler for FybrikModules
func NewFybrikModuleReconciler(mgr ctrl.Manager, name string) *FybrikModuleReconciler {
	return &FybrikModuleReconciler{
		Client: mgr.GetClient(),
		Name:   name,
		Log:    ctrl.Log.WithName("controllers").WithName(name),
		Scheme: mgr.GetScheme(),
	}
}

func ValidateFybrikModule(module *api.FybrikModule, taxonomyFile string) error {
	var allErrs []*field.Error

	// Convert Fybrik module Go struct to JSON
	moduleJSON, err := json.Marshal(module)
	if err != nil {
		return err
	}
	// Validate Fybrik module against taxonomy
	allErrs = validate.TaxonomyCheck(moduleJSON, taxonomyFile, "Fybrik module")

	// Return any error
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "app.fybrik.io", Kind: "FybrikModule"},
		module.Name, allErrs)
}

// SetupWithManager registers Module controller
func (r *FybrikModuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.FybrikModule{}).
		Complete(r)
}
