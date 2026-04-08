package vm

import "fmt"

type Config struct {
	Backend         string
	Profile         string
	Runtime         string
	Type            string
	Arch            string
	CPU             int
	Memory          int
	Disk            int
	MountType       string
	ForwardSSHAgent bool
	NetworkAddress  bool
}

type Backend interface {
	Name() string
	EnsureInstalled() error
	Status(cfg Config) (bool, error)
	CurrentMountLines(cfg Config) ([]string, error)
	Start(cfg Config, mounts []string) error
	Stop(cfg Config) error
	RunRemoteCommand(cfg Config, command string) error
	RunRemoteScript(cfg Config, script string, args []string) error
}

func Resolve(cfg Config) (Backend, error) {
	switch cfg.Backend {
	case "colima":
		return ColimaBackend{}, nil
	default:
		return nil, fmt.Errorf("unsupported vm_backend=%s", cfg.Backend)
	}
}
