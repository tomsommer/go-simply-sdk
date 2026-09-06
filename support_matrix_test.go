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
		Operations []struct{ ID, Method, Path, Status string } `json:"operations"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, op := range m.Operations {
		if op.ID == "" || op.Method == "" || op.Path == "" || (op.Status != "implemented" && op.Status != "unimplemented") {
			t.Fatalf("invalid operation: %+v", op)
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
}
