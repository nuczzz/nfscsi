package pkg

import (
	"context"
	"log"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (nfs *nfsDriver) GetPluginInfo(context.Context, *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	log.Println("GetPluginInfo request")

	return &csi.GetPluginInfoResponse{
		Name:          nfs.Name,
		VendorVersion: nfs.Version,
	}, nil
}

func (nfs *nfsDriver) GetPluginCapabilities(context.Context, *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	log.Println("GetPluginCapabilities request")

	return nil, status.Error(codes.Unimplemented, "")
	//return &csi.GetPluginCapabilitiesResponse{
	//	Capabilities: []*csi.PluginCapability{
	//		{
	//			Type: &csi.PluginCapability_Service_{
	//				Service: &csi.PluginCapability_Service{
	//					// dynamic provisioning功能对应的是controllerService中的
	//					// CreateVolume和DeleteVolume方法
	//					Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
	//				},
	//			},
	//		},
	//	},
	//}, nil
}

func (nfs *nfsDriver) Probe(context.Context, *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	log.Println("Probe request")

	return nil, status.Error(codes.Unimplemented, "")
	//return &csi.ProbeResponse{
	//	Ready: &wrappers.BoolValue{
	//		Value: true,
	//	},
	//}, nil
}
