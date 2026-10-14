# ADR 0180 — Secret council: ballots hidden until every player has voted

Date: 2026-10-14
Status: Accepted
Sprint: S58 — Deck requests, October batch
Issues: #2926 (seam: secret council). Registry rows: `secret-council` (closed by this ADR), `cant-attack-its-owner` (Elrond's half).
Builds on: [ADR 0146](0146-voting.md) (voting, CR 701.38), [ADR 0106](0106-five-small-seams-from-the-s50-rechecks.md) §2 (a creature that can't attack its owner), [ADR 0054](0054-dice-rolls-and-coin-flips.md) decision 4 (undo is not a way to see the future twice).
Numbering: every remote branch was listed with `git ls-tree --name-only <ref> docs/decisions/`. The highest number anywhere was 0179 (0178 is on `fix/2950-darkness-replacement-continuation`), so this is 0180.

## Context

Secret council is an ability word (CR 207.2c): it has no rules meaning of its own. The cards say what makes it different: "each player **secretly** votes for …, **then those votes are revealed**". The vote is still a CR 701.38 vote (CR 701.38a–d apply), so extra votes, "whenever players finish voting" and Illusion of Choice all reach it.

ADR 0146 decision 5 made every ballot public as it is cast, in two places: `EventVoteCast` is a `vote` line in the public log, and the open ballot carries every ballot and the tally on the wire for every seat (`PendingChoiceView.council_vote`). The bot's heuristic leans towards options that already have votes. A secret vote needs every one of those closed until the last ballot is in, for every viewer: the other seats, a spectator, the admin's omniscient view, the bot's view and the MCP seat (both read the filtered seat view).

Six cards wait: Elrond of the White Council, Mob Verdict, Círdan the Shipwright, Trap the Trespassers, Truth or Consequences and Vault 11: Voter's Dilemma. Between them they also want "another player" (the voter may not vote for themselves), "up to one creature" (a voter may vote for nothing) and, on Elrond, "this creature can't attack its owner" granted by a resolved effect.

## Decisions

### 1. A secret vote is the same vote, with `Secret` set

`game.CouncilVote.Secret` (`VotePrompt.Secret`, `effects.Vote.Secret`). Everything else is ADR 0146's: one `option_pick` per ballot, in turn order from the controller, the voter's extra votes asked together (CR 701.38d), Illusion of Choice's chooser, the departure table, the registered continuation. Ballots are asked one at a time, not all at once, because one prompt at a time is how every multi-player choice in the engine runs (ADR 0175), and nothing about the order leaks a ballot: who is voting now was always public.

### 2. Nothing carries a ballot before the reveal

- **The log.** A secret ballot emits `EventSecretVoteCast` (Actor and Source, no option), the log's `vote_secret` line: "Alice cast a secret vote (Mob Verdict)". `EventVoteCast` is not emitted for it.
- **The wire.** `viewOfCouncilVote` puts a secret vote's ballots in an unexported field (as `CardView.knowers` is), so the unfiltered view never serialises them, and sends `secret: true`, a zero tally and `voted` (the voters who have cast a ballot, which the log says anyway). `FilterViewFor` gives each seat back **its own** ballots and nobody else's; a spectator and the admin get none.
- **The bot and the MCP seat** read the filtered seat view, so they see what a human at that seat sees. The heuristic's "lean towards an option with votes" reads a zero tally and does nothing.
- **The snapshot** keeps the ballots on `pendingChoices[].councilVote` as today: a restore point is server data, never sent to a client, so a restored table is still mid-vote with its ballots hidden.

### 3. One reveal

When the last ballot is in, `finishVoteLocked` emits `EventVotesRevealed` (Actor the controller, Source the card, `Ballots` every vote in the order cast, `VoteOptions` the options' labels), then `EventVotesFinished` as before, then the continuation. The log projects the reveal as **one** `votes_revealed` entry listing each seat and what it voted for ("Votes revealed (Mob Verdict): Alice for Bob, Bob for Carol, Carol for Bob"). A voter who chose nothing (decision 5) is listed as voting for nothing; nothing is listed for a voter who cast no ballot. The client shows the reveal in the board's attention strip, beside dice results, as well as in the log.

### 4. Undo cannot reach back across a reveal

The last voter's ballot is the commit that reveals every other ballot, and the undo gate lets that voter take it back. Undoing it would let them vote again knowing how everyone voted, which is ADR 0054 decision 4's "seeing the future twice". So a seat's undo is refused when the commit it would take back emitted `EventVotesRevealed` (`ErrUndoPastSecretVote`, `undo_blocked: "secret_vote"`); the admin may still undo. Undoing a ballot before the reveal is unchanged: the restored clone keeps the earlier ballots, still hidden.

### 5. "Another player" and "up to one"

- `CouncilVote.OthersOnly` ("votes for another player", Mob Verdict): a ballot does not offer the option naming its voter. It is read per ballot, so an Illusion of Choice chooser choosing Bob's vote may not vote for Bob.
- `CouncilVote.Abstain` is the label of a "vote for nothing" option on every ballot ("votes for up to one creature", Vault 11: "No creature"). Choosing it casts that vote as no ballot: the tally is unchanged, `MostVotes` is empty when everyone abstains, and the voter is recorded in `CouncilVote.Abstained` for the reveal. On the wire its `offered` index is -2 (-1 stays "Don't vote again").
- `CouncilVote.SelfBest` is a bot hint like `ForController`: the option naming the voter is the one they want (Círdan: votes you receive draw you cards). The bot votes for itself when that option is offered.

All are additive keys on `pendingChoices[].councilVote` (and `Event.VoteOptions` on the event paths) in the v8 shape. A binary before them drops the keys and runs the vote open, which shows ballots early but corrupts nothing.

### 6. "Can't attack its owner", granted

`ModCantAttackOwner` (`"cantAttackOwner"`, layer 6) is ADR 0106 §2's `AttackTargetRestriction{NotOwner}` written by a scoped effect: the owner is read live, like the printed form. The restriction is attributed to the creature itself, because the effect grants the creature an ability ("they gain 'This creature can't attack its owner'"): a later "loses all abilities" removes it (CR 613.1f), an earlier one does not. It is #3400's `cantAttackPlayer` with the relation in place of a fixed player.

### 7. The cards

- **Elrond of the White Council** (Full). Fellowship or aid. Each fellowship voter chooses one creature they control per fellowship vote (ADR 0175's ordered collection, turn order from the controller); Elrond's controller gains control of all of them at once and each gains decision 6. Then a +1/+1 counter per aid vote on each creature Elrond's controller controls, the new ones included.
- **Mob Verdict** (Full). Another player. For each vote an opponent received, 2 damage to that player and each creature they control, one instance per vote; a card per vote the caster received.
- **Círdan the Shipwright** (Full). A player, enters or attacks. Each player draws a card per vote received, in turn order; then each player who received none chooses, in turn order, whether to put a permanent card from their hand, and those enter together.
- **Trap the Trespassers** (Full). A creature the caster doesn't control; per creature with votes, that many stun counters, then tap it.
- **Truth or Consequences** (Full). Draw a card per truth vote, then an opponent at random takes 3 damage per consequences vote (one instance).
- **Vault 11: Voter's Dilemma** (Full). Chapter I makes a 1/1 Human Soldier per opponent. II and III are a secret vote for up to one creature: no votes, each player draws; otherwise every creature with the most votes or tied is destroyed.

## Consequences

The `secret-council` seam closes and Elrond leaves the `cant-attack-its-owner` row's Waiting list. The cost is two event kinds, two log kinds, one scoped-effect mod kind, five additive vote keys, one wire flag with one projected list, and one undo refusal.

**The thing most likely to be got wrong later.** Reading a secret vote's ballots off the unfiltered `GameView`. They are not there: only `FilterViewFor` puts a seat's own ballots back. A new consumer that wants them must go through the filter, never through `game.CouncilVote`.
