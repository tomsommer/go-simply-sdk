package simply

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSupportMatrixExplicit(t *testing.T) {
	b, err := os.ReadFile("testdata/support-matrix.json")
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Operations []struct {
			ID, Method, Path, Status, Note string
			Deprecated                     bool
		} `json:"operations"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, op := range m.Operations {
		if op.ID == "" || op.Method == "" || op.Path == "" || (op.Status != "implemented" && op.Status != "unimplemented" && op.Status != "not_planned") {
			t.Fatalf("invalid operation: %+v", op)
		}
		if op.Status == "not_planned" && (!op.Deprecated || op.Note == "") {
			t.Fatalf("not planned operation must be deprecated with a note: %+v", op)
		}
		key := op.Method + " " + op.Path
		if seen[key] {
			t.Fatalf("duplicate operation: %s", key)
		}
		seen[key] = true
	}
	if len(m.Operations) != 27 {
		t.Fatalf("expected 27 operations, got %d", len(m.Operations))
	}
	notPlanned := map[string]bool{}
	for _, op := range m.Operations {
		if op.Status == "not_planned" {
			notPlanned[op.ID] = true
		}
	}
	if len(notPlanned) != 2 || !notPlanned["ddnsHelper"] || !notPlanned["dyndns"] {
		t.Fatalf("expected only ddnsHelper and dyndns to be not planned, got %v", notPlanned)
	}
}
