package backlog

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

type ChangelogEntry struct {
	ID    string
	Title string
	Line  int
}

type Changelog struct {
	Entries []ChangelogEntry
}

var changelogEntryRegex = regexp.MustCompile(`(SP-\d+)`)

func ParseChangelog(path string) (*Changelog, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	changelog := &Changelog{}
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Look for SP-XXX in the line
		matches := changelogEntryRegex.FindAllString(line, -1)
		for _, match := range matches {
			changelog.Entries = append(changelog.Entries, ChangelogEntry{
				ID:    match,
				Title: strings.TrimSpace(line),
				Line:  lineNum,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return changelog, nil
}
