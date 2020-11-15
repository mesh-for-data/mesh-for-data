// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: data_catalog_request.proto

package connectors

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// DeepCopyInto supports using CatalogDatasetRequest within kubernetes types, where deepcopy-gen is used.
func (in *CatalogDatasetRequest) DeepCopyInto(out *CatalogDatasetRequest) {
	p := proto.Clone(in).(*CatalogDatasetRequest)
	*out = *p
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CatalogDatasetRequest. Required by controller-gen.
func (in *CatalogDatasetRequest) DeepCopy() *CatalogDatasetRequest {
	if in == nil {
		return nil
	}
	out := new(CatalogDatasetRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInterface is an autogenerated deepcopy function, copying the receiver, creating a new CatalogDatasetRequest. Required by controller-gen.
func (in *CatalogDatasetRequest) DeepCopyInterface() interface{} {
	return in.DeepCopy()
}
