# Spec — v1

## 0. Method note

The interesting part of this project is the loop, and the loop cannot be evaluated without a baseline. Build the measurement before the learning. A version of this that reviews well but cannot demonstrate improvement is a worse deliverable than one that reviews adequately and proves the loop works.

## 1. The reviewer

**Amended.** This section was written for a compiled action that assembled bounded context and made one structured call. ADR 0003 replaced that with a Claude Code skill, and ADRs 0004 to 0008 settled the questions this section left open. What survives of the original requirements:

- Output is structured: id, file, line, category, severity, confidence, one-sentence claim, and a concrete failure scenario. Never free prose parsed afterwards.
- Findings post as inline review comments, one per finding, so that resolution state is observable later. This is not cosmetic — the outcome signal in section 3 depends on it.
- Nothing a linter reports is ever posted.
- A hard ceiling on findings per pull request.

What changed, and where it is recorded:

- **Context selection** — the open question about how much surrounding code the reviewer may read is answered by ADR 0003: it is an agent with file tools, bounded by a tool allowlist and lazily-loaded documentation rather than by a context budget. The costs that question was worried about — variable cost per review, no byte-reproducibility — are real and accepted.
- **Call shape** — the open question about one call per pull request versus one per file is answered by ADR 0004: a staged pipeline of triage, parallel specialists, independent verification and aggregation, with a model chosen per stage.
- **"Dropped before posting" for linter findings** — answered by the gate. Build, lint and tests run as a job the review job depends on, so the tools demonstrably ran and the skill states it as fact rather than detecting duplicates after the event.
- **"Lowest-confidence dropped first"** — replaced by a floor rather than an ordering. ADR 0007: confidence is calculated from a written rubric, set independently by the verifier, and gates posting at `high` by default.
- **Distribution** — ADR 0008. A consuming service holds a stub; the reviewer lives here.

## 2. Precision, defined before it is measured

`docs/metric.md`, written before the harvester.

Define precisely what counts as an accepted finding, a rejected finding, and an ignored one, and state how each is observed from the GitHub API. Ignored is the difficult case: silence means the finding was wrong, or that nobody looked, or that the pull request was merged in a hurry, and these are not the same thing.

Requirements:

- The primary metric is precision: of the findings posted, what fraction were acted on
- Recall is acknowledged as unmeasurable here, and the README does not pretend otherwise
- A stated minimum sample size below which a change in the metric means nothing
- The metric is computed by a command anyone can run, over stored review history

If the metric cannot be defined crisply, the loop cannot be evaluated, and that is the finding to report.

## 3. The outcome harvester

A scheduled job that looks at closed pull requests and records what happened to each posted finding.

Signals, in descending order of reliability:

- The comment thread was resolved, and by whom
- A commit after the comment changed the lines the comment pointed at
- An explicit reaction or reply
- The pull request merged with the comment neither resolved nor addressed

Requirements:

- Outcomes are stored in the repository, in a readable format, so the history survives the action being reinstalled
- Each outcome links back to the finding, the pull request, and the reviewer version that produced it
- The harvester is idempotent and safe to re-run over the same pull requests

This component is the one that makes the project more than a wrapper. Build it early.

## 4. The learned rules

A versioned Markdown file that the reviewer reads as part of its prompt.

Requirements:

- The action never writes to it directly. It opens a pull request proposing additions, with the evidence attached: which findings, which outcomes, how many observations.
- A proposed rule states a condition and an instruction — what to look for, or what to stop reporting — and cites the outcomes that motivated it
- A ceiling on the file's size, with the least-supported rules retired when it is reached. An unbounded rules file becomes an unread prompt.
- Every rule records when it was added and on what evidence, so a rule can be traced and removed

Open question for an ADR: are learned rules global to the repository, or scoped to paths and languages? Scoping is more accurate and needs far more observations before any rule is supported.

Open question for an ADR: what evidence threshold justifies proposing a rule? Too low and the reviewer learns noise; too high and the loop never closes.

## 5. Cost and reproducibility

**Amended.** Two of the three requirements below assumed the reviewer was a program making API calls it controlled. Under ADR 0003 it is an agent loop inside `claude-code-action`, which owns its own caching and reports its own usage.

What stands:

- Token usage and cost reported in the job summary, alongside the review summary the skill writes there
- The **metric** is reproducible even though the review is not: `cmd/metrics` reads only committed outcome records, calls no APIs, and produces the same numbers on the same file. Every figure quoted anywhere in this repository comes from it.

What was dropped, and what replaced it:

- **Prompt caching verification** — not ours to verify. The action manages the conversation; a requirement to prove cache hits we do not control is a requirement to fabricate one.
- **Offline replay** — abandoned as specified, and this is the largest thing lost in the move to an agent. It assumed deterministic inputs and a reviewer that could be re-run over stored ones. Comparing two skill versions now means running both against the same closed pull requests, which costs real money and still does not produce identical inputs.

  The replacement is weaker and honest: every finding records `skill` and `rules` (ADR 0006), so two reviewer versions can be compared *after the fact* over the pull requests each actually reviewed. That is observational where the original was experimental — it needs many more pull requests to say anything, and it cannot answer "what would the old reviewer have said about this diff". The floors in `docs/metric.md` and the holdout exist because that comparison is weak, not because it is strong.

  Anyone who finds a way to make replay affordable should write the ADR. It has not been found here.

## 6. Evidence

`docs/evidence.md`. Run the reviewer over a real repository across a meaningful number of pull requests, with the loop off, then with it on. Report precision for both.

Requirements:

- The repository used is named, and the pull requests are linked
- Findings that were accepted and findings that were rejected are both shown, with examples of each
- If precision did not improve, that is the result, and it is reported in the same detail

## 7. Failure modes

`docs/failure-modes.md`. At minimum:

- What the reviewer does on a diff too large for one call
- Behaviour when the model is unavailable or rate-limited, and why a failed review must never block a merge
- What happens when a learned rule turns out to be wrong, and how it is retracted
- Prompt injection through pull request content, and what the reviewer is prevented from doing as a result
- The self-reinforcement risk: findings suppressed by a learned rule are never posted, so they can never generate a contradicting outcome. State the mitigation, or state plainly that there is none.

The last one is the honest weakness of the whole design. Do not bury it.

## 8. README

The claim, the loop as a diagram, how to install it, the permissions it needs and why, the measured result from section 6, and a link to the failure modes above the fold.

---

## Suggested order

1. The metric document, argued, before any code
2. The reviewer, no learning, findings posted inline
3. The outcome harvester and stored review history
4. Baseline precision measured over real pull requests
5. Attribution and the holdout, so a later comparison can be attributed at all
6. Learned rules file, read-only: the reviewer consumes a hand-written one
7. The proposal pull request that writes to it, with evidence attached
8. Second measurement, and the honest comparison
9. Failure modes and README

Stop after each step.
