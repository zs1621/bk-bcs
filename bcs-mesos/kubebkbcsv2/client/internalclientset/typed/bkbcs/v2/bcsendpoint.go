/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Code generated by client-gen. DO NOT EDIT.

package v2

import (
	v2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"
	scheme "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/internalclientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// BcsEndpointsGetter has a method to return a BcsEndpointInterface.
// A group's client should implement this interface.
type BcsEndpointsGetter interface {
	BcsEndpoints(namespace string) BcsEndpointInterface
}

// BcsEndpointInterface has methods to work with BcsEndpoint resources.
type BcsEndpointInterface interface {
	Create(*v2.BcsEndpoint) (*v2.BcsEndpoint, error)
	Update(*v2.BcsEndpoint) (*v2.BcsEndpoint, error)
	UpdateStatus(*v2.BcsEndpoint) (*v2.BcsEndpoint, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v2.BcsEndpoint, error)
	List(opts v1.ListOptions) (*v2.BcsEndpointList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v2.BcsEndpoint, err error)
	BcsEndpointExpansion
}

// bcsEndpoints implements BcsEndpointInterface
type bcsEndpoints struct {
	client rest.Interface
	ns     string
}

// newBcsEndpoints returns a BcsEndpoints
func newBcsEndpoints(c *BkbcsV2Client, namespace string) *bcsEndpoints {
	return &bcsEndpoints{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the bcsEndpoint, and returns the corresponding bcsEndpoint object, and an error if there is any.
func (c *bcsEndpoints) Get(name string, options v1.GetOptions) (result *v2.BcsEndpoint, err error) {
	result = &v2.BcsEndpoint{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("bcsendpoints").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of BcsEndpoints that match those selectors.
func (c *bcsEndpoints) List(opts v1.ListOptions) (result *v2.BcsEndpointList, err error) {
	result = &v2.BcsEndpointList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("bcsendpoints").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested bcsEndpoints.
func (c *bcsEndpoints) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("bcsendpoints").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a bcsEndpoint and creates it.  Returns the server's representation of the bcsEndpoint, and an error, if there is any.
func (c *bcsEndpoints) Create(bcsEndpoint *v2.BcsEndpoint) (result *v2.BcsEndpoint, err error) {
	result = &v2.BcsEndpoint{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("bcsendpoints").
		Body(bcsEndpoint).
		Do().
		Into(result)
	return
}

// Update takes the representation of a bcsEndpoint and updates it. Returns the server's representation of the bcsEndpoint, and an error, if there is any.
func (c *bcsEndpoints) Update(bcsEndpoint *v2.BcsEndpoint) (result *v2.BcsEndpoint, err error) {
	result = &v2.BcsEndpoint{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("bcsendpoints").
		Name(bcsEndpoint.Name).
		Body(bcsEndpoint).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *bcsEndpoints) UpdateStatus(bcsEndpoint *v2.BcsEndpoint) (result *v2.BcsEndpoint, err error) {
	result = &v2.BcsEndpoint{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("bcsendpoints").
		Name(bcsEndpoint.Name).
		SubResource("status").
		Body(bcsEndpoint).
		Do().
		Into(result)
	return
}

// Delete takes name of the bcsEndpoint and deletes it. Returns an error if one occurs.
func (c *bcsEndpoints) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("bcsendpoints").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *bcsEndpoints) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("bcsendpoints").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched bcsEndpoint.
func (c *bcsEndpoints) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v2.BcsEndpoint, err error) {
	result = &v2.BcsEndpoint{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("bcsendpoints").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
