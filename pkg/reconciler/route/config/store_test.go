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

package config

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/kf/pkg/kf/testutil"
	logtesting "knative.dev/pkg/logging/testing"

	. "knative.dev/pkg/configmap/testing"
)

func TestStoreLoadWithContext(t *testing.T) {
	store := NewDefaultConfigStore(logtesting.TestLogger(t))
	_, routingConfig := ConfigMapsFromTestFile(t, RoutingConfigName)
	store.OnConfigChanged(routingConfig)
	config := FromContext(store.ToContext(context.Background()))

	t.Run("routing config changed", func(t *testing.T) {
		expected, _ := NewRoutingConfigFromConfigMap(routingConfig)
		if diff := cmp.Diff(expected, config.Routing); diff != "" {
			t.Errorf("Unexpected routing config (-want, +got): %v", diff)
		}

		testutil.AssertEqual(t, "ingress name", "test-ingress-svc", expected.IngressServiceName)
		testutil.AssertEqual(t, "ingress ns", "test-ingress-ns", expected.IngressNamespace)
		testutil.AssertEqual(t, "knative ingress", "test-knative-ingress.knative-serving.svc.cluster.local", expected.KnativeIngressGateway)
		testutil.AssertEqual(t, "ingress name", "test-ingress-svc.test-ingress-ns.svc.cluster.local", expected.GatewayHost)
	})
}
