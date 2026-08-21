# review-loop

A pull request reviewer that measures its own precision, and learns from what humans actually did with its comments.

**Nothing here has been measured yet.** No run has happened against a real pull request. `docs/evidence.md` is empty on purpose and will stay that way until there is something to put in it. Read that first if you are here to find out whether this works — the rest of this file describes a design, not a result.

## The claim

The limiting factor in automated code review is precision, not capability.

A reviewer that posts fifteen findings of which four matter trains the team to skim past all fifteen. No improvement to the model fixes that, because the problem is not what the reviewer can see — it is what the team has stopped reading. The loop exists to raise precision on one repository's actual conventions, which no general-purpose reviewer can know.

The claim is falsifiable, and `docs/metric.md` fixes the definitions before any number is produced.

## What it is built for

A fleet. Two hundred microservices in different languages, owned by different teams, each with its own CI.

That constraint is the reason the repository is shaped the way it is, and it is not hypothetical — it is the shape that survives contact with a real organisation. A reviewer distributed by copying its instructions into every service is frozen on the day it ships: improving it becomes a two-hundred-pull-request migration, so it never improves, so the loop this whole project is about never turns. The reviewer has to live in one place and be *called* from everywhere else.

So a service holds a stub of about a dozen lines and nothing else. Everything that decides behaviour — the passes, the models, the agents, the confidence gate, the tool allowlist — lives here and changes here, and moving a tag ships the change to the whole fleet at once. Three things follow from that, each with its own cost:

- **The gate is a separate workflow from the review.** Two hundred services do not share one build. A Go service calls `go-checks.yml`; a service in another language substitutes its own job and calls `review.yml` unchanged, because the review does not care what the gate was, only that one passed.
- **What is learned stays local.** The reviewer is central; `.review/learned-rules.md` and the outcome history are not. A rule learned from one team's rejections is evidence about that team's conventions, and pooling rules across a fleet produces a file that is either too bland to be a lesson or wrong in most places it lands. The cost is that a genuinely general lesson has to be promoted by hand into the reviewer itself.
- **One tag move can break everything simultaneously.** That is the price of the first bullet, and `docs/failure-modes.md` says what it costs and how to get out of it.

**This repository is a single instance of that design, not a deployment of it.** It reviews its own pull requests, through the same reusable workflows and the same stub a service would copy — which is the only honest way to show the shape works. Nothing has been run across an actual fleet, and `docs/evidence.md` is empty. The reasoning, including the alternatives rejected, is in `docs/adr/0008-distribution.md`.

## The loop

```
  pull request
       │
       ▼
  ┌──────────┐   build · vet · lint · test        fails ──▶ no review
  │  gate    │   nothing decidable by a tool
  └────┬─────┘   is ever paid for in tokens
       │ passes
       ▼
  ┌──────────────────────────────────────────────────────┐
  │  1  read the conversation   what has already been said │
  │  2  triage        haiku     what changed, which docs   │
  │  3  specialists   parallel  defects · patterns · tests │
  │                             security · reuse · shape   │
  │  4  verify        sonnet    fresh context, no reasoning│
  │  5  aggregate     opus      gate · dedup · rank · post │
  └────┬─────────────────────────────────────────────────┘
       │ one comment per finding, each carrying a marker
       ▼
  humans resolve · reply · react · ignore
       │
       ▼
  ┌──────────┐   weekly, over closed pull requests
  │ harvest  │   → .review/outcomes.jsonl  (append-only)
  └────┬─────┘   → go run ./cmd/metrics
       │
       ▼
  a pull request proposing a rule, with the evidence attached
       │
       ▼
  a human approves the lesson, or does not
```

Nothing is learned without a human approving it. What is learned is text in a versioned file, not weights.

## Install

One stub per repository. Everything that decides behaviour lives here and changes here — which is the point, because a reviewer frozen by a two-hundred-repository migration cannot learn.

```yaml
# .github/workflows/review.yml in the service being reviewed
name: PR review
on:
  pull_request:
    types: [opened, synchronize, reopened, ready_for_review]

permissions:
  contents: read
  pull-requests: write

jobs:
  checks:
    uses: tokamak93/review-loop/.github/workflows/go-checks.yml@v1

  review:
    needs: checks
    if: ${{ !github.event.pull_request.draft }}
    uses: tokamak93/review-loop/.github/workflows/review.yml@v1
    secrets: inherit
```

Set `ANTHROPIC_API_KEY`. Copy `templates/consumer-harvest.yml` too if the repository has enough traffic to clear the sample floors. Not a Go service? Replace the `checks` job with your own build/lint/test job — the review does not care what the gate was, only that one passed.

Full stubs in `templates/`. Distribution reasoning in `docs/adr/0008-distribution.md`.

## Permissions, and why

| | |
|---|---|
| `contents: read` | Read the code under review. It cannot push. |
| `pull-requests: write` | Post comments. Its only write capability. |
| `Edit` `Write` `WebSearch` `WebFetch` **disallowed** | It cannot change code or reach the network |
| `Bash` allowlisted | Specific read-only `gh` and `git` invocations. No general shell. |

The harvest workflow additionally needs `contents: write`, for `.review/outcomes.jsonl` and nothing else — append-only evidence, committed directly because it is data. Learned rules are behaviour and reach the branch only through a pull request a human approves.

The pull request's contents are attacker-controlled and the reviewer reads all of them. What an injection can achieve is bounded by the table above rather than by detection. **Read `docs/failure-modes.md` before installing this on a public repository.**

## What it will not do

- Block a merge. The review is never a required check, and making it one puts its own worst case on the critical path.
- Replace human review. It is a filter that runs before one.
- Report anything a linter reports. The gate ran first, so the skill can state that as fact.
- Auto-merge, auto-fix, or push a commit.
- Learn anything without a human approving it.
- Claim a recall number. It cannot measure one and says so.

## Honest limits

Three of these are the same mistake applied to three different fields, and all three are in `docs/metric.md` and `docs/failure-modes.md` in more detail.

- **A learned rule that suppresses a category raises precision whether or not the rule was any good.** Suppressed findings are never posted, so they can never be shown to have been right. The holdout — every fifth pull request reviewed with the rules switched off — is the only mechanism that catches this, it costs a fifth of the loop's learning, and it says nothing until 80 findings have accumulated on each side.
- **The confidence gate does the same to confidence.** Only findings the reviewer claims to be sure of are posted, so its confidence looks perfectly calibrated by construction. The metrics command prints `degenerate` rather than pretending that is a passing check.
- **A finding the verifier wrongly rejects is invisible everywhere**, including to the holdout. No mitigation.
- **Silence under prompt injection is unmeasurable**, for the same reason recall is.
- **Where the holdout is disabled or has not cleared the floors, any improvement is unverified** and is reported in that word.

## Repository

| | |
|---|---|
| `.claude/skills/pr-review/` | The reviewer. `CONFIDENCE.md` beside it is the rubric every agent applies. |
| `.claude/skills/review-harvest/` | The harvester |
| `.claude/agents/` | Six specialists and the verifier, each pinning its own model |
| `.github/workflows/` | Three reusable workflows, plus this repository's own stubs |
| `templates/` | What a consuming service copies |
| `cmd/metrics` `internal/` | The metric, in Go. Reads committed data, calls nothing. |
| `docs/metric.md` | The definitions, fixed before anything was measured |
| `docs/failure-modes.md` | What goes wrong, including what has no answer |
| `docs/evidence.md` | Empty |
| `docs/adr/` | Why it is shaped this way, including the parts that were wrong |
