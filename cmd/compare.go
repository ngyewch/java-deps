package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v3"
)

func doCompare(ctx context.Context, cmd *cli.Command) error {
	inputFile := cmd.StringArg(inputFileArg.Name)
	inputFile2 := cmd.StringArg(inputFile2Arg.Name)

	entries1, err := loadEntries(inputFile)
	if err != nil {
		return err
	}
	entries2, err := loadEntries(inputFile2)
	if err != nil {
		return err
	}

	indexMap1 := make(map[int]bool)
	indexMap2 := make(map[int]bool)

	processor := func(handler func(i int, entry1 *ArtifactEntry, j int, entry2 *ArtifactEntry) (bool, error)) error {
		for i, entry1 := range entries1 {
			processed1 := indexMap1[i]
			if processed1 {
				continue
			}
			for j, entry2 := range entries2 {
				processed2 := indexMap2[j]
				if processed2 {
					continue
				}
				stop, err := handler(i, entry1, j, entry2)
				if err != nil {
					return err
				}
				if stop {
					break
				}
			}
		}
		return nil
	}

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()

	// check for same dependencies
	{
		var matchingEntries []*ArtifactEntry
		err = processor(func(index1 int, entry1 *ArtifactEntry, index2 int, entry2 *ArtifactEntry) (bool, error) {
			if (entry1.Filename == entry2.Filename) && (entry1.Size == entry2.Size) && (entry1.Sha256 == entry2.Sha256) && (entry1.ArtifactInfo.Compare(entry2.ArtifactInfo) == 0) {
				matchingEntries = append(matchingEntries, entry1)
				indexMap1[index1] = true
				indexMap2[index2] = true
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			return err
		}

		if len(matchingEntries) > 0 {
			fmt.Println()
			fmt.Printf("Unchanged (%d)\n", len(matchingEntries))
			tbl := table.New("Filename", "Group ID", "Artifact ID", "Version", "Classifier")
			tbl.WithHeaderFormatter(headerFmt)
			for _, entry := range matchingEntries {
				tbl.AddRow(
					entry.Filename,
					entry.ArtifactInfo.GroupId,
					entry.ArtifactInfo.ArtifactId,
					entry.ArtifactInfo.Version,
					entry.ArtifactInfo.Classifier,
				)
			}
			tbl.Print()
		}
	}

	// check for dependency change
	{
		type DependencyChange struct {
			Entry1 *ArtifactEntry
			Entry2 *ArtifactEntry
		}

		var dependencyChanges []*DependencyChange
		err = processor(func(index1 int, entry1 *ArtifactEntry, index2 int, entry2 *ArtifactEntry) (bool, error) {
			if (entry1.ArtifactInfo.GroupId == entry2.ArtifactInfo.GroupId) && (entry1.ArtifactInfo.ArtifactId == entry2.ArtifactInfo.ArtifactId) {
				dependencyChanges = append(dependencyChanges, &DependencyChange{
					Entry1: entry1,
					Entry2: entry2,
				})
				indexMap1[index1] = true
				indexMap2[index2] = true
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			return err
		}

		if len(dependencyChanges) > 0 {
			fmt.Println()
			fmt.Printf("Changed (%d)\n", len(dependencyChanges))
			tbl := table.New("Group ID", "Artifact ID", "Filename 1", "Version 1", "Classifier 1", "Size 1", "Filename 2", "Version 2", "Classifier 2", "Size 2")
			tbl.WithHeaderFormatter(headerFmt)
			for _, change := range dependencyChanges {
				tbl.AddRow(
					change.Entry1.ArtifactInfo.GroupId, change.Entry1.ArtifactInfo.ArtifactId,
					change.Entry1.Filename, change.Entry1.ArtifactInfo.Version, change.Entry1.ArtifactInfo.Classifier, change.Entry1.Size,
					change.Entry2.Filename, change.Entry2.ArtifactInfo.Version, change.Entry2.ArtifactInfo.Classifier, change.Entry2.Size,
				)
			}
			tbl.Print()
		}

		// check for remaining dependencies in first
		{
			var entries []*ArtifactEntry
			for i, entry := range entries1 {
				processed := indexMap1[i]
				if processed {
					continue
				}
				entries = append(entries, entry)
				indexMap1[i] = true
			}

			if len(entries) > 0 {
				fmt.Println()
				fmt.Printf("Only in first (%d)\n", len(entries))
				tbl := table.New("Filename", "Group ID", "Artifact ID", "Version", "Classifier")
				tbl.WithHeaderFormatter(headerFmt)
				for _, entry := range entries {
					tbl.AddRow(
						entry.Filename,
						entry.ArtifactInfo.GroupId,
						entry.ArtifactInfo.ArtifactId,
						entry.ArtifactInfo.Version,
						entry.ArtifactInfo.Classifier,
					)
				}
				tbl.Print()
			}
		}

		// check for remaining dependencies in second
		{
			var entries []*ArtifactEntry
			for i, entry := range entries2 {
				processed := indexMap2[i]
				if processed {
					continue
				}
				entries = append(entries, entry)
				indexMap2[i] = true
			}

			if len(entries) > 0 {
				fmt.Println()
				fmt.Printf("Only in second (%d)\n", len(entries))
				tbl := table.New("Filename", "Group ID", "Artifact ID", "Version", "Classifier")
				tbl.WithHeaderFormatter(headerFmt)
				for _, entry := range entries {
					tbl.AddRow(
						entry.Filename,
						entry.ArtifactInfo.GroupId,
						entry.ArtifactInfo.ArtifactId,
						entry.ArtifactInfo.Version,
						entry.ArtifactInfo.Classifier,
					)
				}
				tbl.Print()
			}
		}
	}

	return nil
}

func loadEntries(path string) ([]*ArtifactEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	var entries []*ArtifactEntry
	jsonDecoder := json.NewDecoder(f)
	err = jsonDecoder.Decode(&entries)
	if err != nil {
		return nil, err
	}
	return entries, nil
}
