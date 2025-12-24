package backlog

import (
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	content := `# Backlog

## Now

- [ ] SP-001 Task 1

## Done

- [x] SP-002 Task 2
`
	tmpfile, err := os.CreateTemp("", "backlog")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	bl, err := Parse(tmpfile.Name())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(bl.Sections) != 2 {
		t.Errorf("Expected 2 sections, got %d", len(bl.Sections))
	}

	if bl.Sections[0].Name != "Now" {
		t.Errorf("Expected section Now, got %s", bl.Sections[0].Name)
	}

	if len(bl.Sections[0].Items) != 1 {
		t.Errorf("Expected 1 item in Now, got %d", len(bl.Sections[0].Items))
	}

	if bl.Sections[0].Items[0].ID != "SP-001" {
		t.Errorf("Expected SP-001, got %s", bl.Sections[0].Items[0].ID)
	}
}
