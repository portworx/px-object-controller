// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/portworx/px-object-controller/client/apis/objectservice/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePXBucketClaims implements PXBucketClaimInterface
type FakePXBucketClaims struct {
	Fake *FakeObjectV1alpha1
	ns   string
}

var pxbucketclaimsResource = schema.GroupVersionResource{Group: "object.portworx.io", Version: "v1alpha1", Resource: "pxbucketclaims"}

var pxbucketclaimsKind = schema.GroupVersionKind{Group: "object.portworx.io", Version: "v1alpha1", Kind: "PXBucketClaim"}

// Get takes name of the pXBucketClaim, and returns the corresponding pXBucketClaim object, and an error if there is any.
func (c *FakePXBucketClaims) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.PXBucketClaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(pxbucketclaimsResource, c.ns, name), &v1alpha1.PXBucketClaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PXBucketClaim), err
}

// List takes label and field selectors, and returns the list of PXBucketClaims that match those selectors.
func (c *FakePXBucketClaims) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PXBucketClaimList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(pxbucketclaimsResource, pxbucketclaimsKind, c.ns, opts), &v1alpha1.PXBucketClaimList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PXBucketClaimList{ListMeta: obj.(*v1alpha1.PXBucketClaimList).ListMeta}
	for _, item := range obj.(*v1alpha1.PXBucketClaimList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested pXBucketClaims.
func (c *FakePXBucketClaims) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(pxbucketclaimsResource, c.ns, opts))

}

// Create takes the representation of a pXBucketClaim and creates it.  Returns the server's representation of the pXBucketClaim, and an error, if there is any.
func (c *FakePXBucketClaims) Create(ctx context.Context, pXBucketClaim *v1alpha1.PXBucketClaim, opts v1.CreateOptions) (result *v1alpha1.PXBucketClaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(pxbucketclaimsResource, c.ns, pXBucketClaim), &v1alpha1.PXBucketClaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PXBucketClaim), err
}

// Update takes the representation of a pXBucketClaim and updates it. Returns the server's representation of the pXBucketClaim, and an error, if there is any.
func (c *FakePXBucketClaims) Update(ctx context.Context, pXBucketClaim *v1alpha1.PXBucketClaim, opts v1.UpdateOptions) (result *v1alpha1.PXBucketClaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(pxbucketclaimsResource, c.ns, pXBucketClaim), &v1alpha1.PXBucketClaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PXBucketClaim), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePXBucketClaims) UpdateStatus(ctx context.Context, pXBucketClaim *v1alpha1.PXBucketClaim, opts v1.UpdateOptions) (*v1alpha1.PXBucketClaim, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(pxbucketclaimsResource, "status", c.ns, pXBucketClaim), &v1alpha1.PXBucketClaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PXBucketClaim), err
}

// Delete takes name of the pXBucketClaim and deletes it. Returns an error if one occurs.
func (c *FakePXBucketClaims) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(pxbucketclaimsResource, c.ns, name, opts), &v1alpha1.PXBucketClaim{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePXBucketClaims) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(pxbucketclaimsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.PXBucketClaimList{})
	return err
}

// Patch applies the patch and returns the patched pXBucketClaim.
func (c *FakePXBucketClaims) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PXBucketClaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(pxbucketclaimsResource, c.ns, name, pt, data, subresources...), &v1alpha1.PXBucketClaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PXBucketClaim), err
}
