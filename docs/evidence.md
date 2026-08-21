# Evidence

**No precision figure exists.** One review has been run. Nobody has acted on its findings, so no outcome has been observed, so the metric this repository is built around is still undefined.

What follows is a worked example: what the pipeline produced on one real change, what it cost, and what it revealed about itself. It is not evidence that the reviewer is any good. Keeping those two things apart is the whole discipline of this document.

## The run

| | |
|---|---|
| Date | 2026-08-21 |
| Target | [`tokamak93/subito-challenge`](https://github.com/tokamak93/subito-challenge) at [`487d608`](https://github.com/tokamak93/subito-challenge/commit/487d608) — the initial commit of a Go purchase-cart service: hexagonal layering, nine ADRs, three test files |
| Size | 3,233 insertions across 35 files; ~2,379 reviewable after excluding generated Swagger and `go.sum` |
| Effort | `medium`, gate `high`, learning off (no rules file exists) |
| Cost | 397,932 tokens across 8 agents, ≈$1.70 — `docs/cost.md` |

**Why this target.** It needed to be real Go with real structure — layers, interfaces, recorded decisions, a test suite — because half the pipeline is only exercised by a repository that has documents to cite and siblings to compare against. It is an old interview challenge, not a showcase, and it is linked at a pinned commit so it stays the thing that was reviewed regardless of what happens to it later. The subject here is the reviewer; the code is the fixture.

**How it was run, precisely.** Not in CI. The pipeline was executed manually in a local Claude Code session: each agent spawned as a general-purpose subagent that read its own definition file from `.claude/agents/` and followed it, with the model pinned to match its frontmatter. Same instructions, same models, same order, same confidence rubric.

It differs from a real run in three ways that matter:

- **Nothing was posted.** No comments exist on any pull request. The posting step, the marker, and the harvester were not exercised at all.
- **There was no conversation to read.** Pass 1 had nothing to suppress against, so the pipeline ran at full width. A real second or third review is much cheaper and the cost figure does not reflect that.
- **The orchestrator was a human-directed session**, not the skill running unattended. Where the skill was ambiguous, a person resolved it — and each time that happened is recorded below as a defect, because in CI nobody would have been there.

## What it produced

Six specialists returned 11 findings. The verifier confirmed all 11. The gate dropped none. The cap posted 6.

**Would have been posted.** The target repository is public, so every one of these is checkable rather than merely asserted — the location column links to the line.

| | Finding | Severity | Where |
|---|---|---|---|
| 1 | VAT rounded half-up per unit then multiplied by quantity, so the rounding residual scales with quantity instead of being rounded once per line | high | [`order_service.go:84-87`](https://github.com/tokamak93/subito-challenge/blob/487d608/internal/core/service/order_service.go#L84-L87) |
| 2 | Goroutine fan-out driven by an unbounded client-supplied array; ADR 0008 claims batch sizing mitigates this, and dividing by a constant does not bound anything | high | [`order_service.go:129-156`](https://github.com/tokamak93/subito-challenge/blob/487d608/internal/core/service/order_service.go#L129-L156) |
| 3 | `Quantity` accepted with no upper bound and multiplied into persisted totals — `int64` overflow reachable from one request | medium | [`dto.go:11`](https://github.com/tokamak93/subito-challenge/blob/487d608/internal/handler/http/dto.go#L11) |
| 4 | Already-cancelled context makes `fetchProductsInParallel` return `(nil, nil)`, which the caller reports to the client as "product not found" | medium | [`order_service.go:138-174`](https://github.com/tokamak93/subito-challenge/blob/487d608/internal/core/service/order_service.go#L138-L174) |
| 5 | No timeout anywhere on the request path, against ADR 0008's claim that timeouts propagate to all goroutines | medium | [`main.go:81-84`](https://github.com/tokamak93/subito-challenge/blob/487d608/cmd/server/main.go#L81-L84) |
| 6 | `Product.VATRate` is dead but reachable from the operator-facing catalog file via tagless JSON matching, so a per-SKU rate would be silently discarded | medium | [`model.go:26`](https://github.com/tokamak93/subito-challenge/blob/487d608/internal/core/domain/model.go#L26) |

Two were checked by hand afterwards and both hold. Finding 1 diverges as claimed but **not in a fixed direction** — `+33` at price 127 rate 21% quantity 100, `-21` at price 101 same rate and quantity — and is exactly zero for both products in the shipped `configs/products.json`, so it is dormant on the seed data and live on any real catalog. Finding 4 was reproduced by a throwaway test: a pre-cancelled context returns `product not found` with the catalog gateway never called at all.

That the direction is not fixed matters, because the specialist claimed "systematically over-collecting" and the verifier struck that word out on its own — see the confirmation-rate section below.

**Dropped at the cap (5):** three test-quality findings, an error-leak, and the missing commit description.

**Returned nothing:** `pr-reuse`, correctly — its primary finding is "you reimplemented something this repository already has", and the repository did not exist before this commit.

## What this does not show

**That all six findings are correct.** They were confirmed by `pr-verify`, which is part of the system under evaluation, not an independent judge of it. Two were later checked by hand and hold; the other four have not been, and a finding holding up is not the same as a finding being *worth posting* — finding 6 may well be a fair reading of a field somebody meant to remove.

That distinction is the whole reason the metric is defined the way it is. "Is this claim true" is answerable by reading the code. "Was this comment worth an engineer's attention" is only answerable by watching what an engineer did with it, which is why `accepted` requires a human to have acted and why none of this counts as precision.

**Anything about precision.** Precision is `accepted / (accepted + rejected + ignored)`, and all three require a human to have seen a comment. Zero comments were posted. `.review/outcomes.jsonl` is empty and `go run ./cmd/metrics` on it returns "no records in range" — correctly.

**Anything about the loop.** No rule has been learned, no holdout slice exists, and the second half of this repository — harvest, outcomes, learned rules — has never run.

## What the run revealed about the reviewer

The most useful output was not the findings. It was three defects in the design, each found by hitting it:

1. **`category` was not a closed type.** `pr-tests` returned `cannot-fail` for two findings that were plainly `test-gap`. As a bare string it passed validation, and would have reached `outcomes.jsonl` and split the `test-gap` bucket in two — halving both counts in exactly the figure a learned rule is proposed from. Now a closed Go type beside `State`.
2. **Finding ids leaked their author to the verifier.** Handing `pr-patterns-3` to a pass whose entire value is not knowing which agent produced what. Fixed by relabelling to `F1…Fn` before Pass 4.
3. **Ranking had no tie-break below severity.** Eight findings tied at `medium` with six slots, and the skill said nothing about which survive. Now breaks on category spread — under which this run would have kept a test finding rather than dropping all three.

4. **The tool restrictions were not enforced, and it showed immediately.** The specialists pin `tools: Read, Grep, Glob, Bash`, and `review.yml` additionally sets `--disallowedTools "Edit,Write,WebSearch,WebFetch"`. Neither applied here: the agents were spawned as general-purpose subagents carrying the full toolset. Within one run, one of them silently rewrote a loop bound in the code it was reviewing — a behaviour-preserving refactor nobody asked for, in a repository it had read access to and nothing more.

   The instructions say "post nothing yourself" in every agent file. Instruction did not hold; the tool allowlist would have. This is the strongest argument in this document for the design being right, produced by accidentally removing the guard, and it is the reason the fourth item on the list below is not optional.

All four were invisible to inspection and obvious on contact. That is the argument for running a thing before writing its README, and this repository failed to make it in the right order.

## The confirmation rate is a yellow flag

**11 of 11 CONFIRMED, zero REJECTED.**

ADR 0005 records that a high rejection rate is information. A zero rejection rate is information too, and from one run the benign reading (the specialists were right) is indistinguishable from the worrying one (the verifier rubber-stamps).

Two observations argue against pure rubber-stamping. It **narrowed** finding 1's claim, rejecting "systematically over-collecting" on the grounds that round-half-up can err in either direction depending on the fractional remainder — keeping the defect, fixing the overreach. And it **verified empirically**, writing a throwaway Go program to confirm that `encoding/json` populates a tagless `VATRate` field case-insensitively, rather than asserting it from memory.

It also **raised three findings from medium to high confidence**, which is permitted — the verifier's number replaces the specialist's, only severity is one-way — but means three of the six posted findings reached the author on the verifier's judgement rather than the specialist's. If the verifier is generous, the gate is not doing what `docs/adr/0007` claims.

This is exactly the number that needs many runs and cannot be settled by one. It is recorded here so that a later low rejection rate is not mistaken for a new problem, or a good one.

## What would make this document real

In order, and none of it is done:

1. Post the findings on an actual pull request, so outcomes become observable
2. Have a human accept or reject each one, on the record
3. Repeat until 80 non-void findings across 20 pull requests — the floors in `docs/metric.md`
4. Report precision from `go run ./cmd/metrics`, with the range beside it
5. Only then introduce a learned rule, and only then compare against the holdout

Step 1 has not happened, and on a private showcase repository with no API key configured in CI, it is not scheduled. That is a deliberate choice about this repository's purpose, not an oversight — but it means everything above stays a worked example, and this document must not start describing it as anything else.
