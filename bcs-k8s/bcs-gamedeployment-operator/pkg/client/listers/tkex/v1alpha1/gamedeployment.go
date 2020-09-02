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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// GameDeploymentLister helps list GameDeployments.
type GameDeploymentLister interface {
	// List lists all GameDeployments in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.GameDeployment, err error)
	// GameDeployments returns an object that can list and get GameDeployments.
	GameDeployments(namespace string) GameDeploymentNamespaceLister
	GameDeploymentListerExpansion
}

// gameDeploymentLister implements the GameDeploymentLister interface.
type gameDeploymentLister struct {
	indexer cache.Indexer
}

// NewGameDeploymentLister returns a new GameDeploymentLister.
func NewGameDeploymentLister(indexer cache.Indexer) GameDeploymentLister {
	return &gameDeploymentLister{indexer: indexer}
}

// List lists all GameDeployments in the indexer.
func (s *gameDeploymentLister) List(selector labels.Selector) (ret []*v1alpha1.GameDeployment, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.GameDeployment))
	})
	return ret, err
}

// GameDeployments returns an object that can list and get GameDeployments.
func (s *gameDeploymentLister) GameDeployments(namespace string) GameDeploymentNamespaceLister {
	return gameDeploymentNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// GameDeploymentNamespaceLister helps list and get GameDeployments.
type GameDeploymentNamespaceLister interface {
	// List lists all GameDeployments in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.GameDeployment, err error)
	// Get retrieves the GameDeployment from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.GameDeployment, error)
	GameDeploymentNamespaceListerExpansion
}

// gameDeploymentNamespaceLister implements the GameDeploymentNamespaceLister
// interface.
type gameDeploymentNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all GameDeployments in the indexer for a given namespace.
func (s gameDeploymentNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.GameDeployment, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.GameDeployment))
	})
	return ret, err
}

// Get retrieves the GameDeployment from the indexer for a given namespace and name.
func (s gameDeploymentNamespaceLister) Get(name string) (*v1alpha1.GameDeployment, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("gamedeployment"), name)
	}
	return obj.(*v1alpha1.GameDeployment), nil
}
