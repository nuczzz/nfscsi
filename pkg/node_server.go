package pkg

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	mount "k8s.io/mount-utils"
)

func (nfs *NFSDriver) NodeStageVolume(context.Context, *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	log.Println("NodeStageVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) NodeUnstageVolume(context.Context, *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	log.Println("NodeUnstageVolume request")

	return nil, status.Error(codes.Unimplemented, "")
}

func (nfs *NFSDriver) NodePublishVolume(_ context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	log.Println("NodePublishVolume request")

	capacity := req.GetVolumeCapability()
	if capacity == nil {
		return nil, errors.Errorf("capacity is nill")
	}

	options := capacity.GetMount().GetMountFlags()
	if req.Readonly {
		options = append(options, "ro")
	}

	targetPath := req.GetTargetPath()
	if targetPath == "" {
		return nil, errors.Errorf("target path is nill")
	}

	source := fmt.Sprintf("%s:%s", nfs.nfsServer, filepath.Join(nfs.nfsRootPath, req.GetVolumeId()))

	notMnt, err := nfs.mounter.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(targetPath, os.FileMode(0755)); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			notMnt = true
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	log.Printf("source: %s, targetPath: %s, options: %v", source, targetPath, options)

	if err := nfs.mounter.Mount(source, targetPath, "nfs", options); err != nil {
		return nil, errors.Wrap(err, "mount nfs path error")
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (nfs *NFSDriver) NodeUnpublishVolume(_ context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	log.Println("NodeUnpublishVolume request")

	targetPath := req.GetTargetPath()
	if err := mount.CleanupMountPoint(targetPath, nfs.mounter, true); err != nil {
		return nil, errors.Wrap(err, "clean mount point error")
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
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

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: nfs.nodeServiceCapabilities,
	}, nil
}

func (nfs *NFSDriver) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	log.Println("NodeGetInfo request")

	return &csi.NodeGetInfoResponse{
		NodeId: nfs.nodeID,
	}, nil
}
