## 自定义资源对象与控制器 AppDeployer 

### 项目思路与功能
项目背景：一般在集群上部署应用都需要**deployment+service**的方式进行部署，在实际过程中会相对比较麻烦，本项目基于此背景下，创建CRD自定义对象，让调用方只要创建一个自定义对象，就能拉起整个**deployment+service**

支持功能：
1. 使用CRD资源对象创建Deployment、Service
2. 支持多容器以sidecar方式部署(环境变量、command等都支持)
3. 支持自定义是否使用Service功能
4. 支持自定义支持Service种类(NodePort or ClusterIP)
5. 支持configmap挂载给pod，并实现热更新(即：不需要重新手动删除pod。)
```bigquery
apiVersion: deploy.jiang.operator/v1
kind: AppDeployer
metadata:
  name: appdeployer-sample
spec:
  # TODO(user): Add fields here
  size: 1 # pod副本
  containers:
    - name: c1
      image: busybox:1.34
      command:
        - "sleep"
        - "3600"
      resources:
        limits:
          memory: "128Mi"
          cpu: "500m"
      volumeMounts:
        - name: appdeployer-sample # 限定与AppDeployer自己的名字相同，不然会报错。
          mountPath: /etc/config # 可自由修改
    - name: c2
      image: busybox:1.34
      command:
        - "sleep"
        - "3600"
      resources:
        limits:
          memory: "128Mi"
          cpu: "500m"
  service: true   # 自定义是否要配置Service
  service_type: NodePort # 目前支持NodePort ClusterIP 转换，注意：如果使用ClusterIP nodePort端口字段必须删除，否则会报错。
  ports:  #端口
    - port: 80
      targetPort: 80 # 容器端口
      nodePort: 30002 #service端口  注意：如果使用ClusterIP nodePort端口字段必须删除，否则会报错。
  # 支持configmap 挂载给pod，并实现自动热更新
  configmap: true

configmap_data:
  data:
    player_initial_lives: "3"
    ui_properties_file_name: "user-interface.properties"


```
思路：

自定义CRD启动时，会启动自己的Controller，并同时关联和拉起Deployment、Service。

