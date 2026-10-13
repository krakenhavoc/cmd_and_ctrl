# Contributing to C2

Thanks for your interest in C2 (cmd & ctrl). It's a free hobby project, built a few hours a
week and now in an open beta, and outside contributions are welcome.

This page is the short version. [AGENTS.md](AGENTS.md) is the full working guide: repo layout,
commands and conventions. It's addressed to AI coding agents, but everything in it applies to
people too.

## Find something to work on

- **Starter issues** are labelled
  [`good first issue`](https://github.com/krakenhavoc/cmd_and_ctrl/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).
  Each is self-contained and says what "done" means.
- **The `tier:` labels show how much an issue involves:**

  | Label | What it involves |
  |---|---|
  | `tier:1-mechanical` | One area of the code, with a recipe or a sibling to copy. |
  | `tier:2-standard` | Several files, following an existing pattern or ADR. |
  | `tier:3-design` | Needs an ADR or an owner decision first. Ask on the issue before starting. |

- **Claim an issue** by commenting on it, so two people don't build the same thing. Skip issues
  labelled `in-progress` or `question`.
- **For something new,** open an issue first. A quick "would you take this?" saves you building
  something that clashes with a decision already made in [docs/decisions/](docs/decisions/).

## Set up

You need Go and Node, or the devcontainer in `.devcontainer/`.

```sh
make server-dev   # the Go server on :8080
make client-dev   # the Vite dev server on :5173
make test         # server tests and client typecheck
make lint         # go vet, ESLint and Prettier
```

- **Client tests:** `npm test` in `client/`.
- **Without a local Go toolchain:** `scripts/go-docker.sh` runs Go in Docker. See
  [AGENTS.md §5](AGENTS.md#5-commands-youll-actually-run).
- **Adding or fixing a card:** start with [docs/adding-cards.md](docs/adding-cards.md).

## Make the change

- **Base branch:** branch from **`develop`**, not `main`. `develop` is the integration branch
  and deploys to [cmd-dev.labxp.io](https://cmd-dev.labxp.io). `main` is the release branch.
- **Tests:** a bug fix comes with a regression test that fails before the fix and passes after
  it.
- **Rules citations:** a rules change cites the Comprehensive Rules (for example "CR 707.9d"). See
  [AGENTS.md §6](AGENTS.md#6-when-youre-unsure).
- **Scope:** keep a pull request to one logical change. It is squash-merged into one commit.

## Open the pull request

- **Target `develop`.** GitHub offers `main` by default, and a feature PR into `main` fails the
  `promotion-guard` check. On the command line, use `gh pr create --base develop`.
- **Title it as a Conventional Commit:** `type(scope): summary`, for example
  `fix(cards): Cultivate can take fewer than two lands`.
  - **Types:** `feat`, `fix`, `chore`, `docs`, `refactor`, `test` or `spike`.
  - **Scope:** the top-level directory (`server`, `client`, …).
  - **Why it matters:** the squash-merge turns the title into the commit on `develop`.
- **Fill in the PR template:** Summary, Sprint, Issues and Test plan.
  - **Sprint:** for work that isn't on a sprint, write `Backlog`.
  - **Issues:** put `Closes #<n>` on a line of its own. The issue closes when the change
    reaches `main`.
- **Commit messages:** the format is in
  [AGENTS.md §4](AGENTS.md#commit-message-format). Don't worry about getting every trailer right.
  A maintainer writes the final squash commit.

## CI on pull requests from forks

**CI does not run on pull requests from forks.** This is on purpose: some of the project's
runners can reach its servers. Your PR will show a failing `cmdctrl-ci` check that says forks
aren't built. That isn't your fault, and you don't need to do anything about it.

- **Run the checks locally before you open the PR:** `make test`, `make lint` and `npm test` in
  `client/`.
- **A maintainer runs CI** on your exact commits by pushing them to a branch in this repository.
  If the checks pass, your change may land through a maintainer's PR that carries your commits.
  You stay the author, and your PR is closed with a link to the one that merged.

## Review

- **Expect review comments.** Even good PRs usually get one round.
- **Stalled PRs:** if a PR has had no reply for a couple of weeks, a maintainer may finish it or
  close it with a note. You're welcome to pick it up again at any time.
- **AI-assisted work** is fine. You're responsible for what you submit, so read and test it
  first.

## Reporting bugs

Use the in-app **Report a bug** button, which attaches the client log and a replay, or open a
[GitHub issue](https://github.com/krakenhavoc/cmd_and_ctrl/issues).

## License

C2 is licensed under the [GNU AGPL v3.0 or later](LICENSE); [NOTICE](NOTICE) lists what the license
does not cover.

- **Your contributions:** by opening a pull request, you agree that your contribution is licensed
  under those same terms.
- **Your right to submit it:** you confirm that you have the right to contribute the code. If you
  write it at or for work, check that your employer allows it.
- **Dependencies:** a new dependency must be under a license compatible with the AGPL, such as
  MIT, BSD, Apache-2.0, ISC or MPL-2.0. Say in the PR if you're unsure.

## Legal

C2 is unofficial fan content under the
[Wizards Fan Content Policy](https://company.wizards.com/en/legal/fancontentpolicy). Don't add
Wizards of the Coast assets to the repository. Card data and images come from
[Scryfall](https://scryfall.com) at runtime.
