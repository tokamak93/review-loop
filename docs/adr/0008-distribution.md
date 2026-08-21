# 0008 — A stub in every repository, the reviewer in one

**Status**: accepted

## Context

ADR 0003 left distribution as the weak point: "the repository becomes something to copy, not something to install. Consumers copy two workflows, two skills, and a lint config."

That is tolerable for one repository and absurd for two hundred. A fleet of microservices each holding its own copy of the review skill means every improvement to the reviewer is a two-hundred-pull-request migration, which means the improvement does not happen. The reviewer would be frozen by its own distribution model — and a reviewer that cannot change cannot learn, which is the only thing this project is for.

The target is the shape that worked in practice elsewhere: a stub in each service that calls a central reviewer, so the central thing can be fixed once.

## Decision

Two mechanisms, because two different things need distributing.

**The workflow is a reusable workflow.** `review.yml` declares `on: workflow_call`. A service holds a stub of about a dozen lines — `templates/consumer-review.yml` — that calls it and pins a tag. Everything that decides behaviour (models, turns, tool allowlist, effort and gate defaults, the concurrency policy) lives in review-loop.

**The skills and agents are a Claude Code plugin.** `.claude-plugin/marketplace.json` publishes this repository as a marketplace; the workflow installs it with `plugin_marketplaces` and `plugins`. The skills are fetched at run time from the pinned ref, not copied into the service.

The marketplace entry uses `source: "./"` with `strict: false` and explicit `skills` and `agents` paths pointing at `.claude/`. One copy of every file, serving both local discovery when working in this repository and remote installation everywhere else. A second copy under `plugins/` would drift, and drift in the reviewer's text is exactly what the metric is meant to catch — it should not be introduced by the packaging.

**The gate is a separate reusable workflow.** `go-checks.yml` is the strict Go gate; a service in another language substitutes its own job and calls `review.yml` unchanged. The review does not care what the gate was, only that one passed, which is why coupling them into one workflow would have been wrong.

**Outcomes and learned rules stay in the consuming repository**, under `.review/`. The reviewer is central; what it has learned is local.

## Why the rules are not central

This is the decision most worth arguing with, so the reasoning is recorded rather than assumed.

`CLAUDE.md` states the claim: the loop raises precision on *this repository's* conventions, which no general-purpose reviewer can know. A rule learned from one team's rejections is evidence about that team. Pooling rules across two hundred services would produce a file that is either bland enough to apply everywhere — in which case it is a prompt, not a lesson — or specific enough to be wrong in most places it lands.

It costs the obvious thing: a genuinely general lesson has to be learned two hundred times, or promoted by hand into the review skill itself. Promotion by hand is the intended path, and it goes through a pull request against this repository like any other change to the reviewer.

## Consequences

**Moving the `v1` tag ships a new reviewer everywhere at once.** That is the benefit and the risk in one sentence. `docs/failure-modes.md` covers what a bad reviewer version does and how to get out of it; the short version is that services pin a tag, the tag is moved deliberately, and this repository reviews its own pull requests off `main` so it feels a bad change first.

**Attribution now has something stable to point at.** `github.job_workflow_sha` is the commit of the reusable workflow's own repository, which is exactly "which reviewer produced this finding" — see ADR 0006.

**The metrics command has to travel to the data.** The harvest workflow checks this repository out beside the consuming one, builds `cmd/metrics` to a fixed path, and allows the harvest skill to run only that binary. The skill cannot compute precision by hand because the tool is the only arithmetic it is permitted.

**A private fleet needs a token that can read this repository.** `plugin_marketplaces` and the `actions/checkout` of the metrics command both fetch from review-loop. Public repository, no problem; a private mirror needs a PAT, and that is a real adoption cost not yet solved here.

**The reusable workflow cannot see the caller's `concurrency` needs.** Both are set centrally: `cancel-in-progress: true` on the gate, where superseded work is pure waste, and `false` on the review, where a cancelled run can leave half its comments posted.

## Alternatives rejected

**A published composite or Docker action.** The strongest distribution story, and it puts the reviewer's behaviour back inside a build artefact — which ADR 0003 rejected, because the loop's whole purpose is to rewrite behaviour that a human can read in a pull request.

**A git submodule or a sync bot pushing the skills into each service.** Works, and every service then carries a copy that can be locally edited, which makes "which reviewer produced this finding" unanswerable across the fleet.

**Copying the skills in the workflow rather than installing a plugin.** Considered: check the repository out into a path and copy `.claude/skills` to `~/.claude/skills`. Fewer moving parts and it works, but it clobbers a directory the runner may already be using, and `plugins` is the mechanism built for this. Kept in reserve if the plugin path proves unreliable in CI.
