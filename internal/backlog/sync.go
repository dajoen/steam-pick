package backlog

import (
	"fmt"
	"os"
	"strings"
)

func Sync(backlogPath, changelogPath string, dryRun bool) error {
	bl, err := Parse(backlogPath)
	if err != nil {
		return fmt.Errorf("failed to parse backlog: %w", err)
	}

	cl, err := ParseChangelog(changelogPath)
	if err != nil {
		return fmt.Errorf("failed to parse changelog: %w", err)
	}

	existingIDs := make(map[string]bool)
	for _, entry := range cl.Entries {
		existingIDs[entry.ID] = true
	}

	var newEntries []string
	for _, s := range bl.Sections {
		if s.Name == "Done" {
			for _, item := range s.Items {
				if item.Status == "done" && !existingIDs[item.ID] {
					// Create a changelog entry
					// Format: - SP-XXX Title
					newEntries = append(newEntries, fmt.Sprintf("- %s %s", item.ID, item.Title))
				}
			}
		}
	}

	if len(newEntries) == 0 {
		fmt.Println("No new items to sync.")
		return nil
	}

	if dryRun {
		fmt.Println("Would add the following entries to CHANGELOG.md:")
		for _, entry := range newEntries {
			fmt.Println(entry)
		}
		return nil
	}

	// Read existing changelog
	content, err := os.ReadFile(changelogPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	// Find where to insert. We look for "## [Unreleased]" or the first "## " section.
	insertIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "## [Unreleased]") {
			insertIdx = i + 1
			break
		}
	}

	// If [Unreleased] not found, look for first header
	if insertIdx == -1 {
		for i, line := range lines {
			if strings.HasPrefix(line, "## ") {
				insertIdx = i
				break
			}
		}
	}

	// If still not found, append to end (or after header)
	if insertIdx == -1 {
		insertIdx = len(lines)
	} else {
		// Skip empty lines after header
		for insertIdx < len(lines) && strings.TrimSpace(lines[insertIdx]) == "" {
			insertIdx++
		}
	}

	// Insert new entries
	newLines := make([]string, 0, len(lines)+len(newEntries))
	newLines = append(newLines, lines[:insertIdx]...)
	newLines = append(newLines, newEntries...)
	newLines = append(newLines, lines[insertIdx:]...)

	output := strings.Join(newLines, "\n")
	if err := os.WriteFile(changelogPath, []byte(output), 0644); err != nil {
		return err
	}

	fmt.Printf("Added %d entries to CHANGELOG.md\n", len(newEntries))
	return nil
}
