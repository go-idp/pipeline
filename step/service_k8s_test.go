package step

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	ktesting "k8s.io/client-go/testing"
)

// testK8sMapper builds a RESTMapper covering v1 and apps/v1 resources.
func testK8sMapper() meta.RESTMapper {
	m := meta.NewDefaultRESTMapper([]schema.GroupVersion{
		{Group: "", Version: "v1"},
		{Group: "apps", Version: "v1"},
	})
	m.Add(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"}, meta.RESTScopeRoot)
	m.Add(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"}, meta.RESTScopeNamespace)
	m.Add(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"}, meta.RESTScopeNamespace)
	m.Add(schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, meta.RESTScopeNamespace)
	return m
}

// applyRecordingClient builds a fake dynamic client whose "patch" reactor
// records every apply call (resource, namespace, name) and returns the
// decoded object. This avoids the fake tracker's strategic-merge-patch
// limitation on unstructured objects while still exercising the full
// k8sApply pipeline (decode -> mapping -> namespace scoping -> patch).
func applyRecordingClient() (*dynamicfake.FakeDynamicClient, *[]ktesting.PatchActionImpl) {
	dynClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	var actions []ktesting.PatchActionImpl

	dynClient.Fake.PrependReactor("patch", "*", func(action ktesting.Action) (bool, runtime.Object, error) {
		pa := action.(ktesting.PatchActionImpl)
		actions = append(actions, pa)

		obj, _, err := unstructured.UnstructuredJSONScheme.Decode(pa.GetPatch(), nil, nil)
		if err != nil {
			return true, nil, err
		}
		return true, obj, nil
	})

	return dynClient, &actions
}

func TestK8sApply_EndToEnd(t *testing.T) {
	s := &Step{
		Name: "deploy",
		Service: &Service{
			Type: "kubernetes",
			Config: `
apiVersion: v1
kind: Namespace
metadata:
  name: demo
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 2
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
        - name: web
          image: nginx:alpine
---
apiVersion: v1
kind: Service
metadata:
  name: web
  namespace: demo
spec:
  selector:
    app: web
  ports:
    - port: 80
`,
		},
	}

	dynClient, actions := applyRecordingClient()
	applied, err := s.k8sApply(context.Background(), dynClient, testK8sMapper(), "default")
	if err != nil {
		t.Fatalf("k8sApply() error: %v", err)
	}

	if len(applied) != 3 {
		t.Fatalf("expected 3 applied resources, got %d", len(applied))
	}

	// cluster-scoped Namespace: applied without a namespace
	if applied[0].namespace != "" {
		t.Errorf("namespace object should be cluster-scoped, got %q", applied[0].namespace)
	}

	// namespaced object without namespace in the manifest defaults to the passed namespace
	if applied[1].namespace != "default" {
		t.Errorf("deployment namespace should default to 'default', got %q", applied[1].namespace)
	}

	// namespaced object with explicit namespace keeps it
	if applied[2].namespace != "demo" {
		t.Errorf("service namespace mismatch: got %q", applied[2].namespace)
	}

	// verify the patch actions carried the correct namespaces
	if len(*actions) != 3 {
		t.Fatalf("expected 3 patch actions, got %d", len(*actions))
	}
	if (*actions)[0].GetNamespace() != "" {
		t.Errorf("namespace patch should be cluster-scoped, got ns %q", (*actions)[0].GetNamespace())
	}
	if (*actions)[1].GetNamespace() != "default" {
		t.Errorf("deployment patch should target ns 'default', got %q", (*actions)[1].GetNamespace())
	}
	if (*actions)[2].GetNamespace() != "demo" {
		t.Errorf("service patch should target ns 'demo', got %q", (*actions)[2].GetNamespace())
	}
}

func TestK8sApply_ApplyTwiceIsIdempotent(t *testing.T) {
	manifest := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  key: value
`

	s := &Step{Name: "deploy", Service: &Service{Type: "kubernetes", Config: manifest}}

	for i := 0; i < 2; i++ {
		dynClient, _ := applyRecordingClient()
		if _, err := s.k8sApply(context.Background(), dynClient, testK8sMapper(), "default"); err != nil {
			t.Fatalf("apply #%d should succeed (idempotent), error: %v", i+1, err)
		}
	}
}

func TestK8sApply_UnmappedKind(t *testing.T) {
	s := &Step{Name: "deploy", Service: &Service{
		Type: "kubernetes",
		Config: `
apiVersion: example.com/v1
kind: CustomResource
metadata:
  name: cr
`,
	}}

	dynClient, _ := applyRecordingClient()
	_, err := s.k8sApply(context.Background(), dynClient, testK8sMapper(), "default")
	if err == nil {
		t.Fatalf("expected error for unmapped kind")
	}
}

func TestK8sApply_NoObjects(t *testing.T) {
	s := &Step{Name: "deploy", Service: &Service{Type: "kubernetes", Config: "# just a comment"}}
	dynClient, _ := applyRecordingClient()

	_, err := s.k8sApply(context.Background(), dynClient, testK8sMapper(), "default")
	if err == nil {
		t.Fatalf("expected 'no k8s objects' error")
	}
}

func TestK8sApply_FieldManagerSet(t *testing.T) {
	s := &Step{Name: "deploy", Service: &Service{
		Type: "kubernetes",
		Config: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  key: value
`,
	}}

	dynClient, actions := applyRecordingClient()
	if _, err := s.k8sApply(context.Background(), dynClient, testK8sMapper(), "default"); err != nil {
		t.Fatalf("k8sApply() error: %v", err)
	}

	if len(*actions) != 1 {
		t.Fatalf("expected 1 patch action, got %d", len(*actions))
	}
	opts := (*actions)[0].PatchOptions
	if opts.FieldManager != "pipeline" {
		t.Errorf("field manager mismatch: %q", opts.FieldManager)
	}
	if opts.Force == nil || !*opts.Force {
		t.Errorf("apply should be forced (server-side apply semantics)")
	}
}
