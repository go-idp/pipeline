package step

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// appliedResource records a successfully applied k8s object for readiness
// checks and diagnostics.
type appliedResource struct {
	gvr       schema.GroupVersionResource
	namespace string
	name      string
	kind      string
}

// runServiceKubernetes applies the k8s manifests natively with client-go
// server-side apply, equivalent to `kubectl apply -f -`.
func (s *Step) runServiceKubernetes(ctx context.Context) error {
	s.serviceLogf("kubernetes: apply manifests (timeout: %ds)", s.Service.Timeout)

	restCfg, namespace, err := s.kubeClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load kubernetes config: %s", err)
	}

	dynClient, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes dynamic client: %s", err)
	}

	kubeClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %s", err)
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(kubeClient.Discovery()))

	applied, err := s.k8sApply(ctx, dynClient, mapper, namespace)
	if err != nil {
		return err
	}

	if err := s.waitK8sReady(ctx, kubeClient, applied); err != nil {
		s.collectK8sDiagnostics(ctx, kubeClient, namespace)
		return err
	}

	s.serviceLogf("kubernetes: applied %d object(s) successfully", len(applied))
	return nil
}

// kubeClientConfig builds the rest config and default namespace from the
// service kubeconfig, $KUBECONFIG, ~/.kube/config, or in-cluster config.
func (s *Step) kubeClientConfig() (*rest.Config, string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if s.Service.Kubeconfig != "" {
		loadingRules.ExplicitPath = s.Service.Kubeconfig
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

	restCfg, err := clientConfig.ClientConfig()
	if err != nil {
		// fall back to in-cluster config
		inClusterCfg, inClusterErr := rest.InClusterConfig()
		if inClusterErr != nil {
			return nil, "", fmt.Errorf("failed to get kubeconfig (%s) and in-cluster config (%s)", err, inClusterErr)
		}
		return inClusterCfg, s.Service.Namespace, nil
	}

	namespace := s.Service.Namespace
	if namespace == "" {
		if ns, _, nsErr := clientConfig.Namespace(); nsErr == nil {
			namespace = ns
		}
	}
	if namespace == "" {
		namespace = "default"
	}

	return restCfg, namespace, nil
}

// k8sApply parses the multi-document manifest and server-side applies every
// object, returning the list of applied resources.
func (s *Step) k8sApply(ctx context.Context, dynClient dynamic.Interface, mapper meta.RESTMapper, namespace string) ([]appliedResource, error) {
	objects, err := decodeK8sManifests(s.Service.Config)
	if err != nil {
		return nil, err
	}

	var applied []appliedResource

	for _, u := range objects {
		gvk := u.GroupVersionKind()
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve resource mapping for %s %s: %s", gvk.Kind, u.GetName(), err)
		}

		ns := u.GetNamespace()
		// only default the namespace for namespaced resources;
		// cluster-scoped resources (Namespace, ClusterRole, CRD, ...) must stay empty
		if ns == "" && mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			ns = namespace
		}

		data, err := json.Marshal(u.Object)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal %s %s: %s", gvk.Kind, u.GetName(), err)
		}

		force := true
		ri := dynClient.Resource(mapping.Resource).Namespace(ns)
		if _, err := ri.Patch(ctx, u.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
			FieldManager: "pipeline",
			Force:        &force,
		}); err != nil {
			return nil, fmt.Errorf("failed to apply %s %s: %s", gvk.Kind, u.GetName(), err)
		}

		s.serviceLogf("kubernetes: applied %s/%s (%s)", ns, u.GetName(), gvk.Kind)
		applied = append(applied, appliedResource{
			gvr:       mapping.Resource,
			namespace: ns,
			name:      u.GetName(),
			kind:      gvk.Kind,
		})
	}

	if len(applied) == 0 {
		return nil, fmt.Errorf("no k8s objects found in the manifest")
	}

	return applied, nil
}

