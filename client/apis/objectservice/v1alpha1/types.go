// +kubebuilder:object:generate=true
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PXBucketClaim is a user's request for a bucket
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=pbc
// +groupName=objectservice.portworx.io
type PXBucketClaim struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// spec defines the desired characteristics of a bucket requested by a user.
	// Required.
	Spec BucketClaimSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`

	// status represents the current information of a bucket.
	// +optional
	Status *BucketClaimSpec `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// PXBucketClaimList is a list of PXBucketClaim objects
type PXBucketClaimList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of PXBucketClaims
	Items []PXBucketClaim `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// BucketClaimSpec describes the common attributes of a volume snapshot.
type BucketClaimSpec struct {
	// BucketClassName is the name of the PXBucketClass
	// requested by the PXBucketClaim.
	// Required.
	BucketClassName *string `json:"bucketClassName,omitempty" protobuf:"bytes,1,opt,name=bucketClassName"`
}

// BucketStatus is the status of the PXBucketClaim
type BucketClaimStatus struct {
	// provisioned indicates if the bucket is created.
	// +optional
	Provisioned *bool `json:"provisioned,omitempty" protobuf:"varint,1,opt,name=provisioned"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PXBucketClass is a user's template for a bucket
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=pbclass
// +groupName=objectservice.portworx.io
type PXBucketClass struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Driver defines the driver to use
	// Required.
	Driver string `json:"driver" protobuf:"bytes,2,opt,name=driver"`

	// Region defines the region to use
	// +optional
	Region string `json:"region" protobuf:"bytes,3,opt,name=region"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// PXBucketClassList is a list of PXBucketClass objects
type PXBucketClassList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of PXBucketClaims
	Items []PXBucketClass `json:"items" protobuf:"bytes,2,rep,name=items"`
}
