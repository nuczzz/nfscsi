package pkg

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	mount "k8s.io/mount-utils"
)

type Options struct {
	Name    string
	Version string
	NodeID  string

	NFSServer    string
	NFSRootPath  string
	NFSMountPath string
}

type NFSDriver struct {
	name    string
	version string
	nodeID  string

	nfsServer    string
	nfsRootPath  string
	nfsMountPath string

	mounter mount.Interface

	controllerServiceCapabilities []*csi.ControllerServiceCapability
	nodeServiceCapabilities       []*csi.NodeServiceCapability
}

var _ csi.IdentityServer = &NFSDriver{}
var _ csi.ControllerServer = &NFSDriver{}
var _ csi.NodeServer = &NFSDriver{}

func NewNFSDriver(opt *Options) *NFSDriver {
	nfs := &NFSDriver{
		name:         opt.Name,
		version:      opt.Version,
		nodeID:       opt.NodeID,
		nfsServer:    opt.NFSServer,
		nfsRootPath:  opt.NFSRootPath,
		nfsMountPath: opt.NFSMountPath,
		mounter:      mount.New(""),
	}

	nfs.addControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		//csi.ControllerServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
	})

	nfs.addNodeServiceCapabilities([]csi.NodeServiceCapability_RPC_Type{
		//csi.NodeServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
		//csi.NodeServiceCapability_RPC_UNKNOWN,
	})

	return nfs
}

func newControllerServiceCapability(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func (nfs *NFSDriver) addControllerServiceCapabilities(capabilities []csi.ControllerServiceCapability_RPC_Type) {
	var csc = make([]*csi.ControllerServiceCapability, 0, len(capabilities))
	for _, c := range capabilities {
		csc = append(csc, newControllerServiceCapability(c))
	}
	nfs.controllerServiceCapabilities = csc
}

func newNodeServiceCapability(cap csi.NodeServiceCapability_RPC_Type) *csi.NodeServiceCapability {
	return &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func (nfs *NFSDriver) addNodeServiceCapabilities(capabilities []csi.NodeServiceCapability_RPC_Type) {
	var nsc = make([]*csi.NodeServiceCapability, 0, len(capabilities))
	for _, n := range capabilities {
		nsc = append(nsc, newNodeServiceCapability(n))
	}
	nfs.nodeServiceCapabilities = nsc
}
