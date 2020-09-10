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

package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/master"
	discoverys "github.com/Tencent/bk-bcs/bcs-common/pkg/module-discovery"
	modulediscovery "github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/register"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/register/kong"

	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
)

//New create
func New() *DiscoveryServer {
	cxt, cfunc := context.WithCancel(context.Background())
	s := &DiscoveryServer{
		exitCancel: cfunc,
		exitCxt:    cxt,
		evtCh:      make(chan *ModuleEvent, 12),
	}
	return s
}

//ModuleEvent event
type ModuleEvent struct {
	// Module name
	Module string
	// GoMicro flag for go-micro registry
	GoMicro bool
	// flag for delete
	Deletion bool
	// Svc api-gateway service definition
	Svc *register.Service
}

//DiscoveryServer holds all resources for services discovery
type DiscoveryServer struct {
	option *ServerOptions
	//manager for gateway information register
	regMgr register.Register
	//adapter for service structure convertion
	adapter *Adapter
	//bk-bcs modules discovery for backend service list
	discovery discoverys.ModuleDiscovery
	//go micro version discovery
	microDiscovery modulediscovery.Discovery
	//self node registe & master node discovery
	bcsRegister master.Master
	//exit func
	exitCancel context.CancelFunc
	//exit context
	exitCxt context.Context
	//Event channel for module-discovery callback
	evtCh chan *ModuleEvent
}

//Init init all running resources, including
// 1. configuration validation
// 2. connecting gateway admin api
// 3. init backend service information adapter
func (s *DiscoveryServer) Init(option *ServerOptions) error {
	if option == nil {
		return fmt.Errorf("Lost ServerOptions")
	}
	s.option = option
	if err := option.Valid(); err != nil {
		return err
	}
	//init gateway master discovery
	if err := s.selfRegister(); err != nil {
		return err
	}
	//init gateway manager
	gatewayAddrs := strings.Split(option.AdminAPI, ",")
	tlsConfig, err := option.GetClientTLS()
	if err != nil {
		return err
	}
	s.regMgr, err = kong.New(gatewayAddrs, tlsConfig)
	if err != nil {
		blog.Errorf("gateway init kong admin api register implementation failed, %s", err.Error())
		return err
	}
	//init service data adapter
	s.adapter = NewAdapter(option, option.Modules)
	//init module disovery
	allModules := append(defaultModules, option.Modules...)
	s.discovery, err = discoverys.NewDiscoveryV2(option.ZkConfig.BCSZk, allModules)
	if err != nil {
		blog.Errorf("gateway init services discovery failed, %s", err.Error())
		return err
	}
	s.discovery.RegisterEventFunc(s.moduleEventNotifycation)

	//init etcd registry feature with modulediscovery base on micro.Registry
	if option.Etcd.Feature {
		blog.Infof("gateway-discovery check etcd registry feature turn on, try to initialize etcd registry")
		etcdTLSConfig, err := option.GetEtcdRegistryTLS()
		if err != nil {
			blog.Errorf("gateway init etcd registry feature failed, no tlsConfig parsed correctlly, %s", err.Error())
			return err
		}
		//initialize micro registry
		addrs := strings.Split(option.Address, ",")
		mregistry := etcd.NewRegistry(
			registry.Addrs(addrs...),
			registry.TLSConfig(etcdTLSConfig),
		)
		if err := mregistry.Init(); err != nil {
			blog.Errorf("gateway init etcd registry feature failed, %s", err.Error())
			return err
		}
		modules := strings.Split(option.Etcd.GrpcModules, ",")
		s.microDiscovery = modulediscovery.NewDiscovery(modules, s.microModuleEvent, mregistry)
		blog.Infof("gateway init etcd registry success, try to init bkbcs module watch")
	}
	return nil
}

//Run running all necessary convertion logic, block
func (s *DiscoveryServer) Run() error {
	//s.discovery.
	//check master status first
	if err := s.dataSynchronization(); err != nil {
		blog.Errorf("gateway-discovery first data synchronization failed, %s", err.Error())
		return err
	}
	tick := time.NewTicker(time.Second * 60)
	for {
		select {
		case <-s.exitCxt.Done():
			blog.Infof("gateway-discovery asked to exit")
			return nil
		case <-tick.C:
			blog.Infof("gateway-discovery time to verify data synchronization....")
			s.dataSynchronization()
		case evt := <-s.evtCh:
			if evt == nil {
				blog.Errorf("module-discovery event channel closed, gateway-discovery error exit")
				return fmt.Errorf("module-discover channel closed")
			}
			blog.Infof("gateway-discovery got module %s changed event", evt.Module)
			//ready to update specified module proxy rules
			if evt.GoMicro {
				s.handleMicroChange(evt)
			} else {
				s.handleModuleChange(evt)
			}
		}
	}
}

//Stop all backgroup routines
func (s *DiscoveryServer) Stop() {
	s.bcsRegister.Clean()
	s.bcsRegister.Finit()
	s.microDiscovery.Stop()
	s.exitCancel()
}

//selfRegister
func (s *DiscoveryServer) selfRegister() error {
	zkAddrs := strings.Split(s.option.BCSZk, ",")
	selfPath := filepath.Join(types.BCS_SERV_BASEPATH, types.BCS_MODULE_GATEWAYDISCOVERY)
	//self node information
	hostname, _ := os.Hostname()
	self := &types.ServerInfo{
		IP:         s.option.ServiceConfig.Address,
		Port:       s.option.ServiceConfig.Port,
		Pid:        os.Getpid(),
		HostName:   hostname,
		Scheme:     "https",
		Version:    version.BcsVersion,
		MetricPort: s.option.MetricConfig.MetricPort,
	}
	var err error
	s.bcsRegister, err = master.NewZookeeperMaster(zkAddrs, selfPath, self)
	if err != nil {
		blog.Errorf("gateway-discovery init zookeeper master machinery failed, %s", err.Error())
		return err
	}
	//ready to start
	if err = s.bcsRegister.Init(); err != nil {
		blog.Errorf("gateway-discovery start master machinery failed, %s", err.Error())
		return err
	}
	if err = s.bcsRegister.Register(); err != nil {
		blog.Errorf("gateway-discvovery register local service instance failed, %s", err.Error())
		return err
	}
	//time for registe & master ready
	time.Sleep(time.Second)
	return nil
}

