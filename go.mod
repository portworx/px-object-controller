module github.com/portworx/px-object-controller

go 1.16

require (
	github.com/containerd/containerd v1.6.6 // indirect
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/kubernetes-csi/csi-lib-utils v0.11.0
	github.com/libopenstorage/openstorage v9.4.17+incompatible
	github.com/opencontainers/runc v1.1.3 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.2-0.20220317124727-77977386932a // indirect
	github.com/zoido/yag-config v0.4.0
	golang.org/x/oauth2 v0.0.0-20220309155454-6242fa91716a // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	k8s.io/api v0.22.5
	k8s.io/apimachinery v0.22.5
	k8s.io/client-go v12.0.0+incompatible
)

replace (
	github.com/kubernetes-incubator/external-storage => github.com/libopenstorage/external-storage v5.1.1-0.20190919185747-9394ee8dd536+incompatible
	github.com/libopenstorage/openstorage => github.com/libopenstorage/openstorage v1.0.1-0.20220608215924-d8e6f0feea09

	// Replacing k8s.io dependencies is required if a dependency or any dependency of a dependency
	// depends on k8s.io/kubernetes. See https://github.com/kubernetes/kubernetes/issues/79384#issuecomment-505725449
	k8s.io/api => k8s.io/api v0.20.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.4
	k8s.io/apiserver => k8s.io/apiserver v0.20.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.4
	k8s.io/client-go => k8s.io/client-go v0.20.4
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.4
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.4
	k8s.io/code-generator => k8s.io/code-generator v0.20.4
	k8s.io/component-base => k8s.io/component-base v0.20.4
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.4
	k8s.io/controller-manager => k8s.io/controller-manager v0.20.4
	k8s.io/cri-api => k8s.io/cri-api v0.20.4
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.4
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.4
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.4
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.4
	k8s.io/kubectl => k8s.io/kubectl v0.20.4
	k8s.io/kubelet => k8s.io/kubelet v0.20.4
	k8s.io/kubernetes => k8s.io/kubernetes v1.20.4
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.4
	k8s.io/metrics => k8s.io/metrics v0.20.4
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.4
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.4
)
