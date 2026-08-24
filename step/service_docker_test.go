package step

import (
	"io"
	"os"
	"testing"
)

// TestNewDockerCliInitializesContext guards against the docker CLI SDK being
// created without Initialize(): in that state CurrentContext() is "" and the
// context store is nil, so dockerCli.Client() fails with
// "unable to resolve docker endpoint: no context store initialized" and
// os.Exit(1)s the whole process (the bug that made docker-swarm / docker-compose
// service steps always fail with "failed to run command: exit status 1").
//
// The test is hermetic: it does not require a running docker daemon (a failed
// ping is tolerated by the CLI SDK) and it isolates HOME so the real
// ~/.docker config is never touched.
func TestNewDockerCliInitializesContext(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("DOCKER_CONTEXT", "")

	dockerClient, err := newDockerClient()
	if err != nil {
		t.Fatalf("newDockerClient: %v", err)
	}
	defer dockerClient.Close()

	var out, errOut io.Writer = os.Stdout, os.Stderr
	cli, err := newDockerCli(dockerClient, out, errOut)
	if err != nil {
		t.Fatalf("newDockerCli: %v", err)
	}

	if got := cli.CurrentContext(); got != "default" {
		t.Fatalf("CurrentContext() = %q, want %q", got, "default")
	}
	if cli.ContextStore() == nil {
		t.Fatal("context store must be initialized")
	}
	// Must return without os.Exit(1) — this was the exact failure point.
	_ = cli.Client()
}
