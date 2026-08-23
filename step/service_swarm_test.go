package step

import (
	"strings"
	"testing"

	swarmtypes "github.com/docker/docker/api/types/swarm"
)

func replicas(n uint64) *uint64 {
	return &n
}

func testSwarmService(name string, mode swarmtypes.ServiceMode, status *swarmtypes.ServiceStatus) swarmtypes.Service {
	return swarmtypes.Service{
		Spec: swarmtypes.ServiceSpec{
			Annotations: swarmtypes.Annotations{
				Name: name,
			},
			Mode: mode,
		},
		ServiceStatus: status,
	}
}

func TestSwarmServicesReady_AllReady(t *testing.T) {
	services := []swarmtypes.Service{
		testSwarmService("eunomia-task-1_web", swarmtypes.ServiceMode{
			Replicated: &swarmtypes.ReplicatedService{Replicas: replicas(2)},
		}, &swarmtypes.ServiceStatus{RunningTasks: 2, DesiredTasks: 2}),
		testSwarmService("eunomia-task-1_db", swarmtypes.ServiceMode{
			Replicated: &swarmtypes.ReplicatedService{Replicas: replicas(1)},
		}, &swarmtypes.ServiceStatus{RunningTasks: 1, DesiredTasks: 1}),
	}

	ready, notReady := swarmServicesReady(services, "eunomia-task-1")
	if !ready {
		t.Fatalf("expected ready, notReady: %v", notReady)
	}
	if len(notReady) != 0 {
		t.Fatalf("expected no not-ready services, got %v", notReady)
	}
}

func TestSwarmServicesReady_NotEnoughReplicas(t *testing.T) {
	services := []swarmtypes.Service{
		testSwarmService("eunomia-task-1_web", swarmtypes.ServiceMode{
			Replicated: &swarmtypes.ReplicatedService{Replicas: replicas(2)},
		}, &swarmtypes.ServiceStatus{RunningTasks: 1, DesiredTasks: 2}),
	}

	ready, notReady := swarmServicesReady(services, "eunomia-task-1")
	if ready {
		t.Fatalf("expected not ready")
	}
	if len(notReady) != 1 {
		t.Fatalf("expected 1 not-ready service, got %v", notReady)
	}
	if notReady[0] != "web (1/2)" {
		t.Errorf("not-ready description mismatch: %q", notReady[0])
	}
}

func TestSwarmServicesReady_NilServiceStatus(t *testing.T) {
	// a service with nil ServiceStatus is still being scheduled => not ready
	services := []swarmtypes.Service{
		testSwarmService("eunomia-task-1_web", swarmtypes.ServiceMode{
			Replicated: &swarmtypes.ReplicatedService{Replicas: replicas(1)},
		}, nil),
	}

	ready, _ := swarmServicesReady(services, "eunomia-task-1")
	if ready {
		t.Fatalf("expected not ready when ServiceStatus is nil")
	}
}

func TestSwarmServicesReady_GlobalMode(t *testing.T) {
	// global mode: at least one running task
	services := []swarmtypes.Service{
		testSwarmService("eunomia-task-1_agent", swarmtypes.ServiceMode{
			Global: &swarmtypes.GlobalService{},
		}, &swarmtypes.ServiceStatus{RunningTasks: 1, DesiredTasks: 1}),
	}

	ready, notReady := swarmServicesReady(services, "eunomia-task-1")
	if !ready {
		t.Fatalf("expected ready, notReady: %v", notReady)
	}
}

func TestSwarmServicesReady_EmptyList(t *testing.T) {
	// no services found for the stack: treat as ready (nothing to wait for)
	ready, notReady := swarmServicesReady(nil, "eunomia-task-1")
	if !ready {
		t.Fatalf("expected ready for empty service list, notReady: %v", notReady)
	}
}

func TestSwarmServicesReady_StackPrefixTrimmed(t *testing.T) {
	services := []swarmtypes.Service{
		testSwarmService("eunomia-task-1_web", swarmtypes.ServiceMode{
			Replicated: &swarmtypes.ReplicatedService{Replicas: replicas(3)},
		}, &swarmtypes.ServiceStatus{RunningTasks: 0, DesiredTasks: 3}),
	}

	_, notReady := swarmServicesReady(services, "eunomia-task-1")
	if len(notReady) != 1 || notReady[0] != "web (0/3)" {
		t.Errorf("expected stack prefix trimmed in description, got %v", notReady)
	}
	if strings.Contains(notReady[0], "eunomia-task-1_") {
		t.Errorf("service name should not contain the stack prefix, got %q", notReady[0])
	}
}
