package pkg

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
)

type CSI interface {
	csi.IdentityServer
	csi.ControllerServer
	csi.NodeServer
}

type nfsDriver struct {
	Name    string
	Version string
	NodeID  string
}

var _ CSI = &nfsDriver{}

func NewNfsDriver(name, version, nodeID string) CSI {
	return &nfsDriver{
		Name:    name,
		Version: version,
		NodeID:  nodeID,
	}
}
