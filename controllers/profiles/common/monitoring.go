// Copyright 2024 Apache Software Foundation (ASF)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"context"

	operatorapi "github.com/apache/incubator-kie-kogito-serverless-operator/api/v1alpha08"
	"github.com/apache/incubator-kie-kogito-serverless-operator/log"
	"github.com/apache/incubator-kie-kogito-serverless-operator/utils/monitoring"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ MonitoringEventingHandler = &monitoringObjectManager{}

type monitoringObjectManager struct {
	serviceMonitor ObjectEnsurer
	*StateSupport
}

func NewMonitoringEventingHandler(support *StateSupport) MonitoringEventingHandler {
	return &monitoringObjectManager{
		serviceMonitor: NewObjectEnsurer(support.C, ServiceMonitorCreator),
		StateSupport:   support,
	}
}

type MonitoringEventingHandler interface {
	Ensure(ctx context.Context, workflow *operatorapi.SonataFlow) ([]client.Object, error)
}

func (k monitoringObjectManager) Ensure(ctx context.Context, workflow *operatorapi.SonataFlow) ([]client.Object, error) {
	var objs []client.Object

	monitoringAvail, err := monitoring.GetPrometheusAvailability(k.Cfg)
	if err != nil {
		klog.V(log.I).InfoS("Error checking Prometheus availability: %v", err)
		return nil, err
	}
	if !monitoringAvail {
		klog.V(log.I).InfoS("Prometheus is not installed")
	} else {
		// create serviceMonitor
		serviceMonitor, _, err := k.serviceMonitor.Ensure(ctx, workflow)
		if err != nil {
			return objs, err
		} else if serviceMonitor != nil {
			objs = append(objs, serviceMonitor)
		}
	}
	return objs, nil
}
