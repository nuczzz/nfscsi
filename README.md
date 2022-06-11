前面几篇文章分别介绍了`dynamic provisioning`、`CSI接口定义`和`CSI插件的注册`等内容，这篇文章基于这些内容，尝试实现一个NFS的CSI，这个CSI主要包含`注册`、`dynamic provisioning`和`Mount`功能。

>本文示例代码：https://github.com/nuczzz/nfscsi

# CSI接口定义

[《kubernetes CSI（上）》](https://mp.weixin.qq.com/s/VnbXGMt8hUPiJGQTu3BNTg)一文中我们介绍了CSI三个grpc service接口的定义，因这部分内容与本文关系密切，我们先大概回顾下。

实现一个CSI需要实现三个grpc service，这三个grpc service的定义如下：

- **IdentotyServer**
```go
// github.com/container-storage-interface/spec/lib/go/csi/csi.pb.go
type IdentityServer interface {
    GetPluginInfo(context.Context, *GetPluginInfoRequest) (*GetPluginInfoResponse, error)
    GetPluginCapabilities(context.Context, *GetPluginCapabilitiesRequest) (*GetPluginCapabilitiesResponse, error)
    Probe(context.Context, *ProbeRequest) (*ProbeResponse, error)
}
```
- **ControllerServer**
```go
// github.com/container-storage-interface/spec/lib/go/csi/csi.pb.go
type ControllerServer interface {
    CreateVolume(context.Context, *CreateVolumeRequest) (*CreateVolumeResponse, error)
    DeleteVolume(context.Context, *DeleteVolumeRequest) (*DeleteVolumeResponse, error)
    ControllerPublishVolume(context.Context, *ControllerPublishVolumeRequest) (*ControllerPublishVolumeResponse, error)
    ControllerUnpublishVolume(context.Context, *ControllerUnpublishVolumeRequest) (*ControllerUnpublishVolumeResponse, error)
    ValidateVolumeCapabilities(context.Context, *ValidateVolumeCapabilitiesRequest) (*ValidateVolumeCapabilitiesResponse, error)
    ListVolumes(context.Context, *ListVolumesRequest) (*ListVolumesResponse, error)
    GetCapacity(context.Context, *GetCapacityRequest) (*GetCapacityResponse, error)
    ControllerGetCapabilities(context.Context, *ControllerGetCapabilitiesRequest) (*ControllerGetCapabilitiesResponse, error)
    CreateSnapshot(context.Context, *CreateSnapshotRequest) (*CreateSnapshotResponse, error)
    DeleteSnapshot(context.Context, *DeleteSnapshotRequest) (*DeleteSnapshotResponse, error)
    ListSnapshots(context.Context, *ListSnapshotsRequest) (*ListSnapshotsResponse, error)
    ControllerExpandVolume(context.Context, *ControllerExpandVolumeRequest) (*ControllerExpandVolumeResponse, error)
    ControllerGetVolume(context.Context, *ControllerGetVolumeRequest) (*ControllerGetVolumeResponse, error)
}
```
- **NodeServer**
```go
// github.com/container-storage-interface/spec/lib/go/csi/csi.pb.go
type NodeServer interface {
    NodeStageVolume(context.Context, *NodeStageVolumeRequest) (*NodeStageVolumeResponse, error)
    NodeUnstageVolume(context.Context, *NodeUnstageVolumeRequest) (*NodeUnstageVolumeResponse, error)
    NodePublishVolume(context.Context, *NodePublishVolumeRequest) (*NodePublishVolumeResponse, error)
    NodeUnpublishVolume(context.Context, *NodeUnpublishVolumeRequest) (*NodeUnpublishVolumeResponse, error)
    NodeGetVolumeStats(context.Context, *NodeGetVolumeStatsRequest) (*NodeGetVolumeStatsResponse, error)
    NodeExpandVolume(context.Context, *NodeExpandVolumeRequest) (*NodeExpandVolumeResponse, error)
    NodeGetCapabilities(context.Context, *NodeGetCapabilitiesRequest) (*NodeGetCapabilitiesResponse, error)
    NodeGetInfo(context.Context, *NodeGetInfoRequest) (*NodeGetInfoResponse, error)
}
```
这些接口的作用可以参考前面的文章以及网上相关资料，先有个大致的印象，并且了解并不是需要实现所有的接口（不需要实现的接口指代可以直接返回一个error）。

# CSI注册

##### 注册过程

在[《kubernetes CSI（中）》](https://mp.weixin.qq.com/s/FVN5Kxckq_NL8P57AhQDlw)一文中，我们分析了CSI插件的注册流程：

【图】

CSI插件注册过程只会调用CSI进程的两个方法，这两个方法分别是`IdentityServer下的GetPluginInfo方法`和`NodeServer下的NodeGetInfo方法`，于是我们先实现这两个方法验证下验证下注册过程（其它方法暂时均直接返回error）：

- **IdentityServer下的GetPluginInfo方法**
```go
// import "github.com/container-storage-interface/spec/lib/go/csi"
func (nfs *NFSDriver) GetPluginInfo(context.Context, *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
    log.Println("GetPluginInfo request")

    return &csi.GetPluginInfoResponse{
        Name:          nfs.Name,
        VendorVersion: nfs.Version,
    }, nil
}
```

- **NodeServer下的NodeGetInfo方法**

```go
// import "github.com/container-storage-interface/spec/lib/go/csi"
func (nfs *NFSDriver) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
    log.Println("NodeGetInfo request")

    return &csi.NodeGetInfoResponse{
        NodeId: nfs.NodeID,
    }, nil
}
```

##### 注册过程产物

注册过程中会有如下产物：

- node-driver-registrar进程的sock文件：/var/lib/kubelet/plugins_registry/{csiDriverName}-reg.sock
- CSI进程的sock文件：/var/lib/kubelet/plugins/{xxx}/csi.sock
- 节点对应Node对象的annotation中会有一个关于该CSI插件的注解
- 会有一个CSINode对象

所以在开始之前，我们先看一下相关信息：
```
[root@VM-12-7-centos nfscsi]# ls /var/lib/kubelet/plugins_registry
[root@VM-12-7-centos nfscsi]# ls /var/lib/kubelet/plugins
[root@VM-12-7-centos nfscsi]#

[root@VM-12-7-centos nfscsi]# kubectl get node
NAME              STATUS   ROLES    AGE   VERSION
vm-12-11-centos   Ready    <none>   54d   v1.15.0
vm-12-7-centos    Ready    master   54d   v1.15.0
[root@VM-12-7-centos nfscsi]# kubectl get node vm-12-7-centos -oyaml| grep annotations -A 8
  annotations:
    flannel.alpha.coreos.com/backend-data: '{"VNI":1,"VtepMAC":"aa:77:1c:26:b5:88"}'
    flannel.alpha.coreos.com/backend-type: vxlan
    flannel.alpha.coreos.com/kube-subnet-manager: "true"
    flannel.alpha.coreos.com/public-ip: 10.0.12.7
    kubeadm.alpha.kubernetes.io/cri-socket: /var/run/dockershim.sock
    node.alpha.kubernetes.io/ttl: "0"
    volumes.kubernetes.io/controller-managed-attach-detach: "true"
  creationTimestamp: "2022-04-11T12:08:22Z"

[root@VM-12-7-centos nfscsi]# kubectl get csinode
No resources found.
```

##### 编译部署

编译过程我们准备如下内容：

- **Dockerfile**
```dockerfile
FROM busybox

COPY build/nfs-csi /

ENTRYPOINT ["/nfs-csi"]
```
- **build.sh**
```shell
#!/bin/bash

set -ex

# 编译
CGO_ENABLED=0 go build -mod=vendor -o build/nfs-csi cmd/main.go

image="nfs-csi:v1.0"

# 打包镜像
docker build -t $image .

# 推送镜像
# docker push image
```
用上述脚本把代码编译打包成镜像`nfs-csi:v1.0`，之后准备部署的yaml（当前阶段为了简单起见，给daemonSet的nodeSelector加了个`kubernetes.io/hostname`标签用于指定只在一个节点上运行pod）：
```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: nfs-csi-node
  namespace: kube-system

---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: nfs-csi-node
rules:
  - apiGroups: ["storage.k8s.io"]
    resources: ["csinodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]
---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: nfs-csi-node
subjects:
  - kind: ServiceAccount
    name: nfs-csi-node
    namespace: kube-system
roleRef:
  kind: ClusterRole
  name: nfs-csi-node
  apiGroup: rbac.authorization.k8s.io

---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: nfs-csi-node
  namespace: kube-system
spec:
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
  selector:
    matchLabels:
      app: nfs-csi-node
  template:
    metadata:
      labels:
        app: nfs-csi-node
    spec:
      hostNetwork: true  # original nfs connection would be broken without hostNetwork setting
      dnsPolicy: Default  # available values: Default, ClusterFirstWithHostNet, ClusterFirst
      serviceAccountName: nfs-csi-node
      nodeSelector:
        kubernetes.io/os: linux
        kubernetes.io/hostname: vm-12-7-centos # 调试阶段可以先指定某一个节点启动
      tolerations:
        - operator: "Exists"
      containers:
        - name: node-driver-registrar
          image: objectscale/csi-node-driver-registrar:v2.5.0
          imagePullPolicy: IfNotPresent
          args:
            - --v=2
            - --csi-address=/csi/csi.sock
            - --kubelet-registration-path=$(DRIVER_REG_SOCK_PATH)
          env:
            - name: DRIVER_REG_SOCK_PATH
              value: /var/lib/kubelet/plugins/csi-nfsplugin/csi.sock
            - name: KUBE_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
            - name: registration-dir
              mountPath: /registration
          resources:
            limits:
              memory: 100Mi
            requests:
              cpu: 10m
              memory: 20Mi
        - name: nfs-csi
          securityContext:
            privileged: true
            capabilities:
              add: ["SYS_ADMIN"]
            allowPrivilegeEscalation: true
          image: nfs-csi:v1.0
          imagePullPolicy: "IfNotPresent"
          args:
            - "--endpoint=$(CSI_ENDPOINT)"
            - "--nodeid=$(NODE_ID)"
          env:
            - name: NODE_ID
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: CSI_ENDPOINT
              value: /csi/csi.sock
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
          resources:
            limits:
              memory: 300Mi
            requests:
              cpu: 10m
              memory: 20Mi
      volumes:
        - name: socket-dir
          hostPath:
            path: /var/lib/kubelet/plugins/csi-nfsplugin
            type: DirectoryOrCreate
        - name: registration-dir
          hostPath:
            path: /var/lib/kubelet/plugins_registry
            type: Directory
```

##### 验证

kubectl apply上述yaml，先观察pod是否正常启动：
```
[root@VM-12-7-centos nfscsi]# kubectl apply -f deploy/node.yaml 
serviceaccount/nfs-csi-node created
clusterrole.rbac.authorization.k8s.io/nfs-csi-node created
clusterrolebinding.rbac.authorization.k8s.io/nfs-csi-node created
daemonset.apps/nfs-csi-node created
[root@VM-12-7-centos nfscsi]# 
[root@VM-12-7-centos nfscsi]# kubectl -n kube-system get ds
NAME           DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR                                                  AGE
kube-proxy     2         2         2       2            2           beta.kubernetes.io/os=linux                                    54d
nfs-csi-node   1         1         1       1            1           kubernetes.io/hostname=vm-12-7-centos,kubernetes.io/os=linux   18s
[root@VM-12-7-centos nfscsi]# kubectl -n kube-system get pod -o wide | grep csi
nfs-csi-node-9nqjp                       2/2     Running   0          69s   10.0.12.7    vm-12-7-centos    <none>           <none>
```
pod正常启动后，先查看node-driver-registrar的日志，注册正常：
```
[root@VM-12-7-centos nfscsi]# kubectl -n kube-system logs nfs-csi-node-9nqjp node-driver-registrar
I0605 07:36:18.515692       1 main.go:166] Version: v2.5.0
I0605 07:36:18.515732       1 main.go:167] Running node-driver-registrar in mode=registration
I0605 07:36:18.516127       1 main.go:191] Attempting to open a gRPC connection with: "/csi/csi.sock"
I0605 07:36:19.516949       1 main.go:198] Calling CSI driver to discover driver name
I0605 07:36:19.518158       1 main.go:208] CSI driver name: "nfscsi"
I0605 07:36:19.518187       1 node_register.go:53] Starting Registration Server at: /registration/nfscsi-reg.sock
I0605 07:36:19.518289       1 node_register.go:62] Registration Server started at: /registration/nfscsi-reg.sock
I0605 07:36:19.518426       1 node_register.go:92] Skipping HTTP server because endpoint is set to: ""
I0605 07:36:42.682092       1 main.go:102] Received GetInfo call: &InfoRequest{}
I0605 07:36:42.682280       1 main.go:109] "Kubelet registration probe created" path="/var/lib/kubelet/plugins/csi-nfsplugin/registration"
I0605 07:36:43.682173       1 main.go:102] Received GetInfo call: &InfoRequest{}
I0605 07:36:43.682208       1 main.go:109] "Kubelet registration probe created" path="/var/lib/kubelet/plugins/csi-nfsplugin/registration"
I0605 07:36:43.698187       1 main.go:120] Received NotifyRegistrationStatus call: &RegistrationStatus{PluginRegistered:true,Error:,}
```
再看看自己编码nfs-csi容器日志，和之前分析的一样，注册过程只调用了`GetPluginInfo`和`NodeGetInfo`方法：
```
[root@VM-12-7-centos nfscsi]# kubectl -n kube-system logs nfs-csi-node-9nqjp nfs-csi
2022/06/05 07:36:18 driverName: nfscsi, version: N/A, nodeID: vm-12-7-centos
2022/06/05 07:36:18 grpc server start
2022/06/05 07:36:19 GetPluginInfo request
2022/06/05 07:36:43 NodeGetInfo request
```
node-driver-registrar和CSI进程的sock文件：
```
[root@VM-12-7-centos nfscsi]# tree /var/lib/kubelet/plugins_registry
/var/lib/kubelet/plugins_registry
└── nfscsi-reg.sock
[root@VM-12-7-centos nfscsi]# tree /var/lib/kubelet/plugins_registry
/var/lib/kubelet/plugins_registry
└── nfscsi-reg.sock
```
node对象的annotation：
```
[root@VM-12-7-centos nfscsi]# kubectl get node vm-12-7-centos -oyaml| grep annotations -A 9
  annotations:
    csi.volume.kubernetes.io/nodeid: '{"nfscsi":"vm-12-7-centos"}'
    flannel.alpha.coreos.com/backend-data: '{"VNI":1,"VtepMAC":"aa:77:1c:26:b5:88"}'
    flannel.alpha.coreos.com/backend-type: vxlan
    flannel.alpha.coreos.com/kube-subnet-manager: "true"
    flannel.alpha.coreos.com/public-ip: 10.0.12.7
    kubeadm.alpha.kubernetes.io/cri-socket: /var/run/dockershim.sock
    node.alpha.kubernetes.io/ttl: "0"
    volumes.kubernetes.io/controller-managed-attach-detach: "true"
  creationTimestamp: "2022-04-11T12:08:22Z"
```
最后验证CSINode对象：
```
[root@VM-12-7-centos nfscsi]# kubectl get csinode
NAME             CREATED AT
vm-12-7-centos   2022-06-05T07:36:43Z
[root@VM-12-7-centos nfscsi]# kubectl get csinode vm-12-7-centos -oyaml
apiVersion: storage.k8s.io/v1beta1
kind: CSINode
metadata:
  creationTimestamp: "2022-06-05T07:36:43Z"
  name: vm-12-7-centos
  ownerReferences:
  - apiVersion: v1
    kind: Node
    name: vm-12-7-centos
    uid: 90889caa-4403-477f-8067-37eb341114bb
  resourceVersion: "6321023"
  selfLink: /apis/storage.k8s.io/v1beta1/csinodes/vm-12-7-centos
  uid: cf173325-31c8-41af-bac3-90466c098158
spec:
  drivers:
  - name: nfscsi
    nodeID: vm-12-7-centos
    topologyKeys: null
```

到这里我们成功完成并验证了CSI的注册。

# dynamic provisioning

在[《Dynamic Provisioning原理分析》](https://mp.weixin.qq.com/s/ZrJdDIwL3f86_OypMgGUJQ)一文中，我们分析了Dynamic Provisioning原理：`所谓的Dynamic Provisioning，其实就是创建pvc后会自动创建卷和pv，并把pv和pvc绑定`。并且我们在该文中实现了一个nfs的provisioner，当时提到实现一个provisioner只需要实现`Provisioner`接口，Provisioner接口定义如下：
```go
// sigs.k8s.io/sig-storage-lib-external-provisioner/v8/controller
type Provisioner interface {
    Provision(context.Context, ProvisionOptions) (*v1.PersistentVolume, ProvisioningState, error)
    Delete(context.Context, *v1.PersistentVolume) error
}
```
这两个方法在CSI接口中，对应的是ControllerServer下的`CreateVolume`和`DeleteVolume`方法，于是我们尝试实现这两个方法。不过在实现这两个方法前需要通过ControllerServer的`ControllerGetCapabilities`方法让调用方知道自己有CreateVolume/DeleteVolume的能力：
```go
// import "github.com/container-storage-interface/spec/lib/go/csi"
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


func NewNFSDriver(opt *Options) *NFSDriver {
    /*...*/

    nfs.addControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
        csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
    })

    return nfs
}


func (nfs *NFSDriver) ControllerGetCapabilities(context.Context, *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
    log.Println("ControllerGetCapabilities request")

    return &csi.ControllerGetCapabilitiesResponse{
        Capabilities: nfs.controllerServiceCapabilities,
    }, nil
}
```
再看看CreateVolume和DeleteVolume方法的实现：
- **CreateVolume**
```go
// import "github.com/container-storage-interface/spec/lib/go/csi"
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
```

- **DeleteVolume**
```go
// import "github.com/container-storage-interface/spec/lib/go/csi"
func (nfs *NFSDriver) DeleteVolume(_ context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
    log.Println("DeleteVolume request")

    log.Println("volumeID: ", req.GetVolumeId())

    return nil, os.Remove(filepath.Join(nfs.nfsMountPath, req.GetVolumeId()))
}
```

##### 部署验证

- **部署**

更新CSI代码后重新编译打包镜像，并且准备并apply如下yaml（注意需要CSIDriver对象告诉kubernetes CSI该插件不需要attach过程：[https://kubernetes-csi.github.io/docs/skip-attach.html](https://kubernetes-csi.github.io/docs/skip-attach.html)）：
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: nfscsi
provisioner: nfscsi

---
apiVersion: storage.k8s.io/v1
kind: CSIDriver
metadata:
  name: nfscsi
spec:
  attachRequired: false

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: nfs-csi-provisioner
  namespace: kube-system

---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: nfs-csi-provisioner
rules:
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["csinodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get"]

---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: nfs-csi-provisioner
subjects:
  - kind: ServiceAccount
    name: nfs-csi-provisioner
    namespace: kube-system
roleRef:
  kind: ClusterRole
  name: nfs-csi-provisioner
  apiGroup: rbac.authorization.k8s.io

---
kind: Deployment
apiVersion: apps/v1
metadata:
  name: nfs-csi-provisioner
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nfs-csi-provisioner
  template:
    metadata:
      labels:
        app: nfs-csi-provisioner
    spec:
      hostNetwork: true  # controller also needs to mount nfs to create dir
      dnsPolicy: Default  # available values: Default, ClusterFirstWithHostNet, ClusterFirst
      serviceAccountName: nfs-csi-provisioner
      nodeSelector:
        kubernetes.io/os: linux  # add "kubernetes.io/role: master" to run controller on master node
        # kubernetes.io/hostname: vm-12-7-centos # 调试阶段可以先指定某一个节点启动
      priorityClassName: system-cluster-critical
      tolerations:
        - key: "node-role.kubernetes.io/master"
          operator: "Exists"
          effect: "NoSchedule"
        - key: "node-role.kubernetes.io/controlplane"
          operator: "Exists"
          effect: "NoSchedule"
        - key: "node-role.kubernetes.io/control-plane"
          operator: "Exists"
          effect: "NoSchedule"
      containers:
        - name: csi-provisioner
          image: objectscale/csi-provisioner:v3.1.0
          imagePullPolicy: IfNotPresent
          args:
            - "-v=2"
            - "--csi-address=$(ADDRESS)"
            - "--leader-election"
            - "--leader-election-namespace=kube-system"
          env:
            - name: ADDRESS
              value: /csi/csi.sock
          volumeMounts:
            - mountPath: /csi
              name: socket-dir
          resources:
            limits:
              memory: 400Mi
            requests:
              cpu: 10m
              memory: 20Mi
        - name: nfs-csi
          image: nfs-csi:v1.0
          imagePullPolicy: IfNotPresent
          securityContext:
            privileged: true
            capabilities:
              add: ["SYS_ADMIN"]
            allowPrivilegeEscalation: true
          args:
            - --nodeid=$(NODE_ID)
            - --endpoint=$(CSI_ENDPOINT)
            - --server="" # 配置nfs server ip
            - --serverPath="" # 配置nfs server root path
          env:
            - name: NODE_ID
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: CSI_ENDPOINT
              value: /csi/csi.sock
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
            - name: nfs-server
              mountPath: /mount
          resources:
            limits:
              memory: 200Mi
            requests:
              cpu: 10m
              memory: 20Mi
      volumes:
        - name: socket-dir
          emptyDir: {}
        - name: nfs-server
          nfs:
            server: "" # 配置nfs server ip
            path: "" # 配置nfs server root path
```

- **验证**

观察对应的provisioner是否起来：
```
[root@VM-12-7-centos ~]# kubectl -n kube-system get pod | grep csi
nfs-csi-provisioner-6869868f45-xg9hs     2/2     Running   0          3d22h
```

准备一个如下的pvc yaml，apply该yaml：
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-pvc
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: nfscsi
  resources:
    requests:
      storage: 1Gi
```

查看是否会自动创建卷、pv，并和pvc绑定：
```
[root@VM-12-7-centos ~]# kubectl get pvc
NAME       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
test-pvc   Bound    pvc-13c67a78-93d7-4904-9e01-36017077c1df   1Gi        RWO            nfscsi         3d22h
[root@VM-12-7-centos ~]# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM              STORAGECLASS   REASON   AGE
pvc-13c67a78-93d7-4904-9e01-36017077c1df   1Gi        RWO            Delete           Bound    default/test-pvc   nfscsi                  3d22h
[root@VM-12-7-centos ~]# ls /root/nfs/
pvc-13c67a78-93d7-4904-9e01-36017077c1df
```

查看provisioner pod日志：
```
[root@VM-12-7-centos ~]# kubectl -n kube-system logs nfs-csi-provisioner-6869868f45-xg9hs -c nfs-csi
2022/06/06 15:49:17 driverName: nfscsi, version: N/A, nodeID: vm-12-7-centos
2022/06/06 15:49:17 grpc server start
2022/06/06 15:49:18 Probe request
2022/06/06 15:49:18 GetPluginInfo request
2022/06/06 15:49:18 GetPluginCapabilities request
2022/06/06 15:49:18 ControllerGetCapabilities request
2022/06/06 15:51:17 CreateVolume request
2022/06/06 15:51:17 req name:  pvc-13c67a78-93d7-4904-9e01-36017077c1df
```

删除pvc的功能也可以参照验证，这里不再赘述。于是`Dynamic Provisioning`的功能已经完成。

# Mount/Unmount

对于一些复杂的存储（例如一些块设备），从开始到使用需要经过如下步骤：

1. 挂载到宿主机上
2. 格式化
3. mount到pod对应目录

这三个步骤分别对应CSI如下方法：

1. ControllerServer下的`ControllerPublishVolume`方法（逆过程对应`ControllerUnpublishVolume`方法）
2. NodeServer下的`NodeStageVolume`方法（逆过程对应`NodeUnstageVolume`方法）
3. NodeServer下的`NodePublishVolume`方法（逆过程对应`NodeUnpublishVolume`方法）

由于本文用到的存储是nfs，可以直接将nfs对应目录挂载到pod对应目录上，因此只需要实现`NodePublishVolume`方法和`NodeUnpublishVolume`方法。和Dynamic Provisioning功能类似，也是需要实现NodeServer下的`NodeGetCapabilities`方法告知调用方相关能力：
```go
// import "github.com/container-storage-interface/spec/lib/go/csi"
func NewNFSDriver(opt *Options) *NFSDriver {
    /*…*/

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

func (nfs *NFSDriver) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
    log.Println("NodeGetCapabilities request")

    return &csi.NodeGetCapabilitiesResponse{
        Capabilities: nfs.nodeServiceCapabilities,
    }, nil
}
```
再看看NodePublishVolume和NodeUnpublishVolume：
- **NodePublishVolume**
```go
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
```
- **NodeUnpublishVolume**
```go
func (nfs *NFSDriver) NodeUnpublishVolume(_ context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
    log.Println("NodeUnpublishVolume request")

    targetPath := req.GetTargetPath()
    if err := mount.CleanupMountPoint(targetPath, nfs.mounter, true); err != nil {
        return nil, errors.Wrap(err, "clean mount point error")
    }

    return &csi.NodeUnpublishVolumeResponse{}, nil
}
```

##### 部署

由于这部分内容更新了一些启动参数，且相关功能是和node-driver-registrar部署在一个pod里的，我们需要重新编译打包镜像，并且更新最开始的注册章节的yaml（需要把之前的daemonSet删除重新创建）：
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: nfs-csi-node
  namespace: kube-system

---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: nfs-csi-node
rules:
  - apiGroups: ["storage.k8s.io"]
    resources: ["csinodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]

---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: nfs-csi-node
subjects:
  - kind: ServiceAccount
    name: nfs-csi-node
    namespace: kube-system
roleRef:
  kind: ClusterRole
  name: nfs-csi-node
  apiGroup: rbac.authorization.k8s.io

---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: nfs-csi-node
  namespace: kube-system
spec:
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
  selector:
    matchLabels:
      app: nfs-csi-node
  template:
    metadata:
      labels:
        app: nfs-csi-node
    spec:
      hostNetwork: true  # original nfs connection would be broken without hostNetwork setting
      dnsPolicy: Default  # available values: Default, ClusterFirstWithHostNet, ClusterFirst
      serviceAccountName: nfs-csi-node
      nodeSelector:
        kubernetes.io/os: linux
        kubernetes.io/hostname: vm-12-7-centos # 调试阶段可以先指定某一个节点启动
      tolerations:
        - operator: "Exists"
      containers:
        - name: node-driver-registrar
          image: objectscale/csi-node-driver-registrar:v2.5.0
          imagePullPolicy: IfNotPresent
          args:
            - --v=2
            - --csi-address=/csi/csi.sock
            - --kubelet-registration-path=$(DRIVER_REG_SOCK_PATH)
          env:
            - name: DRIVER_REG_SOCK_PATH
              value: /var/lib/kubelet/plugins/csi-nfsplugin/csi.sock
            - name: KUBE_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
            - name: registration-dir
              mountPath: /registration
          resources:
            limits:
              memory: 100Mi
            requests:
              cpu: 10m
              memory: 20Mi
        - name: nfs-csi
          securityContext:
            privileged: true
            capabilities:
              add: ["SYS_ADMIN"]
            allowPrivilegeEscalation: true
          image: nfs-csi:v1.0
          imagePullPolicy: "IfNotPresent"
          args:
            - --endpoint=$(CSI_ENDPOINT)
            - --nodeid=$(NODE_ID)
            - --server="" # 配置nfs server ip
            - --serverPath="" # 配置nfs server root path
          env:
            - name: NODE_ID
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: CSI_ENDPOINT
              value: /csi/csi.sock
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
            - name: pods-mount-dir
              mountPath: /var/lib/kubelet/pods
              mountPropagation: "Bidirectional"
          resources:
            limits:
              memory: 300Mi
            requests:
              cpu: 10m
              memory: 20Mi
      volumes:
        - name: socket-dir
          hostPath:
            path: /var/lib/kubelet/plugins/csi-nfsplugin
            type: DirectoryOrCreate
        - name: pods-mount-dir
          hostPath:
            path: /var/lib/kubelet/pods
            type: Directory
        - name: registration-dir
          hostPath:
            path: /var/lib/kubelet/plugins_registry
            type: Directory
```

##### 验证

新建个pod并引用前面的pvv挂载到容器的/pvc目录下（注意用nodeName指定pod调度到安装了CSI插件的节点）：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  nodeName: vm-12-7-centos # 运行在安装了csi插件的node上
  containers:
  - name: nginx
    image: nginx:latest
    imagePullPolicy: IfNotPresent
    volumeMounts:
    - name: nfs-pvc
      mountPath: /pvc
  volumes:
  - name: nfs-pvc
    persistentVolumeClaim:
      claimName: test-pvc

```
kubectl apply该yaml，查看pod是否正常启动并挂载：
```
[root@VM-12-7-centos ~]# kubectl get pod -o wide
NAME                    READY   STATUS    RESTARTS   AGE   IP            NODE              NOMINATED NODE   READINESS GATES
test-pod                1/1     Running   0          25m   10.244.0.10   vm-12-7-centos    <none>           <none>
[root@VM-12-7-centos ~]# kubectl exec -ti test-pod bash
root@test-pod:/# cd pvc
root@test-pod:/pvc# touch test-pod.txt
root@test-pod:/pvc# exit
exit

[root@VM-12-7-centos ~]# ls /root/nfs/pvc-13c67a78-93d7-4904-9e01-36017077c1df/
test-pod.txt
[root@VM-12-7-centos ~]#
```
再看看CSI容器的日志：
```
[root@VM-12-7-centos ~]# kubectl -n kube-system logs nfs-csi-node-tsbdl -c nfs-csi
2022/06/11 10:15:11 driverName: nfscsi, version: N/A, nodeID: vm-12-7-centos
2022/06/11 10:15:11 grpc server start
2022/06/11 10:15:12 GetPluginInfo request
2022/06/11 10:15:13 NodeGetInfo request
2022/06/11 10:17:41 NodeGetCapabilities request
2022/06/11 10:17:41 NodeGetCapabilities request
2022/06/11 10:17:41 NodePublishVolume request
2022/06/11 10:17:41 source: 10.0.12.7:/root/nfs/pvc-13c67a78-93d7-4904-9e01-36017077c1df, targetPath: /var/lib/kubelet/pods/d89f4cb9-02f2-462f-bc29-9ed0dcc0ebbd/volumes/kubernetes.io~csi/pvc-13c67a78-93d7-4904-9e01-36017077c1df/mount, options: []
```
至此，一个基于nfs的简单CSI已完成。完整示例代码放在了[https://github.com/nuczzz/nfscsi](https://github.com/nuczzz/nfscsi)，有兴趣的读者可以参考。

# 总结

本文基于nfs存储，实现了一个支持`Dynamic Provisioning`和自定义`Mount/Umount`的CSI。实现一个CSI主要需要理解CSI的原理，包括注册过程、Dynamic Provisioning、Attach/Detach、Mount/Umount等过程，同时还需要将这些过程和CSI grpc服务的方法对应清楚。

本文实现的nfs CSI比较简单，没有Attach、扩容、快照等功能，对于一些复杂的存储的CSI，需要读者参考相关资料继续探索，希望本文对读者能有一定的帮助。
