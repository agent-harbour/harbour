package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func colimaStatus(profile string) (bool, error) {
	return commandSucceeded("colima", "status", "-p", profile)
}

func currentMountLines(profile string) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	profileConfig := filepath.Join(home, ".colima", profile, "colima.yaml")
	file, err := os.Open(profileConfig)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var mounts []string
	scanner := bufio.NewScanner(file)
	inMounts := false
	location := ""
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "mounts:") {
			inMounts = true
			continue
		}
		if inMounts && trimmed != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			inMounts = false
		}
		if !inMounts {
			continue
		}
		if strings.HasPrefix(trimmed, "- location:") {
			location = strings.TrimSpace(strings.TrimPrefix(trimmed, "- location:"))
			continue
		}
		if strings.HasPrefix(trimmed, "writable:") && location != "" {
			mode := "ro"
			if strings.TrimSpace(strings.TrimPrefix(trimmed, "writable:")) == "true" {
				mode = "rw"
			}
			mounts = append(mounts, fmt.Sprintf("%s|%s", location, mode))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	sort.Strings(mounts)
	return mounts, nil
}

func desiredMountLines(harnessPath string, repoHosts []string) []string {
	mounts := []string{fmt.Sprintf("%s|rw", harnessPath)}
	for _, host := range repoHosts {
		mounts = append(mounts, fmt.Sprintf("%s|rw", host))
	}
	sort.Strings(mounts)
	return mounts
}
