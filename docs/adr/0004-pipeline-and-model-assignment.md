# 0004 — A staged pipeline, and which model does what

**Status**: accepted

## Context

A single agent given the whole review brief does every part of it at the same cost and with the same attention. Judging whether a design will hurt in six months and counting how many unrelated concerns are in a diff are not comparable problems, and paying Opus rates for the second one is waste. Handing the first one to Haiku is worse than waste.

There is also a context problem. One agent holding the diff, the standards documents, the test files and the security questions at once does each of them slightly worse, and any documentation set large enough to be useful will not fit alongside everything else.

## Decision

Three passes.

**Triage** — Haiku. Establishes what the change claims to do, what it touches, which specialists are worth running, and which documents are worth loading. Reads indexes and headings, never document bodies. Returns at most five document paths.

**Specialists in parallel**, each receiving the triage brief rather than re-deriving it:

| Agent | Model | Why that model |
|---|---|---|
| pr-defects | Sonnet | Locating bugs is pattern recognition against a known catalogue |
| pr-patterns | Opus | The one genuinely hard judgement here: will this design hurt later |
| pr-tests | Sonnet | Comparative reading — what the tests assert against what the code does |
| pr-security | Sonnet | Boundary checks against a known catalogue, with sibling code to compare against |
| pr-reuse | Sonnet | Search-and-compare against code that already exists |
| pr-shape | Haiku | Counting and comparing a diff to a description |

**Aggregation** — Opus, in the orchestrating skill. Merges, deduplicates, applies the learned rules, ranks, caps, posts.

Opus for the orchestrator is the one model choice here that looks like plumbing and is not. Most of what the orchestrator does is clerical — dispatch, ranking, `gh api` calls — but two of its steps are the highest-consequence judgements in the pipeline. Deciding that a finding makes the same point a human already made, on substance rather than wording, is the step that stops the reviewer repeating itself and being switched off. Deciding what is worth the team's attention at all is the step precision is actually won in. A cheaper orchestrator would do the clerical nine tenths for less and get the two that matter wrong, and there is no way to buy only the cheap part.

If the numbers ever show suppression and ranking are not where the value sits, this is the first cost to cut. It is exposed as the `model` input on the reusable workflow so that experiment does not require editing the skill.

ADR 0005 later moved re-verification out of this pass into its own stage, and added the effort dial that decides which specialists run and how large the cap is. ADR 0006 added the holdout, which the orchestrator applies by not reading the learned rules at all on held-out pull requests. ADR 0007 added the confidence gate, applied here before ranking.

The suppression step moved earlier still: reading the pull request's existing conversation is now the first pass, before triage, because a review that has nothing left to say should cost one set of API calls rather than a full fan-out. On a third or fourth push that is the common case.

Documentation loads lazily: nothing is read until triage names it.

## Alternatives rejected

**One agent, one pass.** Simpler and cheaper on input tokens, since the diff is sent once rather than five times. Rejected because the review's whole value is precision, and one agent balancing five concerns at once produced the failure this project exists to fix — plausible findings, thinly checked.

**Everything on Opus.** Removes the model-assignment question. Rejected as several times the cost for gains concentrated in one of the five specialists.

**Everything on Sonnet or Haiku.** Cheapest. Rejected because the architecture pass is where a finding a human reviewer could not easily make gets produced, and that is the one worth paying for.

## Consequences

**Input tokens rise.** Each specialist reads the diff. Fan-out trades input tokens for a smaller, sharper prompt per agent and a fresh context each. Whether that trade pays is measurable, and the `agent=` field in the comment marker exists so the harvest step can segment precision by agent and settle it.

**Aggregation becomes the load-bearing pass.** Five agents produce more raw findings than one, and the cap is unchanged at eight. Most of what specialists produce is meant to be discarded. If aggregation is lazy the review gets worse than the single-agent version, not better — the re-verification step for high-severity findings is not optional.

**Latency rises.** Triage is serial ahead of everything. Acceptable: the review is not blocking a merge, and the static-check gate ahead of it already costs more.

**Model assignment is a hypothesis.** It is stated as reasoning, not measured. The per-agent precision breakdown is what will confirm or refute it, and a specialist that never produces an accepted finding should be deleted rather than downgraded.

**Skipping specialists is a real saving and a real risk.** Triage deciding a documentation-only change needs no concurrency review is right. Triage wrongly marking a subsystem out of scope produces a silent miss nothing measures — the same blind spot as a suppression rule in `.review/learned-rules.md`, and it belongs in the same section of the README.
