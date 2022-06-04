package pkg

import (
	"context"
	"log"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (nfs *nfsDriver) NodeStageVolume(context.Context, *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	log.Println("NodeStageVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) NodeUnstageVolume(context.Context, *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	log.Println("NodeUnstageVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) NodePublishVolume(context.Context, *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	log.Println("NodePublishVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) NodeUnpublishVolume(context.Context, *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	log.Println("NodeUnpublishVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) NodeGetVolumeStats(context.Context, *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	log.Println("NodeGetVolumeStats request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) NodeExpandVolume(context.Context, *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	log.Println("NodeExpandVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	log.Println("NodeGetCapabilities request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	log.Println("NodeGetInfo request")

	return &csi.NodeGetInfoResponse{
		NodeId: nfs.NodeID,
	}, nil
}
