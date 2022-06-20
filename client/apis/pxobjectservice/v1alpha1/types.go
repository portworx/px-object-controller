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
