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
	"errors"
	"fmt"
	"strconv"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/internal/envutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/ptr"
)

// DeploymentName gets the name of a Deployment given the app.
func DeploymentName(app *v1alpha1.App) string {
	return app.Name
}

// MakeDeployment creates a K8s Deployment from an app definition.
func MakeDeployment(
	app *v1alpha1.App,
	space *v1alpha1.Space,
) (*appsv1.Deployment, error) {
	image := app.Status.Image
	if image == "" {
		return nil, errors.New("waiting for source image in latestReadySource")
	}

	// don't modify the spec on the app
	podSpec := app.Spec.Template.Spec.DeepCopy()

	// XXX: Add a dummy environment variable that reflects the UpdateRequests.
	// This will cause knative to create a new revision of the service.
	podSpec.Containers[0].Env = append(
		podSpec.Containers[0].Env,
		corev1.EnvVar{
			Name:  fmt.Sprintf("KF_UPDATE_REQUESTS_%v", app.UID),
			Value: strconv.FormatInt(int64(app.Spec.Template.UpdateRequests), 10),
		},
	)

	// At this point in the lifecycle there should be exactly one container
	// if the webhhook is working but create one to avoid panics just in case.
	if len(podSpec.Containers) == 0 {
		podSpec.Containers = append(podSpec.Containers, corev1.Container{})
	}
	podSpec.Containers[0].Image = image
	// Execution environment variables come before others because they're built
	// to be overridden.
	podSpec.Containers[0].Env = append(space.Spec.Execution.Env, podSpec.Containers[0].Env...)
	podSpec.Containers[0].Env = envutil.DeduplicateEnvVars(podSpec.Containers[0].Env)

	// Inject VCAP env vars from secret
	podSpec.Containers[0].EnvFrom = []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: KfInjectedEnvSecretName(app),
				},
			},
		},
	}

	replicas, err := app.Spec.Instances.DeploymentReplicas()
	if err != nil {
		return nil, err
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DeploymentName(app),
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(app),
			},
			Labels: v1alpha1.UnionMaps(app.GetLabels(), app.ComponentLabels("app-scaler")),
		},
		Spec: appsv1.DeploymentSpec{
			ProgressDeadlineSeconds: ptr.Int32(300),
			Selector:                metav1.SetAsLabelSelector(labels.Set(app.ComponentLabels("app-server"))),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: app.ComponentLabels("app-server"),
					// TODO add scaling annotations
					//Annotations: app.Spec.Instances.ScalingAnnotations(),
				},
				Spec: *podSpec,
			},
			RevisionHistoryLimit: ptr.Int32(10),
			Replicas:             ptr.Int32(int32(replicas)),
		},
	}, nil
}
