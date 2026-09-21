package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPromptAgentInstall(t *testing.T) {
	for _, agent := range []string{"codex", "claude"} {
		for _, tc := range []struct {
			input           string
			want, wantError bool
		}{
			{"\n", true, false}, {"yes\n", true, false}, {"n\n", false, false},
			{"no\n", false, false}, {"maybe\nY\n", true, false}, {"", false, true},
		} {
			t.Run(agent+"/"+tc.input, func(t *testing.T) {
				previous := promptInput
				promptInput = bufio.NewReader(strings.NewReader(tc.input))
				t.Cleanup(func() { promptInput = previous })
				_, stderr := captureOutput(t, func() {
					got, err := promptAgentInstall(agent)
					if got != tc.want || (err != nil) != tc.wantError {
						t.Fatalf("got %v, %v", got, err)
					}
				})
				name := "Codex"
				if agent == "claude" {
					name = "Claude Code"
				}
				if !strings.Contains(stderr, "Run the "+name+" installer? [Y/n]") {
					t.Fatalf("wrong prompt: %s", stderr)
				}
			})
		}
	}
}

func TestProvisionAgents(t *testing.T) {
	for _, agent := range []string{"codex", "claude"} {
		for _, tc := range []struct {
			name, installed, pathVersion, failure, wantVersion string
			install, wantError, wantInstaller                  bool
		}{
			{name: "fresh", install: true, wantInstaller: true, wantVersion: "2.1.278"},
			{name: "upgrade", installed: "1.0.0", install: true, wantInstaller: true, wantVersion: "2.1.278"},
			{name: "reinstall", installed: "2.1.278", install: true, wantInstaller: true, wantVersion: "2.1.278"},
			{name: "decline", installed: "1.0.0", wantVersion: "1.0.0"},
			{name: "decline PATH installation", pathVersion: "1.0.0", wantVersion: "1.0.0"},
			{name: "managed installation takes precedence", installed: "2.0.0", pathVersion: "1.0.0", wantVersion: "2.0.0"},
			{name: "new installation takes precedence", pathVersion: "1.0.0", install: true, wantInstaller: true, wantVersion: "2.1.278"},
			{name: "broken managed installation does not fall back", installed: "broken", pathVersion: "1.0.0", wantError: true},
			{name: "decline missing", wantError: true},
			{name: "decline broken", installed: "broken", wantError: true},
			{name: "no version parsing", installed: "custom-build", wantVersion: "custom-build"},
			{name: "replace broken", installed: "broken", install: true, wantInstaller: true, wantVersion: "2.1.278"},
			{name: "download failure", installed: "1.0.0", install: true, failure: "download", wantError: true},
			{name: "installer failure", installed: "1.0.0", install: true, failure: "installer", wantError: true, wantInstaller: true},
			{name: "version failure", installed: "1.0.0", install: true, failure: "version", wantError: true, wantInstaller: true},
		} {
			t.Run(agent+"/"+tc.name, func(t *testing.T) {
				f := newProvisionFixture(t, agent)
				if tc.installed != "" {
					f.writeAgent(tc.installed)
				}
				if tc.pathVersion != "" {
					f.write("test-bin/"+agent, "#!/bin/sh\ntest \"$1\" = --version || exit 1\n"+f.versionOutput(tc.pathVersion))
				}
				t.Setenv("HARBOUR_TEST_FAILURE", tc.failure)
				output, err := f.run(tc.install)
				if (err != nil) != tc.wantError {
					t.Fatalf("error=%v output=%s", err, output)
				}
				if tc.wantVersion != "" && !strings.Contains(output, tc.wantVersion) {
					t.Fatalf("missing version output: %s", output)
				}
				log, _ := os.ReadFile(filepath.Join(f.home, "calls"))
				if strings.Contains(string(log), "installer:") != tc.wantInstaller {
					t.Fatalf("unexpected installer calls: %s", log)
				}
				if tc.wantInstaller {
					want := "installer:latest::\n"
					if agent == "codex" {
						want = "installer:--release latest:1:" + filepath.Join(f.home, ".local/bin") + "\n"
					}
					if string(log) != want {
						t.Fatalf("installer call=%q, want %q", log, want)
					}
				}
				other := "claude"
				if agent == "claude" {
					other = "codex"
				}
				_, otherErr := os.Lstat(filepath.Join(f.home, ".local/bin", other))
				if tc.wantError {
					if otherErr != nil {
						t.Fatalf("previous agent removed: %v", otherErr)
					}
					for name, want := range map[string]string{f.instruction(): "previous instructions", "skills/existing": "previous skills"} {
						data, err := os.ReadFile(filepath.Join(f.home, "."+agent, name))
						if err != nil || string(data) != want {
							t.Fatalf("existing %s changed: %q %v", name, data, err)
						}
					}
				} else {
					if !os.IsNotExist(otherErr) {
						t.Fatalf("other agent command retained: %v", otherErr)
					}
					for _, name := range []string{f.instruction(), "skills"} {
						target, err := os.Readlink(filepath.Join(f.home, "."+agent, name))
						source := "skills"
						if name != "skills" {
							source = "AGENTS.md"
						}
						if err != nil || target != filepath.Join(f.harness, source) {
							t.Fatalf("incorrect %s link: %q %v", name, target, err)
						}
					}
				}
			})
		}
	}
}

