package metrics_test

import (
	"testing"

	"github.com/tokamak93/review-loop/internal/metrics"
	"github.com/tokamak93/review-loop/internal/outcome"
)

type spec struct {
	pr         int
	state      outcome.State
	category   outcome.Category
	confidence string
	learning   string
}

func build(specs ...spec) []outcome.Record {
	records := make([]outcome.Record, 0, len(specs))
	for i, s := range specs {
		if s.learning == "" {
			s.learning = "on"
		}
		records = append(records, outcome.Record{
			Repo: "acme/api", PR: s.pr, CommentID: int64(i + 1),
			Outcome: s.state, Category: s.category, Confidence: s.confidence,
			Learning: s.learning, Path: "internal/ledger/expiry.go",
		})
	}
	return records
}

func TestPrecisionCountsIgnoredAndExcludesVoid(t *testing.T) {
	report := metrics.Compute(build(
		spec{pr: 1, state: outcome.Accepted},
		spec{pr: 1, state: outcome.Rejected},
		spec{pr: 2, state: outcome.Ignored},
		spec{pr: 2, state: outcome.Void},
	))

	if got := report.Overall.Counted(); got != 3 {
		t.Errorf("Counted() = %d, want 3 (void excluded)", got)
	}
	assertRatio(t, "precision", report.Overall.Precision, 1.0/3.0)
	assertRatio(t, "engaged precision", report.Overall.EngagedPrecision, 0.5)
	if report.PRs != 2 {
		t.Errorf("PRs = %d, want 2", report.PRs)
	}
}

func TestEmptySummaryReportsNoNumberRatherThanZero(t *testing.T) {
	var s metrics.Summary
	if _, ok := s.Precision(); ok {
		t.Error("Precision() reported a value with nothing counted")
	}
	if _, ok := s.EngagedPrecision(); ok {
		t.Error("EngagedPrecision() reported a value with nothing engaged")
	}
}

func TestReportIsProvisionalBelowTheFloors(t *testing.T) {
	report := metrics.Compute(build(spec{pr: 1, state: outcome.Accepted}))
	if !report.Provisional {
		t.Fatal("a single finding was not marked provisional")
	}
	if report.Shortfall == "" {
		t.Error("Shortfall is empty; the report must say what is missing")
	}
}

func TestHoldoutIsNotComparableUntilBothSidesClear(t *testing.T) {
	specs := make([]spec, 0, metrics.MinFindings+5)
	for i := range metrics.MinFindings {
		specs = append(specs, spec{pr: i, state: outcome.Accepted})
	}
	for i := range 5 {
		specs = append(specs, spec{pr: 500 + i, state: outcome.Accepted, learning: "off"})
	}
	report := metrics.Compute(build(specs...))

	if report.Holdout.Holdout.Counted() != 5 {
		t.Errorf("holdout counted %d, want 5", report.Holdout.Holdout.Counted())
	}
	if report.Holdout.Comparable {
		t.Error("Comparable = true with 5 holdout findings; the floors were not applied")
	}
}

// With the posting gate at high, no medium-confidence finding is ever posted,
// so the calibration check over confidence has nothing to compare. The report
// must say that rather than print one row that looks like a result.
func TestSingleValuedDimensionIsDegenerate(t *testing.T) {
	report := metrics.Compute(build(
		spec{pr: 1, state: outcome.Accepted, confidence: "high", category: "correctness"},
		spec{pr: 2, state: outcome.Rejected, confidence: "high", category: "security"},
	))

	if !dimension(t, report, "confidence").Degenerate {
		t.Error("confidence with one value was not marked degenerate")
	}
	if dimension(t, report, "category").Degenerate {
		t.Error("category with two values was marked degenerate")
	}
}

func TestOrdinalDimensionsSortAsAScale(t *testing.T) {
	report := metrics.Compute(build(
		spec{pr: 1, state: outcome.Accepted, confidence: "low"},
		spec{pr: 2, state: outcome.Accepted, confidence: "high"},
		spec{pr: 3, state: outcome.Accepted, confidence: "medium"},
	))

	want := []string{"high", "medium", "low"}
	segments := dimension(t, report, "confidence").Segments
	for i, seg := range segments {
		if seg.Value != want[i] {
			t.Fatalf("segment %d = %q, want %q — calibration is unreadable out of order", i, seg.Value, want[i])
		}
	}
}

func dimension(t *testing.T, report metrics.Report, name string) metrics.Dimension {
	t.Helper()
	for _, d := range report.Dimensions {
		if d.Name == name {
			return d
		}
	}
	t.Fatalf("no %q dimension in the report", name)
	return metrics.Dimension{}
}

func assertRatio(t *testing.T, name string, f func() (float64, bool), want float64) {
	t.Helper()
	got, ok := f()
	if !ok {
		t.Fatalf("%s reported no value", name)
	}
	if diff := got - want; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}
