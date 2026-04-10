package vm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Colima struct {
	cfg Config
}

var _ Backend = Colima{}

func (Colima) Name() string {
	return "Colima"
}

func (Colima) EnsureInstalled() error {
	if err := ensureCommand("colima"); err != nil {
		return fmt.Errorf("colima is required for Harbour. Install it with: brew install colima: %w", err)
	}
	return nil
}

func (c Colima) Status() (bool, error) {
	return commandSucceeded("colima", "status", "-p", c.cfg.Profile)
}

func (c Colima) CurrentMountLines() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return readMountLines(filepath.Join(home, ".colima", c.cfg.Profile, "colima.yaml"))
}

func (c Colima) Start(mounts []string) error {
	args := []string{
		"start", c.cfg.Profile,
		"--runtime", c.cfg.Runtime,
		"--vm-type", c.cfg.Type,
		"--arch", c.cfg.Arch,
		"--cpu", fmt.Sprintf("%d", c.cfg.CPU),
		"--memory", fmt.Sprintf("%d", c.cfg.Memory),
		"--disk", fmt.Sprintf("%d", c.cfg.Disk),
		"--mount-type", c.cfg.MountType,
	}
	if c.cfg.ForwardSSHAgent {
		args = append(args, "--ssh-agent")
	}
	if c.cfg.NetworkAddress {
		args = append(args, "--network-address")
	}
	for _, mount := range mounts {
		args = append(args, "--mount", fmt.Sprintf("%s:w", mount))
	}
	fmt.Printf("Executing:\n  colima %s\n", shellQuoteArgs(args))
	return runCommand("colima", args...)
}

func (c Colima) Reconfigure(mounts []string) error {
	if err := c.Stop(); err != nil {
		return err
	}
	return c.Start(mounts)
}

func (c Colima) Stop() error {
	return runCommand("colima", "stop", "-p", c.cfg.Profile)
}

func (c Colima) RunRemoteCommand(command string) error {
	return runCommand("colima", "ssh", "-p", c.cfg.Profile, "--", "/usr/bin/bash", "-lc", command)
}

func (c Colima) RunRemoteScript(script string, args []string) error {
	sshArgs := append([]string{
		"ssh", "-p", c.cfg.Profile, "--", "/usr/bin/bash", "-s", "--",
	}, args...)
	return runCommandInput(script, "colima", sshArgs...)
}

func ensureCommand(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s is required but not installed", name)
	}
	return nil
}

func runCommand(name string, args ...string) error {
	if err := ensureCommand(name); err != nil {
		return err
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommandInput(input string, name string, args ...string) error {
	if err := ensureCommand(name); err != nil {
		return err
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func commandSucceeded(name string, args ...string) (bool, error) {
	if err := ensureCommand(name); err != nil {
		return false, err
	}
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if _, ok := err.(*exec.ExitError); ok {
		return false, nil
	}
	return false, err
}

func commandOutput(name string, args ...string) (string, error) {
	if err := ensureCommand(name); err != nil {
		return "", err
	}
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr := strings.TrimSpace(string(exitErr.Stderr))
		if stderr == "" {
			return "", exitErr
		}
		return "", fmt.Errorf("%s", stderr)
	}
	return "", err
}

func shellQuoteArgs(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, fmt.Sprintf("%q", arg))
	}
	return strings.Join(quoted, " ")
}
