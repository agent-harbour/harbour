package vm

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func readMountLines(path string) ([]string, error) {
	file, err := os.Open(path)
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
	return normalizeMountLines(mounts), nil
}

func normalizeMountLines(mounts []string) []string {
	if len(mounts) == 0 {
		return nil
	}

	sorted := append([]string(nil), mounts...)
	sort.Strings(sorted)

	normalized := sorted[:0]
	for _, mount := range sorted {
		if len(normalized) > 0 && normalized[len(normalized)-1] == mount {
			continue
		}
		normalized = append(normalized, mount)
	}
	return normalized
}
