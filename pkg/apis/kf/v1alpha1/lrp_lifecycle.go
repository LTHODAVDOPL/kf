// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
  "fmt"
  autoscalingv1 "k8s.io/api/autoscaling/v1"
  "k8s.io/apimachinery/pkg/runtime/schema"
  "knative.dev/pkg/apis"
  duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
  duckv1 "knative.dev/pkg/apis/duck/v1"
  corev1 "k8s.io/api/core/v1"
  servingv1alpha1 "github.com/google/kf/third_party/knative-serving/pkg/apis/serving/v1alpha1"
  appsv1 "k8s.io/api/apps/v1"
)

// GetGroupVersionKind returns the GroupVersionKind.
func (r *LongRunningProcess) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("LongRunningProcess")
}

// ConditionType represents a Service condition value
const (
	// LRPConditionReady is set when the LRP is ready to serve
	// traffic.
	LRPConditionReady = apis.ConditionReady

	// LRPConditionServiceReady denotes that the K8s service for the LRP has been
	// created.
	LRPConditionServiceReady apis.ConditionType = "ServiceReady"

	// LRPConditionAutoscalerReady denotes that the K8s Horizontal Pod Autoscaler has
	// been created for the LRP.
	LRPConditionAutoscalerReady apis.ConditionType = "AutoscalerReady"

	// LRPConditionDeploymentReady denotes that the deployment backing the app is
	// ready to serve traffic.
	LRPConditionDeploymentReady apis.ConditionType = "DeploymentReady"
)

func (status *LongRunningProcessStatus) manage() apis.ConditionManager {
	return apis.NewLivingConditionSet(
		LRPConditionReady,
		LRPConditionServiceReady,
		LRPConditionAutoscalerReady,
		LRPConditionDeploymentReady,
	).Manage(status)
}

// IsReady looks at the conditions to see if they are happy.
func (status *LongRunningProcessStatus) IsReady() bool {
	return status.manage().IsHappy()
}

// GetCondition returns the condition by name.
func (status *LongRunningProcessStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return status.manage().GetCondition(t)
}

// InitializeConditions sets the initial values to the conditions.
func (status *LongRunningProcessStatus) InitializeConditions() {
	status.manage().InitializeConditions()
}

// ServiceCondition gets a manager for the service condition.
func (status *LongRunningProcessStatus) ServiceCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), LRPConditionServiceReady, "Service")
}

// AutoscalerCondition gets a manager for the autoscaler condition.
func (status *LongRunningProcessStatus) AutoscalerCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), LRPConditionAutoscalerReady, "Autoscaler")
}

// DeploymentCondition gets a manager for the deployment condition.
func (status *LongRunningProcessStatus) DeploymentCondition() SingleConditionManager {
	return NewSingleConditionManager(status.manage(), LRPConditionDeploymentReady, "Deployment")
}

func (status *LongRunningProcessStatus) duck() *duckv1beta1.Status {
	return &status.Status
}

func (status *LongRunningProcessStatus) PropagateServiceStatus(service *corev1.Service) {
  // Services don't have conditions to indicate readiness.

  url := &apis.URL{
    Scheme: "http",
    Host:     fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, service.Namespace),
  }

  status.RouteStatusFields = servingv1alpha1.RouteStatusFields{
    URL: url,
    Address: duckv1.Addressable{
      URL: url,
    },
  }

  status.ServiceCondition().MarkSuccess()
}

func (status *LongRunningProcessStatus) PropagateAutoscalerStatus(autoscaler *autoscalingv1.HorizontalPodAutoscaler) {
  if autoscaler.ObservedGeneration != autoscaler.Generation {
    status.AutoscalerCondition().MarkReconciliationPending()
    return
  }

  status.CurrentReplicas = autoscaler.Status.CurrentReplicas
  status.DesiredReplicas = autoscaler.Status.DesiredReplicas

  status.AutoscalerCondition().MarkSuccess()
}

func (status *LongRunningProcessStatus) PropagateDeploymentStatus(deployment appsv1.Deployment) {
  if deployment.Status.ObservedGeneration != deployment.Generation {
    status.DeploymentCondition().MarkReconciliationPending()
    return
  }

  // TODO propagate deployment status
  status.ConfigurationStatusFields == servingv1alpha1.ConfigurationStatusFields{
    LatestReadyRevisionName: "",
    LatestCreatedRevisionName: "",
  }

  ds := servingv1alpha1.TransformDeploymentStatus(deployment.Status)

  cond := ds.GetCondition(serving.DeploymentConditionReady)
	if cond == nil {
		return
	}

	switch cond.Status {
	case corev1.ConditionUnknown:
		status.manager().MarkUnknown(LRPConditionDeploymentReady, cond.Reason, cond.Message)
	case corev1.ConditionTrue:
	  status.manager().MarkTrue(LRPConditionDeploymentReady)
	case corev1.ConditionFalse:
		status.manager().MarkFalse(LRPConditionDeploymentReady, cond.Reason, cond.Message)
	}
}
