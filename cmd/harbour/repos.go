package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func reposFilePath(harnessPath string) string {
	return filepath.Join(harnessPath, "repos.yaml")
}

func parseRepoHosts(reposFile string, workspaceRoot string) ([]string, error) {
	file, err := os.Open(reposFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var hosts []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "- ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "- "))
		}
		if !strings.HasPrefix(line, "host_path:") {
			continue
		}
		raw := strings.TrimSpace(strings.TrimPrefix(line, "host_path:"))
		if idx := strings.Index(raw, "#"); idx >= 0 {
			raw = strings.TrimSpace(raw[:idx])
		}
		raw, err = expandHome(raw)
		if err != nil {
			return nil, err
		}
		if raw == "" {
			continue
		}
		if filepath.IsAbs(raw) {
			hosts = append(hosts, raw)
			continue
		}
		if workspaceRoot == "" {
			return nil, fmt.Errorf("workspace_root is not set. Configure it in the Harbour config or run harbour provision")
		}
		hosts = append(hosts, filepath.Join(workspaceRoot, raw))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return hosts, nil
}

func existingRepoHosts(reposFile string, workspaceRoot string, warnMissing bool) ([]string, error) {
	hosts, err := parseRepoHosts(reposFile, workspaceRoot)
	if err != nil {
		return nil, err
	}
	var existing []string
	for _, host := range hosts {
		if info, err := os.Stat(host); err == nil && info.IsDir() {
			existing = append(existing, host)
			continue
		}
		if warnMissing {
			fmt.Fprintf(os.Stderr, "Warning: skipping missing repo mount %s\n", host)
		}
	}
	return existing, nil
}
