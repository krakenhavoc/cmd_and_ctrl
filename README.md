# cmd_and_ctrl

A four-player Magic: The Gathering Commander table that plays in the browser. A Go server runs
the game, and a TypeScript and Svelte client renders it. Most cards resolve themselves: the
engine knows what thousands of catalog cards do, pays costs for you, and asks only the questions a
player would have to answer at a real table. A card the engine doesn't automate yet still plays,
by hand.

It's a hobby project, built for a small group of friends and played at
[cmd.labxp.io](https://cmd.labxp.io). Tables are joined by invite.

[![Sponsor](https://img.shields.io/badge/sponsor-%E2%9D%A4-db61a2?logo=githubsponsors)](https://github.com/sponsors/krakenhavoc)

## What's there

- **Rules that run themselves.** Priority, the stack, combat, triggers, replacement effects,
  layers, the commander rules and the loop breaker. Rules citations follow the Comprehensive
  Rules, and each card declares honestly how complete it is.
- **Bot seats.** Fill an empty seat with a bot that plays a curated deck. It uses a heuristic
  that prices what each card does, and can optionally use a local or hosted language model.
- **A public roadmap.** [cmd.labxp.io/#/roadmap](https://cmd.labxp.io/#/roadmap) lists the
  mechanics and engine seams that are done, partial or missing, and the cards waiting on each.
- **Deck check.** Paste a decklist at [cmd.labxp.io/#/decks](https://cmd.labxp.io/#/decks) to see
  which of its cards play automatically, play with a caveat, or still play by hand.

## Running it locally

You need Go and Node. `make help` lists the targets; the common ones are:

```sh
make server-dev   # the Go server on :8080
make client-dev   # the Vite dev server on :5173
make test         # server tests and client typecheck
```

[AGENTS.md](AGENTS.md) is the working guide: repo layout, commands, environment variables and
conventions. [PLAN.md](PLAN.md) covers the vision and roadmap, and
[docs/decisions/](docs/decisions/) holds the architecture decision records.

## Bugs and deck requests

Use the in-app "Report a bug" button, or open a
[GitHub issue](https://github.com/krakenhavoc/cmd_and_ctrl/issues). To ask for a deck's cards to
be automated, check the deck on the [decks page](https://cmd.labxp.io/#/decks) and use
**Request missing cards**. Requests need a Discord sign-in.

## Sponsoring

If you enjoy the project, you can support it through
[GitHub Sponsors](https://github.com/sponsors/krakenhavoc). It's a plain tip jar: sponsoring
unlocks nothing, because everything here stays free to use under the Fan Content Policy below.

## Legal

cmd_and_ctrl is unofficial fan content. It is not approved or endorsed by Wizards of the Coast.
Portions of the materials used are property of Wizards of the Coast. ©Wizards of the Coast LLC.
Magic: The Gathering is a trademark of Wizards of the Coast LLC. Card data and card images are
provided by [Scryfall](https://scryfall.com), which is not affiliated with this site. See the
[Wizards Fan Content Policy](https://company.wizards.com/en/legal/fancontentpolicy).
