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
	"strconv"

	"github.com/google/kf/pkg/apis/kf/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/ptr"
)

const (
	UserContainerName = "user-container"

	// UserPortName is the arbitrary name given to the port the container will
	// listen on.
	UserPortName = "user-port"

	// DefaultUserPort is the default port for a container to listen on.
	DefaultUserPort = 8080
)

// PodLabels returns the labels for selecting pods of the deployment.
func PodLabels(app *v1alpha1.App) map[string]string {
	return app.ComponentLabels("app-server")
}

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
			Selector: metav1.SetAsLabelSelector(labels.Set(PodLabels(app))),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: PodLabels(app),
					Annotations: map[string]string{
						"sidecar.istio.io/inject":                          "true",
						"traffic.sidecar.istio.io/includeOutboundIPRanges": "*",
					},
				},
				Spec: makePodSpec(app, space),
			},
			RevisionHistoryLimit: ptr.Int32(10),
			Replicas:             ptr.Int32(int32(replicas)),
		},
	}, nil
}

func makePodSpec(app *v1alpha1.App, space *v1alpha1.Space) corev1.PodSpec {
	// don't modify the spec on the app
	spec := app.Spec.Template.Spec.DeepCopy()

	// At this point in the lifecycle there should be exactly one container
	// if the webhhook is working but create one to avoid panics just in case.
	if len(spec.Containers) == 0 {
		spec.Containers = append(spec.Containers, corev1.Container{})
	}

	userPort := getUserPort(app)
	userContainer := &spec.Containers[0]
	userContainer.Name = UserContainerName
	userContainer.Ports = buildContainerPorts(userPort)

	// Execution environment variables come before others because they're built
	// to be overridden.
	userContainer.Env = append(space.Spec.Execution.Env, userContainer.Env...)

	// Add in additinal CF style environment variables
	userContainer.Env = append(userContainer.Env, corev1.EnvVar{
		Name:  "PORT",
		Value: strconv.Itoa(int(userPort)),
	})

	// Inject VCAP env vars from secret
	userContainer.EnvFrom = []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: KfInjectedEnvSecretName(app),
				},
			},
		},
	}

	// Explicitly disable stdin and tty allocation
	userContainer.Stdin = false
	userContainer.TTY = false

	// If the client provides probes, we should fill in the port for them.
	rewriteUserProbe(userContainer.LivenessProbe, userPort)
	rewriteUserProbe(userContainer.ReadinessProbe, userPort)

	return *spec
}

func getUserPort(app *v1alpha1.App) int32 {
	containers := app.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		return DefaultUserPort
	}

	ports := containers[0].Ports

	if len(ports) > 0 && ports[0].ContainerPort != 0 {
		return ports[0].ContainerPort
	}

	// TODO: Consider using container EXPOSE metadata from image before
	// falling back to default value.

	return DefaultUserPort
}

func buildContainerPorts(userPort int32) []corev1.ContainerPort {
	return []corev1.ContainerPort{{
		Name:          UserPortName,
		ContainerPort: userPort,
	}}
}

func rewriteUserProbe(p *corev1.Probe, userPort int32) {
	switch {
	case p == nil:
		return
	case p.HTTPGet != nil:
		p.HTTPGet.Port = intstr.FromInt(int(userPort))
	case p.TCPSocket != nil:
		p.TCPSocket.Port = intstr.FromInt(int(userPort))
	}
}
