# ADR 0226 — Subgames (Shahrazad) and dexterity cards (Chaos Orb, Falling Star)

**Status:** Proposed · 2026-11-24 · S82 — Card coverage: sets outside Standard, newest first. Nothing is built. Two owner questions are open, under [Questions for the owner](#questions-for-the-owner).
**Issue:** [#6103](https://github.com/krakenhavoc/cmd_and_ctrl/issues/6103). The S82 OLD batches that wait on it are LEG b864 (Falling Star), ARN b890 (Shahrazad) and LEA b894 (Chaos Orb) in `plan/s82-batches.json` on `work/s82-batches`.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-11-24: `git fetch --all --prune`, then `docs/decisions/` listed on every remote head (539 branches, `origin/develop` and `origin/main` among them). The highest number on any of them is 0225. This ADR takes **0226**.
**Builds on:** [ADR 0201](0201-restarting-the-game.md) (the one place the engine already replaces a whole game in place), [ADR 0222](0222-formats-beyond-commander-standard-1v1-and-booster-draft.md) §8 (cube, the only format that can reach these cards).

---

## Context

Three cards in the S82 old-sets pass need something the engine does not have:

| Card | Oracle text (Scryfall) | `legalities.commander` |
|---|---|---|
| Shahrazad (ARN) | Players play a Magic subgame, using their libraries as their decks. Each player who doesn't win the subgame loses half their life, rounded up. | banned |
| Chaos Orb (LEA) | {1}, {T}: If this artifact is on the battlefield, flip it onto the battlefield from a height of at least one foot. If this artifact turns over completely at least once during the flip, destroy all nontoken permanents it touches. Then destroy this artifact. | banned |
| Falling Star (LEG) | Flip Falling Star onto the playing area from a height of at least one foot. Falling Star deals 3 damage to each creature it lands on. Tap all creatures dealt damage by Falling Star. If Falling Star doesn't turn completely over at least once during the flip, it has no effect. | banned |

**Where a player could meet them.** All three are banned in Commander, so the deck check refuses them (`deck/validate.go`), and none is Standard-legal. The one route into a game is a cube (ADR 0222 §8), and a cube can be drafted only when every card in it is Full. So a card from this ADR is reachable by a player only if it ships Full with no caveats; a caveated stand-in would sit in the catalog and be unreachable outside the dev card spawner.

### Subgames (CR 729, 2026-09-25 edition)

- **729.1a** A subgame is a completely separate game; the main game is discontinued while it runs and resumes when it ends.
- **729.1b** Nothing created in one game means anything in the other, except what the creating effect says (Shahrazad: each player who doesn't win loses half their life).
- **729.2** New zones; each player moves their whole main-game library to the subgame library and shuffles; the starting player is chosen at random; rule 103 applies (opening hands, mulligans). **729.2c** In Commander, each commander in the main-game command zone moves to the subgame command zone.
- **729.4 / 729.4a** Main-game objects are outside the subgame. A card brought into the subgame from the main game triggers main-game leave abilities, which wait until the main game resumes.
- **729.4b** Main-game counters on a player are not in the subgame and come back afterwards; subgame counters on a player cease to exist.
- **729.5** At the end, each player puts every traditional card they own in the subgame (exile and phased-out included, the command zone excepted) into their main-game library and shuffles. **729.5c** Commanders go back to the main-game command zone. The creating spell finishes resolving, then the main-game triggers from 729.4a go on the stack.
- **729.6** A subgame can be created inside a subgame.

### What the engine has

One `*Game` per table, held with its mutex by the room, the hub, the bot runners and every handler. ADR 0201 already had to replace a game and could not swap the pointer: it builds a fresh `NewGame()` and adopts it with `RestoreFrom`, and carries a short named list across. A subgame is that, twice, with a game in between that must be kept whole:

1. Clone the main game and keep it as a suspended game, mid-resolution of Shahrazad.
2. Build the subgame fresh (libraries and commanders moved, random starting player, the opening-hand procedure) and adopt it in place.
3. When the subgame ends, collect each player's traditional cards, adopt the suspended main game back, put those cards into the libraries and shuffle, apply the result (half life, rounded up, to each non-winner), finish Shahrazad's resolution, then put the 729.4a triggers on the stack.

## Analysis: can the seam be built cleanly now?

**Snapshot.** The suspended main game would be one new optional field holding a nested game snapshot. That is additive within `SnapshotSchemaVersion` 9 and needs no fixture change, so the schema rule (#522) does not by itself forbid it. Two parts of it are not small:

- **The resolution must resume from data.** The main game is suspended in the middle of Shahrazad's effect, and a restore can happen while the subgame runs (undo, a server restart, a snapshot restore). The remainder of the effect ("each player who doesn't win loses half their life") therefore has to be a recorded continuation keyed by the catalog ref, like ADR 0201's `RestartTail`, never a closure on `Game` (the closure ratchet, `closure_fields.txt`).
- **Card identity across two games.** Each library card is one object in the subgame while the main game still names it in its own history (the event log, "exiled with" links, LKI). The subgame must give every card a new object (CR 400.7) and the return must map them back, which ADR 0201 avoided because its old game ceased to exist.

**Everything that reads "the game".** A subgame is a whole game nested in another, and every consumer of the table assumes one:

- **Undo and restore points** must not cross a subgame boundary in a way that drops the suspended game.
- **Match records, game stats and pods** must not record a subgame's end as a game result, and the match must not end when the subgame does.
- **Eliminations.** "Win the subgame" means the last player standing in a four-player subgame; a player who loses it is out of the subgame only. The loss path (`Outcome`, the seat's lost state, the leave-the-game cleanup of CR 800.4a) must be scoped to the subgame.
- **Restarting the game** (ADR 0201) inside a subgame restarts only the subgame (CR 727.6), and a subgame can nest (729.6).
- **Protocol and client.** The client needs to know a subgame is running, that the main game is suspended, and what it holds (life totals and permanents that are out of play for now). This is a new view state on the wire, plus table UI.
- **Bots.** `aiseat` must play a subgame and know the stakes (losing costs half its main-game life, not the match).

That is cross-cutting engine, protocol, client, stats and bot work: `tier:3-design` and larger than one card justifies. The brief for #6103 says to stop at a design in that case, so this ADR proposes the design and builds nothing.

### Dexterity cards

Neither card has a rules meaning off a physical table: what a flipped card "touches" or "lands on" is a fact about cardboard on a surface, and the Comprehensive Rules give no digital equivalent. No ADR or owner decision covers them (the nearest, the 2026-11-16 ante decision in `docs/adding-cards.md`, is about a zone C2 does not have, not a physical act). Any digital version is an invented ruling, so it cannot be Full: a player-chosen set of permanents is stronger than printed, which `docs/adding-cards.md` calls the wrong direction, and a random pick is a different card. Since only Full cards reach a cube, a caveated version would be reachable by nobody.

## Proposal

1. **Chaos Orb and Falling Star:** left out of the catalog for good, recorded in `docs/adding-cards.md` next to the ante decision, so set-coverage batches stop asking. (Owner question 1.)
2. **Shahrazad:** not built now. If the owner wants it, it is its own `tier:3-design` epic built on the three steps under *What the engine has*, with the continuation as data, a nested optional snapshot field (additive in v9), and the stats, protocol, client and bot work listed above. Until then, it is listed as excluded for the same reason. (Owner question 2.)

## Questions for the owner

1. **Dexterity cards (Chaos Orb, Falling Star).** Which?
   - **(a) Leave them out for good** (recommended). They are banned in Commander and reachable only through a cube, which needs Full cards, and no digital version is Full.
   - **(b) A chosen stand-in:** on resolution the controller chooses up to N permanents (Chaos Orb: nontoken permanents; Falling Star: creatures, say N = 2) and the effect applies to them, with the "turns over" clause treated as always met. Ships Partial with a caveat, so it never reaches a cube.
   - **(c) A random stand-in:** the engine picks up to N of the legal permanents at random (a seeded, logged roll as in ADR 0054). Also Partial with a caveat.
2. **Shahrazad.** Which?
   - **(a) Leave it out for good** (recommended): banned in Commander, one card, and a subgame touches every part of the table.
   - **(b) Build the subgame seam** as a scheduled epic (its own sprint), on the design above.
   - **(c) Park it** in the backlog as a low-priority seam with this ADR as its design, and revisit if a cube host asks for it.

## Consequences

- The three S82 batches (LEG b864, ARN b890, LEA b894) stay blocked on #6103 until the owner answers. With answer (a) to both, they close as excluded and #6103 closes with no engine change.
- No code, schema or fixture changes in this ADR.