//dataSynchronization sync all data from bk bcs service discovery to gateway
func (s *DiscoveryServer) dataSynchronization() error {
	if !s.bcsRegister.IsMaster() {
		blog.Infof("gateway-discovery instance is not master, skip data synchronization")
		return nil
	}
	blog.V(3).Infof("gateway-discovery instance is master, ready to sync all datas")
	//first get all gateway route information
	regisetedService, err := s.regMgr.ListServices()
	if err != nil {
		blog.Errorf("gateway-discovery get all registed Service from Register failed, %s. wait for next tick", err.Error())
		return err
	}
	regisetedMap := make(map[string]*register.Service)
	if len(regisetedService) == 0 {
		blog.Warnf("gateway-discovery finds no registed service from Register, maybe this is first synchronization.")
	} else {
		for _, srv := range regisetedService {
			blog.V(3).Infof("gateway-discovery check Service %s is under regiseted", srv.Name)
			regisetedMap[srv.Name] = srv
		}
	}

	var allCaches []*register.Service
	//* module step 1: get all register module information from zookeeper discovery
	allModules := append(defaultModules, s.option.Modules...)
	localCaches, err := s.formatMultiServerInfo(allModules)
	if err != nil {
		blog.Errorf("disovery formate zookeeper Service info when in Synchronization, %s", err.Error())
		return err
	}
	//check zookeeper module info
	if len(localCaches) == 0 {
		blog.Warnf("gateway-discovery finds no bk-bcs service in module-discovery, please check bk-bcs discovery machinery")
	} else {
		allCaches = append(allCaches, localCaches...)
	}
	//* module step 2: check etcd registry feature, if feature is on,
	// get all module information from etcd disocvery
	if s.option.Etcd.Feature {
		etcdModules, err := s.formatMultiEtcdService()
		if err != nil {
			blog.Errorf("discovery format etcd service info when in Synchronization, %s", err.Error())
			return err
		}
		if len(etcdModules) == 0 {
			blog.Warnf("gateway-discovery finds no bk-bcs service in Micro-discovery, please check bk-bcs discovery machinery")
		} else {
			allCaches = append(allCaches, etcdModules...)
		}
	}
	//udpate datas in gateway
	for _, local := range allCaches {
		svc, ok := regisetedMap[local.Name]
		if ok {
			//service reigsted, we affirm that proxy rule is correct
			// so just update backend targets info, if rules of plugins & routes
			// change frequently, we need to verify all changes between oldSvc & newSvc.
			// but now, we confirm that rules are stable. operations can be done quickly by manually
			if err := s.regMgr.ReplaceTargetByService(svc, local.Backends); err != nil {
				blog.Errorf("gateway-discovery update Service %s backend failed in synchronization, %s. backend %+v", svc.Name, err.Error(), local.Backends)
				continue
			}
			blog.V(5).Infof("Update serivce %s backend %+v successfully", svc.Name, local.Backends)
		} else {
			blog.Infof("Service %s is Not affective in api-gateway when synchronization, try creation", local.Name)
			//create service in api-gateway
			if err := s.regMgr.CreateService(local); err != nil {
				blog.Errorf("discovery create Service %s failed in synchronization, %s. details: %+v", local.Name, err.Error(), local)
				continue
			}
			blog.Infof("discovery create %s Service successfully", local.Name)
			blog.V(3).Infof("Service Creation details: %+v", local)
		}
	}
	blog.Infof("gateway-discovery data synchroniztion finish")
	//todo(DevelperJim): try to fix this feature if we don't allow edit api-gateway configuration manually
	//we don't clean additional datas in api-gateway,
	// because we allow registe service information in api-gateway manually
	return nil
}

func (s *DiscoveryServer) gatewayServiceSync(event *ModuleEvent) error {
	//update service route
	exist, err := s.regMgr.GetService(event.Svc.Name)
	if err != nil {
		blog.Errorf("gateway-discovery get register Service %s failed, %s", event.Module, err.Error())
		return err
	}
	if exist == nil {
		blog.Infof("gateway-discovery find no %s module in api-gateway, try to create...", event.Module)
		if err := s.regMgr.CreateService(event.Svc); err != nil {
			blog.Errorf("gateway-discovery create Service %s to api-gateway failed, %s", event.Module, err.Error())
			return err
		}
		blog.Infof("gateway-discovery create Service %s successfully, serviceName: %s", event.Module, event.Svc.Name)
	} else {
		//only update Target for Service
		if err := s.regMgr.ReplaceTargetByService(event.Svc, event.Svc.Backends); err != nil {
			blog.Errorf("gateway-discovery update Service %s Target failed, %s", event.Svc.Name, err.Error())
			return err
		}
		blog.Infof("gateway-discovery update Target for Service %s in api-gateway successfully, serviceName: %s", event.Module, event.Svc.Name)
	}
	return nil
}

//detailServiceVerification all information including service/plugin/target check
func (s *DiscoveryServer) detailServiceVerification(newSvc *register.Service, oldSvc *register.Service) {
	//todo(DeveloperJim): we need complete verification if plugin & route rules changed frequently, not now
}
