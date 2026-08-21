# 0001 — What context the reviewer sees

**Status**: superseded by 0003

## Context

A diff is not enough to judge most defects, and a repository is too much. The reviewer needs a rule for what to include around each hunk that is cheap enough to run on every pull request and stable enough that the same pull request produces the same review.

## Decision

For each hunk: the file's leading declarations, the full enclosing symbol the hunk sits in, and roughly forty lines either side clamped to that symbol. Overlapping ranges within a file are merged. Nothing outside the files the diff touches.

The enclosing symbol is found by a language-agnostic indentation and brace heuristic, not by parsing. This is wrong at the edges — it will over-select in deeply nested code and under-select in languages with unusual block syntax — and it is one file to replace with a real parser if the metric shows it costing findings.

## Alternatives rejected

**Diff only.** The honest floor, and cheaper still. Rejected as a shipping default because a reviewer that cannot see the function it is commenting inside will report defects that the surrounding lines already handle, and false findings are the expensive kind. Kept as a configuration for the baseline comparison.

**Agentic — give the model read and grep tools and let it pull what it needs.** Highest ceiling: it is the only option that can find a defect whose cause is in a file the pull request does not touch. Rejected for v1 on three grounds. Cost is unbounded per pull request, which makes the cost-per-finding figure in `docs/metric.md` meaningless. Reviews stop being reproducible, so the offline replay harness — the thing that makes prompt iteration affordable — no longer replays the same review. And prompt injection through pull request content becomes materially more dangerous once the model can go and read things.

## Consequences

Cross-file defects are out of reach. This is the largest known gap in the reviewer's coverage and belongs in the README rather than being discovered by a user.

Context size is predictable, so the stable prefix caches and cost per pull request can be quoted with a straight face.

Revisit when the metric shows precision plateauing with a visible cluster of misses that are cross-file. Run the agentic variant against the same pull requests then, and let the numbers settle it rather than repeating this argument.
