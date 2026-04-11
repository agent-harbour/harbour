package vm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Lima struct {
	cfg Config
}

var _ Backend = Lima{}

func (Lima) Name() string {
	return "Lima"
}

func (Lima) EnsureInstalled() error {
	if err := ensureCommand("limactl"); err != nil {
		return fmt.Errorf("limactl is required for Harbour. Install Lima first: %w", err)
	}
	return nil
}

func (l Lima) Status() (bool, error) {
	if !l.instanceExists() {
		return false, nil
	}
	status, err := commandOutput("limactl", "list", l.cfg.Profile, "--format", "{{.Status}}")
	if err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(status), "running"), nil
}

func (l Lima) CurrentMountLines() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return readMountLines(filepath.Join(home, ".lima", l.cfg.Profile, "lima.yaml"))
}

func (l Lima) Start(mounts []string) error {
	if !l.instanceExists() {
		if err := l.createInstance(mounts); err != nil {
			return err
		}
	}
	return runCommand("limactl", "start", l.cfg.Profile)
}

func (l Lima) Reconfigure(mounts []string) error {
	if l.instanceExists() {
		if err := l.Stop(); err != nil {
			return err
		}
		if err := runCommand("limactl", "delete", "--force", l.cfg.Profile); err != nil {
			return err
		}
	}
	if err := l.createInstance(mounts); err != nil {
		return err
	}
	return runCommand("limactl", "start", l.cfg.Profile)
}

func (l Lima) Stop() error {
	if !l.instanceExists() {
		return nil
	}
	return runCommand("limactl", "stop", l.cfg.Profile)
}

func (l Lima) RunRemoteCommand(command string) error {
	return runCommand("limactl", "shell", l.cfg.Profile, "/usr/bin/bash", "-lc", command)
}

func (l Lima) RunRemoteScript(script string, args []string) error {
	shellArgs := append([]string{
		"shell", l.cfg.Profile, "/usr/bin/bash", "-s", "--",
	}, args...)
	return runCommandInput(script, "limactl", shellArgs...)
}

func (l Lima) instanceExists() bool {
	_, err := commandOutput("limactl", "list", l.cfg.Profile, "--format", "{{.Name}}")
	return err == nil
}

func (l Lima) createInstance(mounts []string) error {
	args := []string{
		"create",
		"--tty=false",
		"--name", l.cfg.Profile,
		"--vm-type", l.cfg.Type,
		"--arch", l.cfg.Arch,
		"--cpus", fmt.Sprintf("%d", l.cfg.CPU),
		"--memory", fmt.Sprintf("%d", l.cfg.Memory),
		"--disk", fmt.Sprintf("%d", l.cfg.Disk),
		"--mount-type", l.cfg.MountType,
	}
	if l.cfg.NetworkAddress {
		args = append(args, "--network", "lima:shared")
	}
	for _, mount := range mounts {
		args = append(args, "--mount", fmt.Sprintf("%s:w", mount))
	}

	template, err := l.template()
	if err != nil {
		return err
	}
	if template != "" {
		args = append(args, template)
	}

	fmt.Printf("Executing:\n  limactl %s\n", shellQuoteArgs(args))
	return runCommand("limactl", args...)
}

func (l Lima) template() (string, error) {
	switch l.cfg.Runtime {
	case "", "containerd":
		return "", nil
	case "docker":
		return "template:docker", nil
	default:
		return "", fmt.Errorf("unsupported vm_runtime=%s for lima", l.cfg.Runtime)
	}
}
