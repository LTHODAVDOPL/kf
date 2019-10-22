// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	kfv1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/kmeta"
)

// ServiceName is the name of the service for the app
func ServiceName(app *kfv1alpha1.App) string {
	return app.Name
}

// MakeService constructs a K8s service, that is backed by the pod selector
// matching pods created by the revision.
func MakeService(app *kfv1alpha1.App) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(app),
			},
			Labels: v1alpha1.UnionMaps(app.GetLabels(), app.ComponentLabels("service")),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:     UserPortName,
				Protocol: corev1.ProtocolTCP,
				Port:     80,
				// This one is matching the public one, since this is the
				// port queue-proxy listens on.
				TargetPort: intstr.FromInt(int(getUserPort(app))),
			}},
			Selector: PodLabels(app),
		},
	}
}
