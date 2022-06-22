// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/portworx/px-object-controller/client/apis/pxobjectservice/v1alpha1"
	scheme "github.com/portworx/px-object-controller/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PXBucketClassesGetter has a method to return a PXBucketClassInterface.
// A group's client should implement this interface.
type PXBucketClassesGetter interface {
	PXBucketClasses() PXBucketClassInterface
}

// PXBucketClassInterface has methods to work with PXBucketClass resources.
type PXBucketClassInterface interface {
	Create(ctx context.Context, pXBucketClass *v1alpha1.PXBucketClass, opts v1.CreateOptions) (*v1alpha1.PXBucketClass, error)
	Update(ctx context.Context, pXBucketClass *v1alpha1.PXBucketClass, opts v1.UpdateOptions) (*v1alpha1.PXBucketClass, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.PXBucketClass, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.PXBucketClassList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PXBucketClass, err error)
	PXBucketClassExpansion
}

// pXBucketClasses implements PXBucketClassInterface
type pXBucketClasses struct {
	client rest.Interface
}

// newPXBucketClasses returns a PXBucketClasses
func newPXBucketClasses(c *PxobjectserviceV1alpha1Client) *pXBucketClasses {
	return &pXBucketClasses{
		client: c.RESTClient(),
	}
}

// Get takes name of the pXBucketClass, and returns the corresponding pXBucketClass object, and an error if there is any.
func (c *pXBucketClasses) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.PXBucketClass, err error) {
	result = &v1alpha1.PXBucketClass{}
	err = c.client.Get().
		Resource("pxbucketclasses").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of PXBucketClasses that match those selectors.
func (c *pXBucketClasses) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PXBucketClassList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.PXBucketClassList{}
	err = c.client.Get().
		Resource("pxbucketclasses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested pXBucketClasses.
func (c *pXBucketClasses) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("pxbucketclasses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a pXBucketClass and creates it.  Returns the server's representation of the pXBucketClass, and an error, if there is any.
func (c *pXBucketClasses) Create(ctx context.Context, pXBucketClass *v1alpha1.PXBucketClass, opts v1.CreateOptions) (result *v1alpha1.PXBucketClass, err error) {
	result = &v1alpha1.PXBucketClass{}
	err = c.client.Post().
		Resource("pxbucketclasses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(pXBucketClass).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a pXBucketClass and updates it. Returns the server's representation of the pXBucketClass, and an error, if there is any.
func (c *pXBucketClasses) Update(ctx context.Context, pXBucketClass *v1alpha1.PXBucketClass, opts v1.UpdateOptions) (result *v1alpha1.PXBucketClass, err error) {
	result = &v1alpha1.PXBucketClass{}
	err = c.client.Put().
		Resource("pxbucketclasses").
		Name(pXBucketClass.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(pXBucketClass).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the pXBucketClass and deletes it. Returns an error if one occurs.
func (c *pXBucketClasses) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("pxbucketclasses").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *pXBucketClasses) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("pxbucketclasses").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched pXBucketClass.
func (c *pXBucketClasses) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PXBucketClass, err error) {
	result = &v1alpha1.PXBucketClass{}
	err = c.client.Patch(pt).
		Resource("pxbucketclasses").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
