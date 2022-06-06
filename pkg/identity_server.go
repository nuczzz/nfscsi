package pkg

import (
	"context"
	"log"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/wrappers"
)

func (nfs *NFSDriver) GetPluginInfo(context.Context, *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	log.Println("GetPluginInfo request")

	return &csi.GetPluginInfoResponse{
		Name:          nfs.name,
		VendorVersion: nfs.version,
	}, nil
}

func (nfs *NFSDriver) GetPluginCapabilities(context.Context, *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	log.Println("GetPluginCapabilities request")

	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						// dynamic provisioning功能对应的是controllerService中的
						// CreateVolume和DeleteVolume方法
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}, nil
}

func (nfs *NFSDriver) Probe(context.Context, *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	log.Println("Probe request")

	return &csi.ProbeResponse{
		Ready: &wrappers.BoolValue{
			Value: true,
		},
	}, nil
}
