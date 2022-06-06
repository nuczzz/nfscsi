package pkg

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (nfs *NFSDriver) CreateVolume(_ context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	log.Println("CreateVolume request")

	log.Println("req name: ", req.GetName())
	mountPath := filepath.Join(nfs.nfsMountPath, req.GetName())
	if err := os.Mkdir(mountPath, 0755); err != nil {
		log.Printf("mkdir %s error: %s", mountPath, err.Error())
		return nil, errors.Wrap(err, "mkdir error")
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      req.Name,
			CapacityBytes: 0,
		},
	}, nil
}

func (nfs *NFSDriver) DeleteVolume(_ context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	log.Println("DeleteVolume request")

	log.Println("volumeID: ", req.GetVolumeId())

	return nil, os.Remove(filepath.Join(nfs.nfsMountPath, req.GetVolumeId()))
}

func (nfs *NFSDriver) ControllerPublishVolume(context.Context, *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	log.Println("ControllerPublishVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) ControllerUnpublishVolume(context.Context, *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	log.Println("ControllerUnpublishVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) ValidateVolumeCapabilities(context.Context, *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	log.Println("ValidateVolumeCapabilities request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) ListVolumes(context.Context, *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	log.Println("ListVolumes request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) GetCapacity(context.Context, *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	log.Println("GetCapacity request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) ControllerGetCapabilities(context.Context, *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	log.Println("ControllerGetCapabilities request")

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: nfs.controllerServiceCapabilities,
	}, nil
}

func (nfs *NFSDriver) CreateSnapshot(context.Context, *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	log.Println("CreateSnapshot request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) DeleteSnapshot(context.Context, *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	log.Println("DeleteSnapshot request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) ListSnapshots(context.Context, *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	log.Println("ListSnapshots request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) ControllerExpandVolume(context.Context, *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	log.Println("ControllerExpandVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) ControllerGetVolume(context.Context, *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	log.Println("ControllerGetVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}
