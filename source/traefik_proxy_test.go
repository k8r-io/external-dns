/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package source

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	traefikV1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	fakeKube "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/external-dns/endpoint"
)

// This is a compile-time validation that traefikSource is a Source.
var _ Source = &traefikSource{}

const defaultTraefikNamespace = "traefik"

func TestTraefikProxyIngressRouteEndpoints(t *testing.T) {
	t.Parallel()

	for _, ti := range []struct {
		title        string
		ingressRoute traefikV1alpha1.IngressRoute
		expected     []*endpoint.Endpoint
	}{
		{
			title: "IngressRoute with hostname annotation",
			ingressRoute: traefikV1alpha1.IngressRoute{
				TypeMeta: metav1.TypeMeta{
					APIVersion: traefikV1alpha1.SchemeGroupVersion.String(),
					Kind:       "IngressRoute",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroute-annotation",
					Namespace: defaultTraefikNamespace,
					Annotations: map[string]string{
						"external-dns.alpha.kubernetes.io/hostname": "a.example.com",
						"external-dns.alpha.kubernetes.io/target":   "target.domain.tld",
						"kubernetes.io/ingress.class":               "traefik",
					},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName:    "a.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressroute/traefik/ingressroute-annotation",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
			},
		},
		{
			title: "IngressRoute with host rule",
			ingressRoute: traefikV1alpha1.IngressRoute{
				TypeMeta: metav1.TypeMeta{
					APIVersion: traefikV1alpha1.SchemeGroupVersion.String(),
					Kind:       "IngressRoute",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroute-host-match",
					Namespace: defaultTraefikNamespace,
					Annotations: map[string]string{
						"external-dns.alpha.kubernetes.io/target": "target.domain.tld",
						"kubernetes.io/ingress.class":             "traefik",
					},
				},
				Spec: traefikV1alpha1.IngressRouteSpec{
					Routes: []traefikV1alpha1.Route{
						{
							Match: "Host(`b.example.com`)",
						},
					},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName:    "b.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressroute/traefik/ingressroute-host-match",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
			},
		},
		{
			title: "IngressRoute with hostheader rule",
			ingressRoute: traefikV1alpha1.IngressRoute{
				TypeMeta: metav1.TypeMeta{
					APIVersion: traefikV1alpha1.SchemeGroupVersion.String(),
					Kind:       "IngressRoute",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroute-hostheader-match",
					Namespace: defaultTraefikNamespace,
					Annotations: map[string]string{
						"external-dns.alpha.kubernetes.io/target": "target.domain.tld",
						"kubernetes.io/ingress.class":             "traefik",
					},
				},
				Spec: traefikV1alpha1.IngressRouteSpec{
					Routes: []traefikV1alpha1.Route{
						{
							Match: "HostHeader(`c.example.com`)",
						},
					},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName:    "c.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressroute/traefik/ingressroute-hostheader-match",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
			},
		},
		{
			title: "IngressRoute with multiple host rules",
			ingressRoute: traefikV1alpha1.IngressRoute{
				TypeMeta: metav1.TypeMeta{
					APIVersion: traefikV1alpha1.SchemeGroupVersion.String(),
					Kind:       "IngressRoute",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroute-multi-host-match",
					Namespace: defaultTraefikNamespace,
					Annotations: map[string]string{
						"external-dns.alpha.kubernetes.io/target": "target.domain.tld",
						"kubernetes.io/ingress.class":             "traefik",
					},
				},
				Spec: traefikV1alpha1.IngressRouteSpec{
					Routes: []traefikV1alpha1.Route{
						{
							Match: "Host(`d.example.com`) || Host(`e.example.com`)",
						},
					},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName:    "d.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressroute/traefik/ingressroute-multi-host-match",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
				{
					DNSName:    "e.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressroute/traefik/ingressroute-multi-host-match",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
			},
		},
		{
			title: "IngressRoute with multiple host rules and annotation",
			ingressRoute: traefikV1alpha1.IngressRoute{
				TypeMeta: metav1.TypeMeta{
					APIVersion: traefikV1alpha1.SchemeGroupVersion.String(),
					Kind:       "IngressRoute",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroute-multi-host-annotations-match",
					Namespace: defaultTraefikNamespace,
					Annotations: map[string]string{
						"external-dns.alpha.kubernetes.io/hostname": "f.example.com",
						"external-dns.alpha.kubernetes.io/target":   "target.domain.tld",
						"kubernetes.io/ingress.class":               "traefik",
					},
				},
				Spec: traefikV1alpha1.IngressRouteSpec{
					Routes: []traefikV1alpha1.Route{
						{
							Match: "Host(`g.example.com`, `h.example.com`)",
						},
					},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName:    "f.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressroute/traefik/ingressroute-multi-host-annotations-match",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
				{
					DNSName:    "g.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressroute/traefik/ingressroute-multi-host-annotations-match",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
				{
					DNSName:    "h.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressroute/traefik/ingressroute-multi-host-annotations-match",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
			},
		},
		{
			title: "IngressRoute omit wildcard",
			ingressRoute: traefikV1alpha1.IngressRoute{
				TypeMeta: metav1.TypeMeta{
					APIVersion: traefikV1alpha1.SchemeGroupVersion.String(),
					Kind:       "IngressRoute",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroute-omit-wildcard-host",
					Namespace: defaultTraefikNamespace,
					Annotations: map[string]string{
						"external-dns.alpha.kubernetes.io/target": "target.domain.tld",
						"kubernetes.io/ingress.class":             "traefik",
					},
				},
				Spec: traefikV1alpha1.IngressRouteSpec{
					Routes: []traefikV1alpha1.Route{
						{
							Match: "Host(`*`)",
						},
					},
				},
			},
			expected: nil,
		},
	} {
		ti := ti
		t.Run(ti.title, func(t *testing.T) {
			t.Parallel()

			fakeKubernetesClient := fakeKube.NewSimpleClientset()
			scheme := runtime.NewScheme()
			scheme.AddKnownTypes(ingressrouteGVR.GroupVersion(), &traefikV1alpha1.IngressRoute{}, &traefikV1alpha1.IngressRouteList{})
			scheme.AddKnownTypes(ingressrouteTCPGVR.GroupVersion(), &traefikV1alpha1.IngressRouteTCP{}, &traefikV1alpha1.IngressRouteTCPList{})
			scheme.AddKnownTypes(ingressrouteUDPGVR.GroupVersion(), &traefikV1alpha1.IngressRouteUDP{}, &traefikV1alpha1.IngressRouteUDPList{})
			fakeDynamicClient := fakeDynamic.NewSimpleDynamicClient(scheme)

			ir := unstructured.Unstructured{}

			ingressRouteAsJSON, err := json.Marshal(ti.ingressRoute)
			assert.NoError(t, err)

			assert.NoError(t, ir.UnmarshalJSON(ingressRouteAsJSON))

			// Create proxy resources
			_, err = fakeDynamicClient.Resource(ingressrouteGVR).Namespace(defaultTraefikNamespace).Create(context.Background(), &ir, metav1.CreateOptions{})
			assert.NoError(t, err)

			source, err := NewTraefikSource(context.TODO(), fakeDynamicClient, fakeKubernetesClient, defaultTraefikNamespace, "kubernetes.io/ingress.class=traefik")
			assert.NoError(t, err)
			assert.NotNil(t, source)

			count := &unstructured.UnstructuredList{}
			for len(count.Items) < 1 {
				count, _ = fakeDynamicClient.Resource(ingressrouteGVR).Namespace(defaultTraefikNamespace).List(context.Background(), metav1.ListOptions{})
			}

			endpoints, err := source.Endpoints(context.Background())
			assert.NoError(t, err)
			assert.Len(t, endpoints, len(ti.expected))
			assert.Equal(t, endpoints, ti.expected)
		})
	}
}

func TestTraefikProxyIngressRouteTCPEndpoints(t *testing.T) {
	t.Parallel()

	for _, ti := range []struct {
		title           string
		ingressRouteTCP traefikV1alpha1.IngressRouteTCP
		expected        []*endpoint.Endpoint
	}{
		{
			title: "IngressRouteTCP with hostname annotation",
			ingressRouteTCP: traefikV1alpha1.IngressRouteTCP{
				TypeMeta: metav1.TypeMeta{
					APIVersion: traefikV1alpha1.SchemeGroupVersion.String(),
					Kind:       "IngressRouteTCP",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroutetcp-annotation",
					Namespace: defaultTraefikNamespace,
					Annotations: map[string]string{
						"external-dns.alpha.kubernetes.io/hostname": "a.example.com",
						"external-dns.alpha.kubernetes.io/target":   "target.domain.tld",
						"kubernetes.io/ingress.class":               "traefik",
					},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName:    "a.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressroutetcp/traefik/ingressroutetcp-annotation",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
			},
		},
		{
			title: "IngressRouteTCP with host sni rule",
			ingressRouteTCP: traefikV1alpha1.IngressRouteTCP{
				TypeMeta: metav1.TypeMeta{
					APIVersion: traefikV1alpha1.SchemeGroupVersion.String(),
					Kind:       "IngressRouteTCP",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroutetcp-hostsni-match",
					Namespace: defaultTraefikNamespace,
					Annotations: map[string]string{
						"external-dns.alpha.kubernetes.io/target": "target.domain.tld",
						"kubernetes.io/ingress.class":             "traefik",
					},
				},
				Spec: traefikV1alpha1.IngressRouteTCPSpec{
					Routes: []traefikV1alpha1.RouteTCP{
						{
							Match: "HostSNI(`b.example.com`)",
						},
					},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName:    "b.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressroutetcp/traefik/ingressroutetcp-hostsni-match",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
			},
		},
		{
			title: "IngressRouteTCP with multiple host sni rules",
			ingressRouteTCP: traefikV1alpha1.IngressRouteTCP{
				TypeMeta: metav1.TypeMeta{
					APIVersion: traefikV1alpha1.SchemeGroupVersion.String(),
					Kind:       "IngressRouteTCP",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroutetcp-multi-host-match",
					Namespace: defaultTraefikNamespace,
					Annotations: map[string]string{
						"external-dns.alpha.kubernetes.io/target": "target.domain.tld",
						"kubernetes.io/ingress.class":             "traefik",
					},
				},
				Spec: traefikV1alpha1.IngressRouteTCPSpec{
					Routes: []traefikV1alpha1.RouteTCP{
						{
							Match: "HostSNI(`d.example.com`) || HostSNI(`e.example.com`)",
						},
					},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName:    "d.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressroutetcp/traefik/ingressroutetcp-multi-host-match",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
				{
					DNSName:    "e.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressroutetcp/traefik/ingressroutetcp-multi-host-match",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
			},
		},
		{
			title: "IngressRouteTCP with multiple host sni rules and annotation",
			ingressRouteTCP: traefikV1alpha1.IngressRouteTCP{
				TypeMeta: metav1.TypeMeta{
					APIVersion: traefikV1alpha1.SchemeGroupVersion.String(),
					Kind:       "IngressRouteTCP",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroutetcp-multi-host-annotations-match",
					Namespace: defaultTraefikNamespace,
					Annotations: map[string]string{
						"external-dns.alpha.kubernetes.io/hostname": "f.example.com",
						"external-dns.alpha.kubernetes.io/target":   "target.domain.tld",
						"kubernetes.io/ingress.class":               "traefik",
					},
				},
				Spec: traefikV1alpha1.IngressRouteTCPSpec{
					Routes: []traefikV1alpha1.RouteTCP{
						{
							Match: "HostSNI(`g.example.com`, `h.example.com`)",
						},
					},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName:    "f.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressroutetcp/traefik/ingressroutetcp-multi-host-annotations-match",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
				{
					DNSName:    "g.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressroutetcp/traefik/ingressroutetcp-multi-host-annotations-match",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
				{
					DNSName:    "h.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressroutetcp/traefik/ingressroutetcp-multi-host-annotations-match",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
			},
		},
		{
			title: "IngressRouteTCP omit wildcard host sni",
			ingressRouteTCP: traefikV1alpha1.IngressRouteTCP{
				TypeMeta: metav1.TypeMeta{
					APIVersion: traefikV1alpha1.SchemeGroupVersion.String(),
					Kind:       "IngressRouteTCP",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressroutetcp-omit-wildcard-host",
					Namespace: defaultTraefikNamespace,
					Annotations: map[string]string{
						"external-dns.alpha.kubernetes.io/target": "target.domain.tld",
						"kubernetes.io/ingress.class":             "traefik",
					},
				},
				Spec: traefikV1alpha1.IngressRouteTCPSpec{
					Routes: []traefikV1alpha1.RouteTCP{
						{
							Match: "HostSNI(`*`)",
						},
					},
				},
			},
			expected: nil,
		},
	} {
		ti := ti
		t.Run(ti.title, func(t *testing.T) {
			t.Parallel()

			fakeKubernetesClient := fakeKube.NewSimpleClientset()
			scheme := runtime.NewScheme()
			scheme.AddKnownTypes(ingressrouteGVR.GroupVersion(), &traefikV1alpha1.IngressRoute{}, &traefikV1alpha1.IngressRouteList{})
			scheme.AddKnownTypes(ingressrouteTCPGVR.GroupVersion(), &traefikV1alpha1.IngressRouteTCP{}, &traefikV1alpha1.IngressRouteTCPList{})
			scheme.AddKnownTypes(ingressrouteUDPGVR.GroupVersion(), &traefikV1alpha1.IngressRouteUDP{}, &traefikV1alpha1.IngressRouteUDPList{})
			fakeDynamicClient := fakeDynamic.NewSimpleDynamicClient(scheme)

			ir := unstructured.Unstructured{}

			ingressRouteAsJSON, err := json.Marshal(ti.ingressRouteTCP)
			assert.NoError(t, err)

			assert.NoError(t, ir.UnmarshalJSON(ingressRouteAsJSON))

			// Create proxy resources
			_, err = fakeDynamicClient.Resource(ingressrouteTCPGVR).Namespace(defaultTraefikNamespace).Create(context.Background(), &ir, metav1.CreateOptions{})
			assert.NoError(t, err)

			source, err := NewTraefikSource(context.TODO(), fakeDynamicClient, fakeKubernetesClient, defaultTraefikNamespace, "kubernetes.io/ingress.class=traefik")
			assert.NoError(t, err)
			assert.NotNil(t, source)

			count := &unstructured.UnstructuredList{}
			for len(count.Items) < 1 {
				count, _ = fakeDynamicClient.Resource(ingressrouteTCPGVR).Namespace(defaultTraefikNamespace).List(context.Background(), metav1.ListOptions{})
			}

			endpoints, err := source.Endpoints(context.Background())
			assert.NoError(t, err)
			assert.Len(t, endpoints, len(ti.expected))
			assert.Equal(t, endpoints, ti.expected)
		})
	}
}

func TestTraefikProxyIngressRouteUDPEndpoints(t *testing.T) {
	t.Parallel()

	for _, ti := range []struct {
		title           string
		ingressRouteUDP traefikV1alpha1.IngressRouteUDP
		expected        []*endpoint.Endpoint
	}{
		{
			title: "IngressRouteTCP with hostname annotation",
			ingressRouteUDP: traefikV1alpha1.IngressRouteUDP{
				TypeMeta: metav1.TypeMeta{
					APIVersion: traefikV1alpha1.SchemeGroupVersion.String(),
					Kind:       "IngressRouteUDP",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressrouteudp-annotation",
					Namespace: defaultTraefikNamespace,
					Annotations: map[string]string{
						"external-dns.alpha.kubernetes.io/hostname": "a.example.com",
						"external-dns.alpha.kubernetes.io/target":   "target.domain.tld",
						"kubernetes.io/ingress.class":               "traefik",
					},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName:    "a.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressrouteudp/traefik/ingressrouteudp-annotation",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
			},
		},
		{
			title: "IngressRouteTCP with multiple hostname annotation",
			ingressRouteUDP: traefikV1alpha1.IngressRouteUDP{
				TypeMeta: metav1.TypeMeta{
					APIVersion: traefikV1alpha1.SchemeGroupVersion.String(),
					Kind:       "IngressRouteUDP",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingressrouteudp-multi-annotation",
					Namespace: defaultTraefikNamespace,
					Annotations: map[string]string{
						"external-dns.alpha.kubernetes.io/hostname": "a.example.com, b.example.com",
						"external-dns.alpha.kubernetes.io/target":   "target.domain.tld",
						"kubernetes.io/ingress.class":               "traefik",
					},
				},
			},
			expected: []*endpoint.Endpoint{
				{
					DNSName:    "a.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressrouteudp/traefik/ingressrouteudp-multi-annotation",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
				{
					DNSName:    "b.example.com",
					Targets:    []string{"target.domain.tld"},
					RecordType: endpoint.RecordTypeCNAME,
					RecordTTL:  0,
					Labels: endpoint.Labels{
						"resource": "ingressrouteudp/traefik/ingressrouteudp-multi-annotation",
					},
					ProviderSpecific: endpoint.ProviderSpecific{},
				},
			},
		},
	} {
		ti := ti
		t.Run(ti.title, func(t *testing.T) {
			t.Parallel()

			fakeKubernetesClient := fakeKube.NewSimpleClientset()
			scheme := runtime.NewScheme()
			scheme.AddKnownTypes(ingressrouteGVR.GroupVersion(), &traefikV1alpha1.IngressRoute{}, &traefikV1alpha1.IngressRouteList{})
			scheme.AddKnownTypes(ingressrouteTCPGVR.GroupVersion(), &traefikV1alpha1.IngressRouteTCP{}, &traefikV1alpha1.IngressRouteTCPList{})
			scheme.AddKnownTypes(ingressrouteUDPGVR.GroupVersion(), &traefikV1alpha1.IngressRouteUDP{}, &traefikV1alpha1.IngressRouteUDPList{})
			fakeDynamicClient := fakeDynamic.NewSimpleDynamicClient(scheme)

			ir := unstructured.Unstructured{}

			ingressRouteAsJSON, err := json.Marshal(ti.ingressRouteUDP)
			assert.NoError(t, err)

			assert.NoError(t, ir.UnmarshalJSON(ingressRouteAsJSON))

			// Create proxy resources
			_, err = fakeDynamicClient.Resource(ingressrouteUDPGVR).Namespace(defaultTraefikNamespace).Create(context.Background(), &ir, metav1.CreateOptions{})
			assert.NoError(t, err)

			source, err := NewTraefikSource(context.TODO(), fakeDynamicClient, fakeKubernetesClient, defaultTraefikNamespace, "kubernetes.io/ingress.class=traefik")
			assert.NoError(t, err)
			assert.NotNil(t, source)

			count := &unstructured.UnstructuredList{}
			for len(count.Items) < 1 {
				count, _ = fakeDynamicClient.Resource(ingressrouteUDPGVR).Namespace(defaultTraefikNamespace).List(context.Background(), metav1.ListOptions{})
			}

			endpoints, err := source.Endpoints(context.Background())
			assert.NoError(t, err)
			assert.Len(t, endpoints, len(ti.expected))
			assert.Equal(t, endpoints, ti.expected)
		})
	}
}
