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
// +groupName=pxobjectservice.portworx.io
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

	// error is the last observed error during bucket creation, if any.
	Error *error `json:"error,omitempty" protobuf:"bytes,2,opt,name=error"`
}
