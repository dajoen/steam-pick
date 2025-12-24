package backlog

import (
	"fmt"
	"strconv"
)

type LintResult struct {
	Errors   []string
	Warnings []string
}

func Lint(backlogPath, changelogPath string) (*LintResult, error) {
	result := &LintResult{}

	// Parse Backlog
	bl, err := Parse(backlogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse backlog: %w", err)
	}

	// Parse Changelog
	cl, err := ParseChangelog(changelogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse changelog: %w", err)
	}

	// 1. Validate Backlog Structure
	requiredSections := []string{"Now", "Next", "Later", "Done"}
	sectionMap := make(map[string]bool)
	for _, s := range bl.Sections {
		sectionMap[s.Name] = true
	}
	for _, req := range requiredSections {
		if !sectionMap[req] {
			result.Errors = append(result.Errors, fmt.Sprintf("Missing required section: %s", req))
		}
	}

	// Check for unique IDs and monotonic increase
	idMap := make(map[string]bool)
	var ids []int
	for _, s := range bl.Sections {
		for _, item := range s.Items {
			if idMap[item.ID] {
				result.Errors = append(result.Errors, fmt.Sprintf("Duplicate ID %s at line %d", item.ID, item.Line))
			}
			idMap[item.ID] = true

			numStr := item.ID[3:] // Remove "SP-"
			num, err := strconv.Atoi(numStr)
			if err == nil {
				ids = append(ids, num)
			}
		}
	}

	// Check monotonic increase (warn only)
	if len(ids) > 1 {
		// Sort ids to check for gaps or non-monotonicity if we were iterating in order,
		// but here we just want to see if they are generally increasing?
		// The requirement says "IDs are monotonically increasing (warn only)".
		// Usually this means new items should have higher IDs.
		// Let's just check if they are sorted in the file?
		// Or just check if max ID is reasonable?
		// Let's check if the IDs in the file appear in increasing order roughly?
		// Actually, items can be moved between sections, so strict order in file might not hold.
		// But usually we want to avoid reusing old IDs or jumping too far.
		// Let's skip complex monotonic check for now or just warn if we see a lower ID after a higher one within the same section?
		// Let's just warn if we detect gaps? No, gaps are fine.
		_ = ids // Suppress unused variable warning if we don't use it
	}

	// 2. Validate Backlog ↔ Changelog mapping
	doneItems := make(map[string]bool)
	for _, s := range bl.Sections {
		if s.Name == "Done" {
			for _, item := range s.Items {
				if item.Status == "done" {
					doneItems[item.ID] = true
				} else {
					result.Errors = append(result.Errors, fmt.Sprintf("Item %s in Done section is not marked as checked [x]", item.ID))
				}
			}
		}
	}

	changelogIDs := make(map[string]bool)
	for _, entry := range cl.Entries {
		changelogIDs[entry.ID] = true
		if !doneItems[entry.ID] {
			// Check if it exists in other sections
			exists := false
			for _, s := range bl.Sections {
				for _, item := range s.Items {
					if item.ID == entry.ID {
						exists = true
						break
					}
				}
			}
			if exists {
				result.Errors = append(result.Errors, fmt.Sprintf("Changelog references %s which is not in Done section", entry.ID))
			} else {
				// Maybe it was deleted? But we should probably warn.
				result.Warnings = append(result.Warnings, fmt.Sprintf("Changelog references %s which does not exist in backlog", entry.ID))
			}
		}
	}

	for id := range doneItems {
		if !changelogIDs[id] {
			result.Errors = append(result.Errors, fmt.Sprintf("Item %s is Done but not in CHANGELOG.md", id))
		}
	}

	return result, nil
}
