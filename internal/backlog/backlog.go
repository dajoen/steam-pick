package backlog

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Item struct {
	ID     string
	Title  string
	Status string // "todo", "done"
	Line   int
}

type Section struct {
	Name  string
	Items []Item
	Line  int
}

type Backlog struct {
	Sections []Section
}

var (
	sectionHeaderRegex = regexp.MustCompile(`^##\s+(Now|Next|Later|Done)`)
	itemRegex          = regexp.MustCompile(`^-\s+\[([ x])\]\s+(SP-\d+)\s+(.*)$`)
)

func Parse(path string) (*Backlog, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	backlog := &Backlog{}
	var currentSection *Section
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if matches := sectionHeaderRegex.FindStringSubmatch(line); len(matches) > 1 {
			sectionName := matches[1]
			backlog.Sections = append(backlog.Sections, Section{
				Name: sectionName,
				Line: lineNum,
			})
			currentSection = &backlog.Sections[len(backlog.Sections)-1]
			continue
		}

		if matches := itemRegex.FindStringSubmatch(line); len(matches) > 3 {
			if currentSection == nil {
				return nil, fmt.Errorf("line %d: item found outside of a section", lineNum)
			}
			status := "todo"
			if matches[1] == "x" {
				status = "done"
			}
			currentSection.Items = append(currentSection.Items, Item{
				ID:     matches[2],
				Title:  matches[3],
				Status: status,
				Line:   lineNum,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return backlog, nil
}
