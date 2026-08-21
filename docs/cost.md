# What a review costs

Measured, once, on 2026-08-21. One run is not a cost model — it is one data point with a large error bar, and the sections below say where the bar is widest.

## The measurement

A single review at `medium` effort over a 3,233-line, 35-file change: the initial commit of a Go purchase-cart service (hexagonal layering, nine ADRs, three test files). Deliberately at the large end — most pull requests are a tenth of this.

Eight agents, 397,932 tokens.

| Agent | Model | Tokens | Est. cost | Findings |
|---|---|---|---|---|
| pr-triage | Haiku | 29,825 | $0.04 | — |
| pr-defects | Sonnet | 69,606 | $0.20 | 1 |
| pr-patterns | Opus | 47,927 | **$0.34** | 4 |
| pr-tests | Sonnet | 62,343 | $0.18 | 3 |
| pr-security | Sonnet | 55,244 | $0.16 | 2 |
| pr-reuse | Sonnet | 43,423 | $0.12 | **0** |
| pr-shape | Haiku | 25,847 | $0.04 | 1 |
| pr-verify | Sonnet | 63,717 | $0.18 | — |
| orchestrator | Opus | not instrumented | ~$0.40–0.50 | — |

**≈ $1.70 per review**, of which roughly a quarter is orchestration that was never measured.

## Why these numbers are soft

Three reasons, in descending order of how much they could move the total.

**The input/output split is assumed, not measured.** Only combined token counts were captured. Output costs five times input on every model, so the 90/10 split assumed here is the single largest source of error. If these agents actually run at 80/20, the total is roughly 40% higher.

**Orchestration was not instrumented at all.** The figure above is an estimate of the orchestrator's own passes — reading the conversation, merging and deduplicating, relabelling, ranking, posting. It is the second-largest line item and the least known.

**Prompt caching is not reflected.** Every run re-sends the skill text, the agent definitions and the confidence rubric. In CI under `claude-code-action` those are a stable prefix across runs and should cache; nothing here measures whether they do. Real costs are likely lower than this table.

## Per tier

Extrapolated from the same run by removing the agents each tier does not run, applying the `-lite` model downgrades, and scoping verification to what the cap could post.

| | This change (3,233 lines) | Typical PR (~300 lines) |
|---|---|---|
| `low` | ~$0.63 | ~$0.20 |
| `medium` | ~$1.70 | ~$0.55 |
| `high` | ~$2.00 | ~$0.65 |

`low` lands about 63% under `medium` from three compounding cuts: two specialists instead of five, both one model tier down, and verification scoped to the top six findings instead of all eleven.

At a hundred pull requests a month, that is roughly **$20/month at `low`** against **$55 at `medium`** — against one engineer-hour, which costs more than a year of either. The interesting question about this reviewer was never whether it is affordable. It is whether what it posts is worth reading, and `docs/metric.md` is the only thing that answers that.

## Where the money actually goes

From this run, and worth re-checking once there is more than one:

**Diff size dominates.** Every specialist reads the change. Cost scales with what it has to read, not with what it finds — which is why a 3,000-line pull request costs three times a 1,000-line one and produces nowhere near three times the value. The cheapest available saving is a team that opens smaller pull requests, and `pr-shape` exists partly to encourage that.

**Documentation lazy-loading works.** Triage cost $0.04 on a repository with nine ADRs and a 467-line README, because it read the *titles* and returned four paths. Loading that documentation set wholesale into five specialists would have cost more than the entire rest of the review.

**pr-patterns is the most expensive agent and the best value.** It cost $0.34 — a fifth of the run — and produced both high-severity findings and all four that cite an ADR by line. These are the findings a human reviewer most often cannot make, because they require having read a decision record nobody re-reads. It is the last thing to cut.

**pr-reuse cost $0.12 and produced nothing**, which was predictable: its primary finding is "you reimplemented something this repository already has", and the repository did not exist before this commit. Triage marked it applicable anyway. That is a triage miss with a price tag, and the reason `pr-reuse` now runs only at `high`.

## The cheapest saving available

Not a model change. **Not re-reviewing.** The review re-runs on every push and suppresses what has already been said, so the second and third reviews of a pull request cost a fraction of the first — on a fourth push the correct output is often nothing at all, and the run should cost one set of API calls.

That is why reading the pull request conversation is the *first* pass rather than the last. An earlier design ran it after verification, which meant a re-review paid the full price of triage, five specialists and a verifier to discover it had nothing to add.

## What is not in these numbers

- The static gate — build, vet, lint, tests — which runs on GitHub's included minutes and costs nothing per review
- The weekly harvest run
- Failed or truncated runs, which cost tokens and produce no comments
- Cancelled runs, which cannot happen on the review job by design (`cancel-in-progress: false`)

## Reproducing this

There is no cost command, and the numbers above were assembled by hand from agent token counts. That is a gap: `docs/metric.md` insists every precision figure come from a command anyone can run, and cost is held to a weaker standard here than precision is.

The honest fix is to record per-run token usage into `.review/` alongside outcomes, so cost per finding and cost per *accepted* finding become computable. Cost per accepted finding is the number that actually matters and nothing in this repository can produce it yet.
