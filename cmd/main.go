package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"

	"nfscsi/pkg"
)

var (
	endpoint   = flag.String("endpoint", "/csi/csi.sock", "CSI unix socket")
	driverName = flag.String("driver-name", "nfscsi", "CSI driver name")
	nodeId     = flag.String("nodeid", "nfs-csi-node", "node id of CSI")
	version    = flag.String("version", "N/A", "version of CSI")
	server     = flag.String("server", "", "nfs server ip")
	serverPath = flag.String("serverPath", "", "nfs server root mount path")
	mountPath  = flag.String("mountPath", "/mount", "local mount path")
)

func main() {
	flag.Parse()

	log.Printf("driverName: %s, version: %s, nodeID: %s", *driverName, *version, *nodeId)

	// remove endpoint before start
	if err := os.Remove(*endpoint); err != nil && !os.IsNotExist(err) {
		log.Fatalf("remove endpoint %s error", *endpoint)
	}

	ln, err := net.Listen("unix", *endpoint)
	if err != nil {
		log.Fatalf("listen unix endpoint %s error: %s", *endpoint, err.Error())
	}
	defer func() {
		_ = ln.Close()
		_ = os.Remove(*endpoint)
	}()

	grpcServer := grpc.NewServer()

	nfsDriver := pkg.NewNFSDriver(&pkg.Options{
		Name:         *driverName,
		Version:      *version,
		NodeID:       *nodeId,
		NFSServer:    *server,
		NFSRootPath:  *serverPath,
		NFSMountPath: *mountPath,
	})

	csi.RegisterIdentityServer(grpcServer, nfsDriver)
	csi.RegisterControllerServer(grpcServer, nfsDriver)
	csi.RegisterNodeServer(grpcServer, nfsDriver)

	log.Println("grpc server start")
	defer log.Println("grpc server exit")

	if err = grpcServer.Serve(ln); err != nil {
		log.Fatalf("grpc serve error: %s", err.Error())
	}
}
