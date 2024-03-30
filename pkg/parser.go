package pkg

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

func LoadNamedConfYAML(zones []SlaveZone) (string, error) {
	yamlData, err := yaml.Marshal(&zones)
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}

	// For demonstration, let's print the YAML to stdout
	fmt.Println(string(yamlData))

	// Optionally, write to a temporary file
	tempFile, err := os.CreateTemp("", "zones-*.yaml")
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %v", err)
	}
	defer tempFile.Close()

	_, err = io.WriteString(tempFile, string(yamlData))
	if err != nil {
		return "", fmt.Errorf("error writing to temp file: %v", err)
	}

	// Ensure you flush any buffered data
	if err := tempFile.Sync(); err != nil {
		return "", fmt.Errorf("error flushing temp file: %v", err)
	}
	return tempFile.Name(), nil
}

func ParseNamedConf(filePath string) ([]SlaveZone, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var zones []SlaveZone
	var currentZone *SlaveZone

	scanner := bufio.NewScanner(file)
	zonePattern := regexp.MustCompile(`zone\s+"(.+)"\s+in\s+\{`)
	masterPattern := regexp.MustCompile(`masters\s+\{`)
	filePattern := regexp.MustCompile(`file\s+"(.+)";`)
	masterLinePattern := regexp.MustCompile(`[\s\t]+(((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4})`)
	inZone := false
	inMasters := false
	// isSlave := false

	// This is a very basic parser, eventually we need to improve it
	for scanner.Scan() {
		line := scanner.Text()
		if zonePattern.MatchString(line) {
			// fmt.Println(line)
			matches := zonePattern.FindStringSubmatch(line)
			domain := matches[1]
			if !strings.HasSuffix(domain, ".") {
				domain += "."
			}
			currentZone = &SlaveZone{Name: domain}
			inZone = true
			// } else if strings.Contains(line, "type slave") {
			// 	isSlave = true
		} else if inZone && masterPattern.MatchString(line) {
			// fmt.Println(line)
			inMasters = true
		} else if inZone && inMasters && masterLinePattern.MatchString(line) {
			// fmt.Println(line)
			matches := masterLinePattern.FindStringSubmatch(line)
			currentZone.Masters = append(currentZone.Masters, matches[1])
		} else if inZone && inMasters && strings.Contains(line, "};") {
			inMasters = false
		} else if inZone && filePattern.MatchString(line) {
			matches := filePattern.FindStringSubmatch(line)
			currentZone.File = matches[1]
			zones = append(zones, *currentZone)
			inZone = false
			inMasters = false
			// isSlave = false
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	// return nil, nil
	return zones, nil
}
