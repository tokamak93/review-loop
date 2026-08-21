# 0002 — One call per pull request

**Status**: superseded by 0003

## Context

The reviewer can send the whole diff in one request or fan out one request per file and consolidate. The choice determines what class of defect is reachable, what a review costs, and what happens on a large pull request.

## Decision

One call per pull request, with the assembled context from ADR 0001 for every file in a single request.

Above a token threshold the diff is split by file into as few requests as fit, and the split is recorded on every finding it produced. That metadata is not decoration: it lets the metric compare split reviews against whole ones and shows what the cross-file view is actually worth.

## Alternatives rejected

**One call per file, always.** Scales to any pull request, parallelises, and isolates a bad file so it cannot derail the rest. Rejected because it forecloses the findings that justify this project's existence — a caller and callee changed in the same pull request in incompatible ways is invisible to a per-file reviewer, and it is exactly the defect a linter cannot reach either. It also needs a consolidation pass to deduplicate, and that pass is a second thing that can be wrong in ways nobody notices.

## Consequences

Cross-file reasoning is available on ordinary pull requests, which the target repositories mostly have — the standard is around two hundred lines, and a diff that size fits comfortably.

One cost figure and one cache prefix per review, so the numbers in `docs/metric.md` stay simple.

A single malformed or hostile file can degrade the whole review rather than one finding. Accepted, and the failure is at least loud.

The threshold is configuration and its default is a guess. Record what it was on every run so the first real evidence can correct it.
