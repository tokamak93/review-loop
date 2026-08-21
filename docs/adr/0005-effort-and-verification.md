# 0005 — An effort dial, and verification as its own pass

**Status**: accepted

## Context

Two problems, and one answer to both.

Specialists assess their own confidence from a brief and partial reading. That number then decides what reaches a human, which makes it the least trustworthy input in the pipeline doing the most consequential job. ADR 0004 put a re-check inside aggregation, but a step buried inside another step is a step that gets skipped under pressure.

Separately, different repositories want different things from this reviewer. A team adopting it wants very few, very solid comments. A team that trusts it wants breadth. Serving both by editing the skill file means every consumer maintains a fork.

## Decision

**Verification is its own pass**, run by `pr-verify` on a fresh context, after merge and deduplication and before aggregation. It receives the findings alone — not the brief, not the specialists' reasoning, not which agent produced what — and returns CONFIRMED, PLAUSIBLE or REJECTED per finding. It may lower severity and confidence and sharpen a claim; it may not raise severity and may not add findings.

**Effort is a single input**, `REVIEW_EFFORT`, defaulting to `medium`. It sells **coverage, never rigour**: triage, verification and the confidence gate run at every tier.

| | Specialists | Verified | Cap | Posts |
|---|---|---|---|---|
| `low` | pr-defects-lite, pr-patterns-lite | top 6 | 4 | CONFIRMED only |
| `medium` | pr-defects, pr-patterns, pr-tests, pr-security, pr-shape | top 8 | 6 | CONFIRMED only |
| `high` | + pr-reuse | all | 8 | CONFIRMED and PLAUSIBLE |

*(Amended after the first real run. The original `low` was pr-defects plus pr-shape, which was wrong: pr-shape counts lines and reads the description, and a tier whose most reliable output is "your description is empty" is not a cheap review, it is a bot a team uninstalls. The floor has to be the smallest set that is still a code review — one agent hunting bugs, one judging whether the design will hurt later.)*

Two mechanisms make `low` cheap without making it untrustworthy:

**Specialists run one model tier down**, via `-lite` agent files that pin a cheaper model and delegate to the full-price agent's instructions, so there is one copy of every rule. **The verifier is never downgraded.** A weaker specialist finds less — a coverage reduction, which effort is allowed to sell. A weaker verifier checks worse — a rigour reduction, which it is not.

**Verification is scoped, not skipped.** Findings are ranked by severity before Pass 4 and only as many as the tier could post are verified. Verifying findings the cap was always going to drop is waste that costs nothing to remove; skipping verification entirely would return the gate to the specialists' self-assessment, which is the thing this ADR exists to refuse.

The property that falls out: **cheapness costs recall, not precision.** A `low` review says less. It does not say worse.

Both are recorded in the comment marker, so `docs/metric.md` can segment precision by verdict and by effort.

## Why independence is the whole point

An agent asked to check its own finding re-runs the reasoning that produced it and arrives where it started. Verification only means anything if the verifier can reach a different answer, which requires it to start from the claim and the file rather than from the argument.

Hence the deliberate information starvation. Telling the verifier that pr-patterns on Opus produced a finding tells it how to feel about the finding, which is exactly what it must not know.

## Alternatives rejected

**Verification inside aggregation**, as ADR 0004 had it. Cheaper by one agent call. Rejected because the aggregator has seen every specialist's reasoning and cannot unsee it, and because "also re-check the important ones" is not a step, it is an intention.

**A confidence threshold instead of a verdict.** Post everything above medium confidence. Rejected: it trusts precisely the self-assessed number this ADR exists because we do not trust.

**Per-category or per-agent configuration** rather than one dial. More expressive and it makes the reviewer's behaviour a configuration surface teams tune by feel. One dial keeps the tuning honest — turn it up when the numbers say so.

## Consequences

Cost rises by one Sonnet pass over the findings, which is small: the verifier reads a handful of claims and the files behind them, not the diff.

Latency rises by one serial stage. Acceptable — nothing is blocked on this review.

**The rejection rate becomes a first-class signal.** A specialist whose findings are routinely rejected is not a specialist worth running, and this is the first thing in the pipeline that can say so without waiting weeks for human outcomes. Record the counts every run.

**PLAUSIBLE at `high` is an experiment with a stated exit.** It exists to answer whether unverifiable findings are ever worth an author's attention. The marker records the verdict, so precision can be computed for PLAUSIBLE findings separately. If it comes out low, `high` should stop posting them and the tier should mean broader specialists rather than weaker evidence.

**A false REJECTED is silent.** A correct finding the verifier talks itself out of is never posted and never measured — the same blind spot as a suppression rule, and it belongs beside it in the README's limits. The mitigation is that rejection requires naming the specific guard, caller, test or invariant that makes the finding wrong; "could not reproduce" is not a reason.
