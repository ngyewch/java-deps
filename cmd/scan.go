package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"
)

func doScan(ctx context.Context, cmd *cli.Command) error {
	outputFile := cmd.StringArg(outputFileArg.Name)
	inputDir := cmd.StringArg(pathArg.Name)

	if outputFile == "" {
		return fmt.Errorf("%s is required", outputFileArg.Name)
	}
	if inputDir == "" {
		return fmt.Errorf("%s is required", pathArg.Name)
	}

	scanner := NewScanner(outputFile, inputDir)
	err := scanner.Scan()
	if err != nil {
		return err
	}

	return nil
}

type Scanner struct {
	outputFile string
	inputDir   string
}

func NewScanner(outputFile string, inputDir string) *Scanner {
	return &Scanner{
		outputFile: outputFile,
		inputDir:   inputDir,
	}
}

func (scanner *Scanner) Scan() error {
	dirEntries, err := os.ReadDir(scanner.inputDir)
	if err != nil {
		return err
	}
	var artifactEntries []*ArtifactEntry
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() || !strings.HasSuffix(dirEntry.Name(), ".jar") {
			continue
		}
		path := filepath.Join(scanner.inputDir, dirEntry.Name())
		entry, err := ScanJar(path)
		if err != nil {
			return err
		}
		if entry != nil {
			artifactEntries = append(artifactEntries, entry)
		}
	}
	slices.SortStableFunc(artifactEntries, func(entry1 *ArtifactEntry, entry2 *ArtifactEntry) int {
		return entry1.ArtifactInfo.Compare(entry2.ArtifactInfo)
	})

	f, err := os.Create(scanner.outputFile)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	jsonEncoder := json.NewEncoder(f)
	jsonEncoder.SetIndent("", "  ")
	err = jsonEncoder.Encode(artifactEntries)
	if err != nil {
		return err
	}

	return nil
}