![](https://github.com/googs1025/Kubernetes-operator-AppDeployer/blob/main/images/%E6%B5%81%E7%A8%8B%E5%9B%BE.jpg?raw=true)

### 附注
1. 本项目依赖 kubebuilder kustomize k8s集群(kubeadm安装)，请先安装这些依赖
```bash
[root@VM-0-16-centos samples]# kubebuilder version
Version: main.version{KubeBuilderVersion:"3.4.0", KubernetesVendor:"1.23.5", GitCommit:"75241ab9ff9457de77e902645792cee41ba29fed", BuildDate:"2022-04-28T17:09:31Z", GoOs:"linux", GoArch:"amd64"}
[root@VM-0-16-centos samples]# kustomize version
{Version:kustomize/v3.5.2 GitCommit:79a891f4881cfc780e77789a1d240d8f4bfa2598 BuildDate:2019-12-17T03:48:17Z GoOs:linux GoArch:amd64}
[root@VM-0-16-centos samples]# kubectl version
Client Version: version.Info{Major:"1", Minor:"22", GitVersion:"v1.22.3", GitCommit:"c92036820499fedefec0f847e2054d824aea6cd1", GitTreeState:"clean", BuildDate:"2021-10-27T18:41:28Z", GoVersion:"go1.16.9", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"22", GitVersion:"v1.22.3", GitCommit:"c92036820499fedefec0f847e2054d824aea6cd1", GitTreeState:"clean", BuildDate:"2021-10-27T18:35:25Z", GoVersion:"go1.16.9", Compiler:"gc", Platform:"linux/amd64"}
```
2. 本项目在Dockerfile Makefile中有稍作修改，如运行项目出现报错，可自行适配

### 项目测试
1. make 执行一下
```bash
[root@VM-0-16-centos k8s-operator-appDeployer]# make
/root/k8s-operator-appDeployer/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go build -o bin/manager main.go
```
2. make install 安装CRD
```bash
[root@VM-0-16-centos k8s-operator-appDeployer]# make install
GOBIN=/root/k8s-operator-appDeployer/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0
/root/k8s-operator-appDeployer/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
kustomize build config/crd | kubectl apply -f -
customresourcedefinition.apiextensions.k8s.io/appdeployers.deploy.jiang.operator configured
```
3. make run 启动控制器
```bash
[root@VM-0-16-centos k8s-operator-appDeployer]# make run
/root/k8s-operator-appDeployer/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/root/k8s-operator-appDeployer/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go run ./main.go
1.6697956888219404e+09  INFO    controller-runtime.metrics      Metrics server is starting to listen    {"addr": ":8080"}
1.669795688822234e+09   INFO    setup   starting manager
1.669795688822544e+09   INFO    Starting server {"path": "/metrics", "kind": "metrics", "addr": "[::]:8080"}
1.6697956888225658e+09  INFO    Starting server {"kind": "health probe", "addr": "[::]:8081"}
1.669795688822693e+09   INFO    controller.appdeployer  Starting EventSource    {"reconciler group": "deploy.jiang.operator", "reconciler kind": "AppDeployer", "source": "kind source: *v1.AppDeployer"}
1.6697956888227239e+09  INFO    controller.appdeployer  Starting EventSource    {"reconciler group": "deploy.jiang.operator", "reconciler kind": "AppDeployer", "source": "kind source: *v1.Deployment"}
1.6697956888227327e+09  INFO    controller.appdeployer  Starting EventSource    {"reconciler group": "deploy.jiang.operator", "reconciler kind": "AppDeployer", "source": "kind source: *v1.Service"}
1.66979568882274e+09    INFO    controller.appdeployer  Starting EventSource    {"reconciler group": "deploy.jiang.operator", "reconciler kind": "AppDeployer", "source": "kind source: *v1.Deployment"}
1.669795688822749e+09   INFO    controller.appdeployer  Starting EventSource    {"reconciler group": "deploy.jiang.operator", "reconciler kind": "AppDeployer", "source": "kind source: *v1.Service"}
1.669795688822759e+09   INFO    controller.appdeployer  Starting Controller     {"reconciler group": "deploy.jiang.operator", "reconciler kind": "AppDeployer"}
1.6697956889256444e+09  INFO    controller.appdeployer  Starting workers        {"reconciler group": "deploy.jiang.operator", "reconciler kind": "AppDeployer", "worker count": 1}
1.6697956889257421e+09  INFO    controller.appdeployer  Start Reconcile Loop    {"reconciler group": "deploy.jiang.operator", "reconciler kind": "AppDeployer", "name": "appdeployer-sample", "namespace": "default"}
1.6697956889300551e+09  INFO    controller.appdeployer  CreateOrUpdate  {"reconciler group": "deploy.jiang.operator", "reconciler kind": "AppDeployer", "name": "appdeployer-sample", "namespace": "default", "Deployment": "updated"}
1.6697956889343245e+09  INFO    controller.appdeployer  CreateOrUpdate  {"reconciler group": "deploy.jiang.operator", "reconciler kind": "AppDeployer", "name": "appdeployer-sample", "namespace": "default", "Service": "updated"}
```
4. 创建对象 
```bash
[root@VM-0-16-centos k8s-operator-appDeployer]# cd config/samples/
[root@VM-0-16-centos samples]# kubectl apply -f deploy_v1_appdeployer.yaml 
appdeployer.deploy.jiang.operator/appdeployer-sample configured
```
### 项目部署
1. 部署
```bigquery
# 打包controller docker 镜像
make docker-build
# 部署controller 项目
make deploy
```
2. 查看
```bash
[root@VM-0-16-centos samples]# kubectl get ns | grep operator-develop-system
operator-develop-system   Active   39m
[root@VM-0-16-centos samples]# kubectl get pods -noperator-develop-system
NAME                                                   READY   STATUS    RESTARTS   AGE
operator-develop-controller-manager-85bb967474-fqn6n   2/2     Running   0          40m
```

### 项目主要目录结构
```bigquery
├── Dockerfile # docker 部署
├── Makefile   # makefile使用
├── PROJECT
├── README.md
├── README1.md
├── api # 自定义资源对象
├── config
│   ├── crd
│   ├── default
│   ├── manager
│   ├── rbac
│   └── samples
│       └── deploy_v1_appdeployer.yaml
├── controllers # 控制器实现逻辑
│   ├── appdeployer_controller.go #控制器
│   ├── resource.go # 辅助函数
└── main.go
```
