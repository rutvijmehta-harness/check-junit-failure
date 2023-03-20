package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-zglob"
	"github.com/rutvijmehta-harness/check-junit-failure/gojunit"
	"github.com/sirupsen/logrus"
)

const (
	globEnv = "PLUGIN_JUNIT_PATH"
)

// ParseTests parses XMLs and returns error if there are any failures
func ParseTests(paths []string, log *logrus.Logger) error {
	files := getFiles(paths, log)

	if len(files) == 0 {
		log.Errorln("could not find any files matching the provided report path")
		return nil
	}

	for _, file := range files {
		suites, err := gojunit.IngestFile(file)
		if err != nil {
			log.WithError(err).WithField("file", file).
				Errorln(fmt.Sprintf("could not parse file %s", file))
			continue
		}
		for _, suite := range suites { //nolint:gocritic
			for _, test := range suite.Tests { //nolint:gocritic
				if test.Result.Status == "failed" {
					return fmt.Errorf("found failed test with class %s and testcase %s", test.Classname, test.Name)
				}
			}
		}
	}
	return nil
}

// getFiles returns uniques file paths provided in the input after expanding the input paths
func getFiles(paths []string, log *logrus.Logger) []string {
	var files []string
	for _, p := range paths {
		path, err := expandTilde(p)
		if err != nil {
			log.WithError(err).WithField("path", p).
				Errorln("errored while trying to expand paths")
			continue
		}
		matches, err := zglob.Glob(path)
		if err != nil {
			log.WithError(err).WithField("path", path).
				Errorln("errored while trying to resolve path regex")
			continue
		}

		files = append(files, matches...)
	}
	return uniqueItems(files)
}

func uniqueItems(items []string) []string {
	var result []string

	set := make(map[string]bool)
	for _, item := range items {
		if _, ok := set[item]; !ok {
			result = append(result, item)
			set[item] = true
		}
	}
	return result
}

// expandTilde method expands the given file path to include the home directory
// if the path is prefixed with `~`. If it isn't prefixed with `~`, the path is
// returned as-is.
func expandTilde(path string) (string, error) {
	if path == "" {
		return path, nil
	}

	if path[0] != '~' {
		return path, nil
	}

	if len(path) > 1 && path[1] != '/' && path[1] != '\\' {
		return "", errors.New("cannot expand user-specific home dir")
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to fetch home directory: %s", err)
	}
	return filepath.Join(dir, path[1:]), nil
}

func getPaths(globVal string) []string {
	paths := make([]string, 0)
	globValSplit := strings.Split(globVal, ",")
	for _, val := range globValSplit {
		if val == "" {
			continue
		}
		val = strings.TrimSpace(val)
		paths = append(paths, val)
	}
	return paths
}

func main() {
	log := logrus.New()
	log.Out = os.Stdout

	// Read globs from env variable
	globVal := os.Getenv(globEnv)
	if globVal == "" {
		log.Errorln(fmt.Errorf("%s env variable is not set", globEnv))
		os.Exit(1)
	}

	paths := getPaths(globVal)
	log.Infoln(fmt.Sprintf("Parsing test cases in globs: %s", paths))
	if err := ParseTests(paths, log); err != nil {
		log.Errorln(fmt.Sprintf("Error while parsing tests: %s", err))
		os.Exit(1)
	}
}
