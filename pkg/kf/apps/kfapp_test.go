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

package apps

import (
	"fmt"
	"os"
	"testing"

	v1alpha1 "github.com/google/kf/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/pkg/kf/describe"
	"github.com/google/kf/pkg/kf/testutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

func ExampleKfApp() {
	space := NewKfApp()
	// Setup
	space.SetName("nsname")

	// Values
	fmt.Println(space.GetName())

	// Output: nsname
}

func TestKfApp_ToApp(t *testing.T) {
	app := NewKfApp()
	app.SetName("foo")
	actual := app.ToApp()

	expected := &v1alpha1.App{
		TypeMeta: metav1.TypeMeta{
			Kind:       "App",
			APIVersion: "kf.dev/v1alpha1",
		},
		Spec: v1alpha1.AppSpec{
			Template: v1alpha1.AppSpecTemplate{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{}},
				},
			},
		},
	}
	expected.Name = "foo"

	testutil.AssertEqual(t, "generated service", expected, actual)
}

func ExampleKfApp_GetEnvVars() {
	myApp := NewKfApp()
	myApp.SetEnvVars([]corev1.EnvVar{
		{Name: "FOO", Value: "2"},
		{Name: "BAR", Value: "0"},
	})

	env := myApp.GetEnvVars()

	for _, e := range env {
		fmt.Println("Key", e.Name, "Value", e.Value)
	}

	// Output: Key FOO Value 2
	// Key BAR Value 0
}

func ExampleKfApp_GetEnvVars_emptyApp() {
	myApp := NewKfApp()

	env := myApp.GetEnvVars()

	fmt.Println(env)

	// Output: []
}

func ExampleKfApp_MergeEnvVars() {
	myApp := NewKfApp()
	myApp.SetEnvVars([]corev1.EnvVar{
		{Name: "FOO", Value: "0"},
		{Name: "BAR", Value: "0"},
	})

	myApp.MergeEnvVars([]corev1.EnvVar{
		{Name: "FOO", Value: "1"},  // will replace old
		{Name: "BAZZ", Value: "0"}, // will be added
	})

	env := myApp.GetEnvVars()

	for _, e := range env {
		fmt.Println("Key", e.Name, "Value", e.Value)
	}

	// Output: Key BAR Value 0
	// Key BAZZ Value 0
	// Key FOO Value 1
}

func ExampleKfApp_DeleteEnvVars() {
	myApp := NewKfApp()
	myApp.SetEnvVars([]corev1.EnvVar{
		{Name: "FOO", Value: "0"},
		{Name: "BAR", Value: "0"},
	})

	myApp.DeleteEnvVars([]string{"FOO", "DOES_NOT_EXIST"})

	for _, e := range myApp.GetEnvVars() {
		fmt.Println("Key", e.Name, "Value", e.Value)
	}

	// Output: Key BAR Value 0
}

func ExampleKfApp_GetNamespace() {
	myApp := NewKfApp()
	myApp.SetNamespace("my-ns")

	fmt.Println(myApp.GetNamespace())

	// Output: my-ns
}

func ExampleKfApp_GetServiceAccount() {
	myApp := NewKfApp()
	fmt.Printf("Default: %q\n", myApp.GetServiceAccount())

	myApp.SetServiceAccount("my-sa")
	fmt.Printf("After set: %q\n", myApp.GetServiceAccount())

	// Output: Default: ""
	// After set: "my-sa"
}

func ExampleKfApp_GetImage() {
	myApp := NewKfApp()
	fmt.Printf("Default: %q\n", myApp.GetImage())

	myApp.SetImage("my-company/my-app")
	fmt.Printf("After set: %q\n", myApp.GetImage())

	// Output: Default: ""
	// After set: "my-company/my-app"
}

func ExampleKfApp_GetContainerPorts() {
	myApp := NewKfApp()
	fmt.Printf("Default: %v\n", myApp.GetContainerPorts())

	myApp.SetContainerPorts([]corev1.ContainerPort{{Name: "HTTP", ContainerPort: 8080}})

	for _, port := range myApp.GetContainerPorts() {
		fmt.Printf("Open %d (%s)\n", port.ContainerPort, port.Name)

	}

	// Output: Default: []
	// Open 8080 (HTTP)
}

func ExampleKfApp_GetHealthCheck() {
	check, err := NewHealthCheck("http", "/healthz", 50)
	if err != nil {
		panic(err)
	}

	myApp := NewKfApp()
	fmt.Printf("Default: %v\n", myApp.GetHealthCheck())

	myApp.SetHealthCheck(check)

	fmt.Println("After set:")
	describe.HealthCheck(os.Stdout, myApp.GetHealthCheck())

	// Output: Default: nil
	// After set:
	// Health Check:
	//   Timeout:   50s
	//   Type:      http
	//   Endpoint:  /healthz
}

func ExampleKfApp_GetResourceRequests() {
	requests := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("100m"),
		corev1.ResourceMemory: resource.MustParse("1Gi"),
	}

	myApp := NewKfApp()
	myApp.SetResourceRequests(requests)

	out := myApp.GetResourceRequests()
	for _, key := range []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory} {
		qty := out[key]
		fmt.Println(key, "=", qty.String())
	}

	// Output: cpu = 100m
	// memory = 1Gi
}

func ExampleKfApp_GetClusterURL() {
	app := NewKfApp()
	app.Status.Address = &duckv1alpha1.Addressable{
		Addressable: duckv1beta1.Addressable{
			URL: &apis.URL{
				Host:   "app-a.some-namespace.svc.cluster.local",
				Scheme: "http",
			},
		},
	}

	fmt.Println(app.GetClusterURL())

	// Output: http://app-a.some-namespace.svc.cluster.local
}

func ExampleKfApp_GetArgs() {
	myApp := NewKfApp()
	fmt.Printf("Default: %v\n", myApp.GetArgs())

	myApp.SetArgs([]string{"arg1", "arg2"})
	fmt.Printf("After set: %v\n", myApp.GetArgs())

	// Output: Default: []
	// After set: [arg1 arg2]
}

func ExampleKfApp_GetCommand() {
	myApp := NewKfApp()
	fmt.Printf("Default: %v\n", myApp.GetCommand())

	myApp.SetCommand([]string{"/bin/bash", "-x"})
	fmt.Printf("After set: %v\n", myApp.GetCommand())

	// Output: Default: []
	// After set: [/bin/bash -x]
}
