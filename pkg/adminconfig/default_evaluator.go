// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	api "fybrik.io/fybrik/manager/apis/app/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

// DefaultConfig implements EvaluatorInterface
// It provides a default configuration as an alternative to evaluation of the written rego policies
type DefaultConfig struct {
	Manager *InfrastructureManager
	Data    *Infrastructure
}

// NewDefaultConfig constructs a new DefaultConfig object
func NewDefaultConfig() *DefaultConfig {
	return &DefaultConfig{Manager: nil, Data: nil}
}

// SetupWithInfrastructureManager connects the evaluator to the infrastructure manager for obtaining infrastructure details
func (r *DefaultConfig) SetupWithInfrastructureManager(mgr *InfrastructureManager) {
	r.Manager = mgr
	r.Data = nil
	// get infrastructure details using a new manager
	if data, err := mgr.SetInfrastructure(); err != nil {
		r.Data = data
	}
}

// Evaluate replaces hard-coded decisions in manager by default configuration
// The following logic is implemented:
/* 	Read capability is deployed in a read-type scenario.
    Read capability is deployed at the workload scope.
	Write capability is not yet supported and thus won’t be deployed.
	Transforms should always be performed close to the data.
	Copy is deployed if a user has requested it explicitly.
	Copy is deployed if there is no read module that supports the asset format.
	Copy is deployed in a read scenario if dataset resides in a different geography and governance actions are required.
*/
func (r *DefaultConfig) Evaluate(in *EvaluatorInput) (EvaluatorOutput, error) {
	if r.Data == nil {
		var err error
		if r.Data, err = r.Manager.SetInfrastructure(); err != nil {
			return EvaluatorOutput{Valid: false}, err
		}
	}
	decisions := map[api.CapabilityType]Decision{}
	// Read capability is deployed in a read-type scenario.
	deployRead := corev1.ConditionFalse
	if in.Request.Usage[api.ReadFlow] {
		deployRead = corev1.ConditionTrue
	}
	decisions[api.Read] = Decision{Deploy: deployRead,
		DeploymentRestrictions: Restrictions{
			Clusters:           []string{in.Workload.Cluster.Name},
			ModuleRestrictions: map[string]string{"capabilities.scope": "workload"},
		},
	}
	decisions[api.Write] = Decision{Deploy: corev1.ConditionFalse}

	copyDecision := DefaultDecision(r.Data)
	if in.Request.Usage[api.CopyFlow] {
		copyDecision.Deploy = corev1.ConditionTrue
	}

	clustersInRegion := []string{}
	for _, cluster := range r.Data.Clusters {
		if cluster.Metadata.Region == in.Request.Metadata.Geography {
			clustersInRegion = append(clustersInRegion, cluster.Name)
		}
	}
	if deployRead == corev1.ConditionTrue && len(in.GovernanceActions) > 0 && in.Workload.Cluster.Metadata.Region != in.Request.Metadata.Geography {
		copyDecision.Deploy = corev1.ConditionTrue
		copyDecision.DeploymentRestrictions.Clusters = clustersInRegion
	}

	transformDecision := DefaultDecision(r.Data)
	transformDecision.DeploymentRestrictions.Clusters = clustersInRegion

	decisions[api.Transform] = transformDecision
	decisions[api.Copy] = copyDecision

	return EvaluatorOutput{Valid: true, DatasetID: in.Request.DatasetID, ConfigDecisions: decisions}, nil
}
