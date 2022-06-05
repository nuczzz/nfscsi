package pkg

import (
	"context"
	"log"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewNodeDriver(nfsDriver *NFSDriver) csi.NodeServer {
	return csi.NodeServer(nfsDriver)
}

func (nfs *NFSDriver) NodeStageVolume(context.Context, *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	log.Println("NodeStageVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) NodeUnstageVolume(context.Context, *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	log.Println("NodeUnstageVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) NodePublishVolume(context.Context, *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	log.Println("NodePublishVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) NodeUnpublishVolume(context.Context, *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	log.Println("NodeUnpublishVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) NodeGetVolumeStats(context.Context, *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	log.Println("NodeGetVolumeStats request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) NodeExpandVolume(context.Context, *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	log.Println("NodeExpandVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	log.Println("NodeGetCapabilities request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	log.Println("NodeGetInfo request")

	return &csi.NodeGetInfoResponse{
		NodeId: nfs.nodeID,
	}, nil
}
