// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true
type Asset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AssetSpec `json:"spec,omitempty"`
}

// AssetDetails defines model for AssetDetails.
type AssetDetails struct {
	// Connection information, runtime schema provided in taxonomy.json#/definitions/Connection
	Connection json.RawMessage `json:"connection"`
	// +optional
	// DataFormat information, runtime schema provided in taxonomy.json#/definitions/Dataformat
	DataFormat runtime.RawExtension `json:"dataFormat,omitempty"`
}

// AssetMetadata defines model for AssetMetadata.
type AssetMetadata struct {
	// Name of the data set
	Name string `json:"name"`
	// +optional
	Owner string `json:"owner,omitempty"`
	// +optional
	// The geography of the asset
	Geography string `json:"geography,omitempty"`
	// +optional
	// Tags associated with the asset, runtime schema provided in taxonomy.json#/definitions/Tags
	Tags json.RawMessage `json:"tags,omitempty"`
	// +optional
	Columns []Column `json:"columns,omitempty"`
}

type Column struct {
	Name string `json:"name"`
	// +optional
	// Tags associated with the column, runtime schema provided in taxonomy.json#/definitions/Tags
	Tags json.RawMessage `json:"tags,omitempty"`
}

// AssetSpec defines model for AssetSpec.
type AssetSpec struct {
	// Asset details
	AssetDetails  AssetDetails  `json:"details"`
	AssetMetadata AssetMetadata `json:"resource"`

	// This has the vault plugin path where the data credentials will be stored as kubernetes secrets.
	// This value is assumed to be known to the catalog connector.
	// Implements::credentials
	vaultPluginPath string `json:"vaultPluginPath"`
}
