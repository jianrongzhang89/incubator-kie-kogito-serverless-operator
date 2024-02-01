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

package services

import (
	"context"
	"fmt"

	"github.com/apache/incubator-kie-kogito-serverless-operator/controllers/knative"

	"github.com/apache/incubator-kie-kogito-serverless-operator/container-builder/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const knativeBrokerAnnotation = "eventing.knative.dev/broker.class"

func KnativeResourceExists(ctx context.Context, c client.Client, dest *duckv1.Destination, nsDefault string) (*unstructured.Unstructured, error) {
	dynamicClient, err := dynamic.NewForConfig(c.GetConfig())
	if err != nil {
		return nil, err
	}
	resourceId := schema.GroupVersionResource{
		Group:    dest.Ref.Group,
		Version:  dest.Ref.APIVersion,
		Resource: dest.Ref.Kind,
	}
	namespace := dest.Ref.Namespace
	if len(namespace) == 0 {
		namespace = nsDefault
	}
	return dynamicClient.Resource(resourceId).Namespace(namespace).Get(ctx, dest.Ref.Name, metav1.GetOptions{})
}

func IsKnativeBroker(ctx context.Context, c client.Client, dest *duckv1.Destination, namespace string) (bool, error) {
	if dest == nil {
		return false, fmt.Errorf("broker destinition is empty")
	}
	knativeAvail := GetKnativeAvailability(c)
	if !knativeAvail.Eventing {
		return false, fmt.Errorf("knative eventing is not deployed in the cluster")
	}
	obj, err := KnativeResourceExists(ctx, c, dest, namespace)
	if err != nil {
		return false, fmt.Errorf("failed to find knative resource %v", dest)
	}
	annotations := obj.GetAnnotations()
	if len(annotations) > 0 {
		if _, ok := annotations[knativeBrokerAnnotation]; ok {
			return true, nil
		}
	}
	return false, nil
}

func GetKnativeAvailability(c client.Client) *knative.Availability {
	result := new(knative.Availability)
	if c.Scheme().IsGroupRegistered("serving.knative.dev") {
		result.Serving = true
	}
	if c.Scheme().IsGroupRegistered("eventing.knative.dev") {
		result.Eventing = true
	}
	return result
}
 