type provisionFixture struct {
	t                         *testing.T
	home, harness, bin, agent string
}

func newProvisionFixture(t *testing.T, agent string) provisionFixture {
	t.Helper()
	home := t.TempDir()
	f := provisionFixture{t, home, filepath.Join(home, "harness with spaces"), filepath.Join(home, "test-bin"), agent}
	f.write("harness with spaces/AGENTS.md", "Test harness\n")
	f.write("harness with spaces/skills/example/SKILL.md", "Test skill\n")
	f.write("."+agent+"/skills/existing", "previous skills")
	f.write("."+agent+"/"+f.instruction(), "previous instructions")
	other := "claude"
	if agent == "claude" {
		other = "codex"
	}
	f.write(".local/bin/"+other, "#!/bin/sh\nexit 0\n")
	if err := os.MkdirAll(f.bin, 0755); err != nil {
		t.Fatal(err)
	}
	// Only controlled commands are visible. No real sudo, network or agent installer.
	for _, name := range []string{"sh", "bash", "mkdir", "dirname", "rm", "ln", "mktemp", "cp"} {
		path, err := exec.LookPath(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(path, filepath.Join(f.bin, name)); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"file", "gh", "make", "rg"} {
		f.write("test-bin/"+name, "#!/bin/sh\nexit 0\n")
	}
	f.write("test-bin/sudo", "#!/bin/sh\necho 'unexpected sudo' >&2\nexit 1\n")
	url := "https://claude.ai/install.sh"
	if agent == "codex" {
		url = "https://chatgpt.com/codex/install.sh"
	}
	f.write("test-bin/curl", fmt.Sprintf(`#!/bin/sh
set -eu
test "$1" = -fsSL
test "$2" = '%s'
test "$3" = -o
test "${HARBOUR_TEST_FAILURE:-}" != download
cp "$HOME/installer.sh" "$4"
`, url))
	f.write("installer.sh", `#!/bin/sh
set -eu
printf 'installer:%s:%s:%s\n' "$*" "${CODEX_NON_INTERACTIVE:-}" "${CODEX_INSTALL_DIR:-}" >> "$HOME/calls"
test "${HARBOUR_TEST_FAILURE:-}" != installer
cp "$HOME/agent-fixture" "$HOME/.local/bin/`+agent+`"
`)
	f.write("agent-fixture", `#!/bin/sh
set -eu
test "$1" = --version
test "${HARBOUR_TEST_FAILURE:-}" != version
`+f.versionOutput("2.1.278"))
	return f
}
func (f provisionFixture) instruction() string {
	if f.agent == "claude" {
		return "CLAUDE.md"
	}
	return "AGENTS.md"
}
func (f provisionFixture) versionOutput(version string) string {
	if f.agent == "codex" {
		return "echo \"codex-cli " + version + "\"\n"
	}
	return "echo \"" + version + " (Claude Code)\"\n"
}
func (f provisionFixture) writeAgent(version string) {
	contents := "#!/bin/sh\n" + f.versionOutput(version)
	if version == "broken" {
		contents = "#!/bin/sh\nexit 1\n"
	}
	f.write(".local/bin/"+f.agent, contents)
}
func (f provisionFixture) write(name, contents string) {
	f.t.Helper()
	path := filepath.Join(f.home, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0755); err != nil {
		f.t.Fatal(err)
	}
}
func (f provisionFixture) run(install bool) (string, error) {
	args := []string{f.agent, fmt.Sprint(install), filepath.Join(f.harness, "AGENTS.md"), filepath.Join(f.harness, "skills"), "", "0", "0"}
	cmd := exec.Command("bash", append([]string{"-c", provisionVMScript, "provision_vm.sh"}, args...)...)
	cmd.Dir = f.home
	cmd.Env = append(os.Environ(), "HOME="+f.home, "PATH="+f.bin)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
