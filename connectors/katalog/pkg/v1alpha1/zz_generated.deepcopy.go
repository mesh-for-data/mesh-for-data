// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Asset) DeepCopyInto(out *Asset) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Asset.
func (in *Asset) DeepCopy() *Asset {
	if in == nil {
		return nil
	}
	out := new(Asset)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Asset) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AssetSpec) DeepCopyInto(out *AssetSpec) {
	*out = *in
	in.AssetDetails.DeepCopyInto(&out.AssetDetails)
	in.AssetMetadata.DeepCopyInto(&out.AssetMetadata)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AssetSpec.
func (in *AssetSpec) DeepCopy() *AssetSpec {
	if in == nil {
		return nil
	}
	out := new(AssetSpec)
	in.DeepCopyInto(out)
	return out
}
