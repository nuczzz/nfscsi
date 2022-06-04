package pkg

import (
	"context"
	"log"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (nfs *nfsDriver) CreateVolume(context.Context, *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	log.Println("CreateVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) DeleteVolume(context.Context, *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	log.Println("DeleteVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) ControllerPublishVolume(context.Context, *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	log.Println("ControllerPublishVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) ControllerUnpublishVolume(context.Context, *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	log.Println("ControllerUnpublishVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) ValidateVolumeCapabilities(context.Context, *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	log.Println("ValidateVolumeCapabilities request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) ListVolumes(context.Context, *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	log.Println("ListVolumes request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) GetCapacity(context.Context, *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	log.Println("GetCapacity request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) ControllerGetCapabilities(context.Context, *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	log.Println("ControllerGetCapabilities request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) CreateSnapshot(context.Context, *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	log.Println("CreateSnapshot request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) DeleteSnapshot(context.Context, *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	log.Println("DeleteSnapshot request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) ListSnapshots(context.Context, *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	log.Println("ListSnapshots request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) ControllerExpandVolume(context.Context, *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	log.Println("ControllerExpandVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *nfsDriver) ControllerGetVolume(context.Context, *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	log.Println("ControllerGetVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}
