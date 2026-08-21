# CLAUDE.md

Context and working agreement for this repository.

## What this project is

A GitHub Action that reviews pull requests with Claude and gets measurably better at it, because it learns from what humans did with its previous comments.

The loop: review a pull request, post findings, observe what happened to each one — resolved, dismissed, reacted to, silently ignored — and fold that signal into a versioned rules file that the next review reads. Nothing is learned without a human approving the lesson.

## The claim being tested

That the limiting factor in automated code review is precision, not capability. A reviewer that posts fifteen findings of which four matter trains the team to skim past all fifteen, and no improvement to the model fixes that. The loop exists to raise precision on this repository's actual conventions, which no general-purpose reviewer can know.

The claim is falsifiable and must be measured. If precision does not move over a meaningful number of pull requests, the README says so.

## Non-goals

- Not a replacement for human review. It is a filter that runs before one.
- Not a linter. If `golangci-lint` or the type checker catches it, this must never mention it.
- No auto-merge, no auto-fix, no pushing commits.
- No autonomous learning. Every learned rule reaches `main` through a pull request a human approves.
- Not a model evaluation harness. It measures this reviewer on this repository, not models against each other.
- No fine-tuning. What is learned is text in a versioned file.
- Never a required check. A reviewer that can block a merge puts its own false positives on the critical path.

## Working agreement

Measurement comes before learning. The outcome harvester and the precision metric are built first, and run against an unmodified reviewer, or there is no baseline to improve on and no way to tell whether the loop does anything.

A change that does not move the metric does not ship, however sensible it sounds.

Findings are cheap to add and expensive to trust. Prefer a reviewer that says less.

Ask before adding a finding category, before changing what the action is allowed to write to, and before changing any default that decides what gets posted — effort, the confidence gate, the holdout ratio. Those three are the dials the whole metric is measured against, and moving one silently makes every number before and after it incomparable.

## Conventions

The reviewer's behaviour is Markdown, not code. That is what makes it reviewable in a pull request and therefore learnable. See ADR 0003.

- `.claude/skills/` holds the two skills — the review orchestrator and the harvest step — and `CONFIDENCE.md` beside the review skill, which is the one rubric every agent applies when setting the field the gate reads. `.claude/agents/` holds the six specialists and the verifier, each pinning its own model in frontmatter.
- **There is one compiled artifact and it is not the reviewer.** `cmd/metrics` reads the committed outcome history and divides. It never runs in the review job, never sees a diff, never calls a model — which is why `docs/metric.md` can promise figures anyone recomputes. Everything else stays text.
- The record type in `internal/outcome/outcome.go` **is** the outcome schema. Do not write a second copy of it in prose; point at it.
- Workflows stay thin and reusable. A consuming service holds a stub of about a dozen lines; everything that decides behaviour lives here. If a workflow starts growing logic, that logic belongs in a skill. ADR 0008.
- The skills ship as a plugin from this repository, from the same files `.claude/` uses locally. One copy. A second copy under `plugins/` would drift, and drift in the reviewer's text is what the metric exists to catch.
- Static checks — build, vet, a strict `golangci-lint`, tests — run as a gating job before Claude is invoked, and are a *separate* reusable workflow so a non-Go service can substitute its own. Nothing statically decidable is ever paid for in tokens, and the skill may state that as fact because the tools demonstrably ran.
- Specialists never post. They return findings in the contract in the review skill, the verifier rules on them and sets the confidence that decides, and the orchestrator alone writes to the pull request.
- One comment per finding, always, carrying the marker. Every marker value is lower case. The loop reads each comment's resolution state and reactions individually, and a comment carrying several findings destroys that signal.
- Every marker carries the reviewer version, the rules revision, the gate and whether learning was on. Without those, a change in precision cannot be attributed to the change in text that caused it. ADR 0006.
- Outcomes are data and go straight to the default branch, append-only. Learned rules are behaviour and reach it only through a pull request a human approves. The distinction is deliberate; ADR 0008 records it.
- The learned rules file is human-readable Markdown. If a maintainer cannot read the lesson, the loop is not auditable.
- Least privilege: `contents: read` and `pull-requests: write` for review, and `Edit`, `Write`, `WebSearch` and `WebFetch` explicitly disallowed. A reviewer that can write code needs write access everywhere it runs.
- Prompts and skill text are the source code of this project. Change them with the same care, and never in a way the metric cannot detect.

## Where the honest weaknesses are recorded

`docs/metric.md` threats 4, 5 and 6, and `docs/failure-modes.md`. They are the same mistake applied to three fields: a learned rule, the confidence gate and a false rejection each remove findings from the record, so precision improves whether or not the reviewer did.

The holdout bounds the first. Nothing bounds the third. If the qualifier "unverified" is ever quietly dropped from a reported improvement, this project has stopped being an argument and started being a demo.
