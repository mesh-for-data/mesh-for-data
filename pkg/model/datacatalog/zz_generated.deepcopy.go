// +build !ignore_autogenerated

// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

// Code generated by controller-gen. DO NOT EDIT.

package datacatalog

import (
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GetAssetRequest) DeepCopyInto(out *GetAssetRequest) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GetAssetRequest.
func (in *GetAssetRequest) DeepCopy() *GetAssetRequest {
	if in == nil {
		return nil
	}
	out := new(GetAssetRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GetAssetResponse) DeepCopyInto(out *GetAssetResponse) {
	*out = *in
	in.ResourceMetadata.DeepCopyInto(&out.ResourceMetadata)
	in.Details.DeepCopyInto(&out.Details)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GetAssetResponse.
func (in *GetAssetResponse) DeepCopy() *GetAssetResponse {
	if in == nil {
		return nil
	}
	out := new(GetAssetResponse)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceColumn) DeepCopyInto(out *ResourceColumn) {
	*out = *in
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = new(taxonomy.Tags)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceColumn.
func (in *ResourceColumn) DeepCopy() *ResourceColumn {
	if in == nil {
		return nil
	}
	out := new(ResourceColumn)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceDetails) DeepCopyInto(out *ResourceDetails) {
	*out = *in
	in.Connection.DeepCopyInto(&out.Connection)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceDetails.
func (in *ResourceDetails) DeepCopy() *ResourceDetails {
	if in == nil {
		return nil
	}
	out := new(ResourceDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceMetadata) DeepCopyInto(out *ResourceMetadata) {
	*out = *in
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = new(taxonomy.Tags)
		(*in).DeepCopyInto(*out)
	}
	if in.Columns != nil {
		in, out := &in.Columns, &out.Columns
		*out = make([]ResourceColumn, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceMetadata.
func (in *ResourceMetadata) DeepCopy() *ResourceMetadata {
	if in == nil {
		return nil
	}
	out := new(ResourceMetadata)
	in.DeepCopyInto(out)
	return out
}
