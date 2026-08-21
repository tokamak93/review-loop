package outcome_test

import (
	"strings"
	"testing"

	"github.com/tokamak93/review-loop/internal/outcome"
)

const valid = `{"finding_id":"pr-defects-1","repo":"acme/api","pr":41,"comment_id":900,` +
	`"comment_url":"https://example.invalid/900","kind":"inline","path":"internal/ledger/expiry.go",` +
	`"line":118,"category":"correctness","severity":"medium","confidence":"high","agent":"pr-defects",` +
	`"verdict":"confirmed","effort":"medium","confidence_gate":"high","learning":"on",` +
	`"skill_sha":"4f21ac9","rules_sha":"0000000","posted_at":"2026-08-01T10:00:00Z",` +
	`"pr_closed_at":"2026-08-03T09:00:00Z","pr_merged":true,"outcome":"accepted",` +
	`"outcome_reason":"lines changed in a1b2c3 after the comment","harvested_at":"2026-08-10T06:00:00Z"}`

func TestLoadAcceptsAValidRecord(t *testing.T) {
	records, err := outcome.Load(strings.NewReader(valid + "\n"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
	if got := records[0].Outcome; got != outcome.Accepted {
		t.Errorf("outcome = %q, want accepted", got)
	}
	if records[0].InHoldout() {
		t.Error("InHoldout() = true for learning=on")
	}
}

// The harvester is required to be idempotent. A repeated comment means it is
// not, and collapsing the duplicate quietly would leave the counts wrong in a
// way nobody would notice.
func TestLoadRejectsADuplicateComment(t *testing.T) {
	_, err := outcome.Load(strings.NewReader(valid + "\n" + valid + "\n"))
	if err == nil {
		t.Fatal("Load accepted a duplicated comment id")
	}
	if !strings.Contains(err.Error(), "already recorded on line 1") {
		t.Errorf("error = %q, want it to name the first line", err)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	line := strings.Replace(valid, `"kind":"inline"`, `"kind":"inline","surprise":1`, 1)
	if _, err := outcome.Load(strings.NewReader(line)); err == nil {
		t.Fatal("Load accepted an unknown field")
	}
}

func TestValidateRejectsIncompleteRecords(t *testing.T) {
	tests := map[string]struct{ find, replace, want string }{
		"empty category":  {`"category":"correctness"`, `"category":""`, "not one of the defined categories"},
		"missing gate":    {`"confidence_gate":"high"`, `"confidence_gate":""`, "confidence_gate is empty"},
		"bad learning":    {`"learning":"on"`, `"learning":"maybe"`, "learning is"},
		"bad outcome":     {`"outcome":"accepted"`, `"outcome":"pending"`, "outcome is"},
		"void, no reason": {`"outcome":"accepted","outcome_reason":"lines changed in a1b2c3 after the comment"`, `"outcome":"void","outcome_reason":""`, "void outcome with no reason"},
		"no pr":           {`"pr":41`, `"pr":0`, "pr is 0"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			line := strings.Replace(valid, tc.find, tc.replace, 1)
			if line == valid {
				t.Fatalf("test fixture did not substitute %q", tc.find)
			}
			_, err := outcome.Load(strings.NewReader(line))
			if err == nil {
				t.Fatalf("Load accepted %s", name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to contain %q", err, tc.want)
			}
		})
	}
}

// Observed in the first real run: pr-tests invented `cannot-fail` for two
// findings. With Category as a bare string this passed validation and would
// have split the test-gap bucket in two, halving both counts in exactly the
// figure a learned rule is proposed from.
func TestLoadRejectsAnInventedCategory(t *testing.T) {
	line := strings.Replace(valid, `"category":"correctness"`, `"category":"cannot-fail"`, 1)
	_, err := outcome.Load(strings.NewReader(line))
	if err == nil {
		t.Fatal("Load accepted an invented category")
	}
	if !strings.Contains(err.Error(), "cannot-fail") {
		t.Errorf("error = %q, want it to name the offending value", err)
	}
}

func TestEveryCategoryInTheClosedListValidates(t *testing.T) {
	if got := len(outcome.Categories()); got != 14 {
		t.Errorf("Categories() has %d entries, want 14 — the skill's contract and this list have drifted", got)
	}
	for _, c := range outcome.Categories() {
		if !c.Valid() {
			t.Errorf("Categories() returned %q but Valid() rejects it", c)
		}
	}
}

func TestPathPrefix(t *testing.T) {
	tests := map[string]string{
		"internal/ledger/expiry.go": "internal/ledger",
		"cmd/api/main.go":           "cmd/api",
		"internal/expiry.go":        "internal",
		"main.go":                   "",
		"":                          "",
	}
	for path, want := range tests {
		t.Run(path, func(t *testing.T) {
			if got := (outcome.Record{Path: path}).PathPrefix(); got != want {
				t.Errorf("PathPrefix(%q) = %q, want %q", path, got, want)
			}
		})
	}
}
