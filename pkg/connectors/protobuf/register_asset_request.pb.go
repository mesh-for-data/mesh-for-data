// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.1
// source: register_asset_request.proto

package protobuf

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type RegisterAssetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Creds                *Credentials    `protobuf:"bytes,1,opt,name=creds,proto3" json:"creds,omitempty"`
	DatasetDetails       *DatasetDetails `protobuf:"bytes,2,opt,name=dataset_details,json=datasetDetails,proto3" json:"dataset_details,omitempty"`
	DestinationCatalogId string          `protobuf:"bytes,3,opt,name=destination_catalog_id,json=destinationCatalogId,proto3" json:"destination_catalog_id,omitempty"`
	CredentialPath       string          `protobuf:"bytes,4,opt,name=credential_path,json=credentialPath,proto3" json:"credential_path,omitempty"` // link to vault plugin for reading k8s secret with user credentials
}

func (x *RegisterAssetRequest) Reset() {
	*x = RegisterAssetRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_register_asset_request_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterAssetRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterAssetRequest) ProtoMessage() {}

func (x *RegisterAssetRequest) ProtoReflect() protoreflect.Message {
	mi := &file_register_asset_request_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterAssetRequest.ProtoReflect.Descriptor instead.
func (*RegisterAssetRequest) Descriptor() ([]byte, []int) {
	return file_register_asset_request_proto_rawDescGZIP(), []int{0}
}

func (x *RegisterAssetRequest) GetCreds() *Credentials {
	if x != nil {
		return x.Creds
	}
	return nil
}

func (x *RegisterAssetRequest) GetDatasetDetails() *DatasetDetails {
	if x != nil {
		return x.DatasetDetails
	}
	return nil
}

func (x *RegisterAssetRequest) GetDestinationCatalogId() string {
	if x != nil {
		return x.DestinationCatalogId
	}
	return ""
}

func (x *RegisterAssetRequest) GetCredentialPath() string {
	if x != nil {
		return x.CredentialPath
	}
	return ""
}

var File_register_asset_request_proto protoreflect.FileDescriptor

var file_register_asset_request_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x61, 0x73, 0x73, 0x65, 0x74,
	0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x1a, 0x11, 0x63, 0x72, 0x65, 0x64,
	0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x15, 0x64,
	0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x5f, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xe9, 0x01, 0x0a, 0x14, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65,
	0x72, 0x41, 0x73, 0x73, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2d, 0x0a,
	0x05, 0x63, 0x72, 0x65, 0x64, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x2e, 0x43, 0x72, 0x65, 0x64, 0x65, 0x6e,
	0x74, 0x69, 0x61, 0x6c, 0x73, 0x52, 0x05, 0x63, 0x72, 0x65, 0x64, 0x73, 0x12, 0x43, 0x0a, 0x0f,
	0x64, 0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x5f, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x73, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c,
	0x73, 0x52, 0x0e, 0x64, 0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c,
	0x73, 0x12, 0x34, 0x0a, 0x16, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x5f, 0x63, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x14, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x61,
	0x74, 0x61, 0x6c, 0x6f, 0x67, 0x49, 0x64, 0x12, 0x27, 0x0a, 0x0f, 0x63, 0x72, 0x65, 0x64, 0x65,
	0x6e, 0x74, 0x69, 0x61, 0x6c, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0e, 0x63, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x50, 0x61, 0x74, 0x68,
	0x42, 0x47, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x2e, 0x64, 0x61, 0x74, 0x6d, 0x65, 0x73, 0x68, 0x5a,
	0x38, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x69, 0x62, 0x6d, 0x2f,
	0x74, 0x68, 0x65, 0x2d, 0x6d, 0x65, 0x73, 0x68, 0x2d, 0x66, 0x6f, 0x72, 0x2d, 0x64, 0x61, 0x74,
	0x61, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x73,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_register_asset_request_proto_rawDescOnce sync.Once
	file_register_asset_request_proto_rawDescData = file_register_asset_request_proto_rawDesc
)

func file_register_asset_request_proto_rawDescGZIP() []byte {
	file_register_asset_request_proto_rawDescOnce.Do(func() {
		file_register_asset_request_proto_rawDescData = protoimpl.X.CompressGZIP(file_register_asset_request_proto_rawDescData)
	})
	return file_register_asset_request_proto_rawDescData
}

var file_register_asset_request_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_register_asset_request_proto_goTypes = []interface{}{
	(*RegisterAssetRequest)(nil), // 0: connectors.RegisterAssetRequest
	(*Credentials)(nil),          // 1: connectors.Credentials
	(*DatasetDetails)(nil),       // 2: connectors.DatasetDetails
}
var file_register_asset_request_proto_depIdxs = []int32{
	1, // 0: connectors.RegisterAssetRequest.creds:type_name -> connectors.Credentials
	2, // 1: connectors.RegisterAssetRequest.dataset_details:type_name -> connectors.DatasetDetails
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_register_asset_request_proto_init() }
func file_register_asset_request_proto_init() {
	if File_register_asset_request_proto != nil {
		return
	}
	file_credentials_proto_init()
	file_dataset_details_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_register_asset_request_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterAssetRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_register_asset_request_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_register_asset_request_proto_goTypes,
		DependencyIndexes: file_register_asset_request_proto_depIdxs,
		MessageInfos:      file_register_asset_request_proto_msgTypes,
	}.Build()
	File_register_asset_request_proto = out.File
	file_register_asset_request_proto_rawDesc = nil
	file_register_asset_request_proto_goTypes = nil
	file_register_asset_request_proto_depIdxs = nil
}