// decodeK8sManifests parses a multi-document k8s manifest into unstructured
// objects, skipping empty documents.
func decodeK8sManifests(config string) ([]*unstructured.Unstructured, error) {
	decoder := k8syaml.NewYAMLOrJSONDecoder(strings.NewReader(config), 4096)

	var objects []*unstructured.Unstructured
	docIndex := 0

	for {
		docIndex++
		var rawJSON json.RawMessage
		if err := decoder.Decode(&rawJSON); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to parse k8s manifest document %d: %s", docIndex, err)
		}

		// skip empty documents
		trimmed := strings.TrimSpace(string(rawJSON))
		if trimmed == "" || trimmed == "null" {
			continue
		}

		obj, _, err := unstructured.UnstructuredJSONScheme.Decode(rawJSON, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decode k8s manifest document %d: %s", docIndex, err)
		}

		u, ok := obj.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("unsupported object type in document %d", docIndex)
		}

		gvk := u.GroupVersionKind()
		if gvk.Kind == "" || gvk.Version == "" {
			return nil, fmt.Errorf("manifest document %d is missing apiVersion/kind", docIndex)
		}

		objects = append(objects, u)
	}

	if len(objects) == 0 {
		return nil, fmt.Errorf("no k8s objects found in the manifest")
	}

	return objects, nil
}

// waitK8sReady polls the applied Deployments until every replica is ready,
// bounded by the service timeout. Other resource kinds are considered ready
// once applied.
func (s *Step) waitK8sReady(ctx context.Context, kubeClient *kubernetes.Clientset, applied []appliedResource) error {
	var deployments []appliedResource
	for _, r := range applied {
		if r.kind == "Deployment" && r.gvr.Group == "apps" {
			deployments = append(deployments, r)
		}
	}

	if len(deployments) == 0 {
		return nil
	}

	timeout := time.Duration(s.Service.Timeout) * time.Second
	interval := 6 * time.Second
	deadline := time.Now().Add(timeout)
	checks := 0

	for {
		checks++
		notReady := []string{}
		for _, d := range deployments {
			dep, err := kubeClient.AppsV1().Deployments(d.namespace).Get(ctx, d.name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get deployment %s/%s: %s", d.namespace, d.name, err)
			}

			var desired int32
			if dep.Spec.Replicas != nil {
				desired = *dep.Spec.Replicas
			}

			if dep.Status.ReadyReplicas < desired {
				notReady = append(notReady, fmt.Sprintf("%s/%s (%d/%d)", d.namespace, d.name, dep.Status.ReadyReplicas, desired))
			}
		}

		if len(notReady) == 0 {
			s.serviceLogf("kubernetes: deployments ready (check %d)", checks)
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("kubernetes deployment startup timeout after %d checks, not ready: %s", checks, strings.Join(notReady, "; "))
		}

		s.serviceLogf("kubernetes: waiting for deployment startup (check %d/%d): %s", checks, int(timeout/interval), strings.Join(notReady, "; "))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}

// collectK8sDiagnostics prints pods and events on failure, mirroring the
// diagnostics of the legacy shell implementation.
func (s *Step) collectK8sDiagnostics(ctx context.Context, kubeClient *kubernetes.Clientset, namespace string) {
	s.serviceLogf("kubernetes: collecting cluster diagnostics (namespace: %s) ...", namespace)

	pods, err := kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		s.serviceLogf("kubernetes: failed to list pods: %s", err)
	} else {
		for _, pod := range pods.Items {
			phase := string(pod.Status.Phase)
			ready := ""
			for _, c := range pod.Status.Conditions {
				if c.Type == "Ready" {
					ready = string(c.Status)
				}
			}
			s.serviceLogf("kubernetes: pod %s phase=%s ready=%s", pod.Name, phase, ready)
		}
	}

	events, err := kubeClient.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		s.serviceLogf("kubernetes: failed to list events: %s", err)
	} else {
		count := 0
		for _, e := range events.Items {
			if count >= 80 {
				break
			}
			s.serviceLogf("kubernetes: event %s %s: %s", e.InvolvedObject.Kind, e.InvolvedObject.Name, e.Message)
			count++
		}
	}
}
