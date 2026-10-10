# ADR 0146 — Voting: will of the council and council's dilemma (CR 701.38)

**Status:** Accepted · 2026-10-10 · Backlog — parked on a dependency or a single card
**Issues:** [#2143](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2143) (seam: voting and will of the council — Galadriel, Elven-Queen). Expropriate and Selvala's Stampede were added to its waiting list by the S58 deck requests ([#2077](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2077)). Registry row: `voting` (closed by this ADR). Follow-ups: [#2926](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2926) (secret council), [#2927](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2927) ("whenever players finish voting"), [#2928](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2928) (the other voting cards).
**Owner decisions:** none needed. Every choice below is fixed by the rules or follows an existing ADR.
**Numbering:** after `git fetch --all --prune`, every remote branch head was listed with `git ls-tree --name-only <ref> docs/decisions/`. The highest number anywhere was **0145**, so this is **0146**.
**Builds on:** [ADR 0018](0018-triggers-on-the-stack.md) §6 (a prompt that blocks the table), the #568 option pick and its keyed continuation (#2854), [ADR 0060](0060-leaving-the-game.md) (the departure table), [ADR 0111](0111-action-dock.md) (the dock and its vote request), [ADR 0114](0114-the-ring-tempts-you.md) (Galadriel's tempt).

---

## Context

### The rules

From the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`):

- **CR 701.38a:** "Some spells and abilities instruct players to vote for one choice from a list of options to determine some aspect of the effect of that spell or ability. To vote, each player, starting with a specified player and proceeding in turn order, chooses one of those choices."
- **CR 701.38b:** "The listed choices may be objects, words with no rules meaning that are each connected to a different effect, or other variables relevant to the resolution of the spell or ability."
- **CR 701.38c:** "voting" refers only to an actual vote.
- **CR 701.38d:** "If an effect gives a player multiple votes, those votes all happen at the same time the player would otherwise have voted."
- **CR 207.2c:** will of the council, council's dilemma and secret council are ability words, with no rules meaning of their own.

Will of the council reads which option "gets more votes", with a tie going to the option the card names for it. Council's dilemma does something "for each" vote. An object vote ("votes for a nonland permanent you don't control") reads the object or objects "with the most votes or tied for most votes". Nothing is targeted.

### What the engine had

`game.Vote` (politics.go) is the S10 politics scaffold behind the dock's "Call a vote…" (ADR 0111 PR 7): any player opens a vote on any topic, anybody may cast or re-cast a ballot at any time, anybody may end it, and nothing in the rules reads it. The client draws it as the dock's `voteRequest` (`client/src/lib/choiceDock.ts`), a `step` request with a button per option, its tally, and "end vote", and spectators get `VotingPanel`. The engine had no prompt that collects one answer from each player in turn order and hands the tally to the rest of the effect.

## Decisions

### 1. A rules vote is not the sandbox vote

A resolving spell or ability opens a rules vote; every player votes in turn order; nobody may change a ballot; nobody may end it early; the effect reads the tally. The sandbox vote is the opposite on each count, and sharing `game.Vote` would put an "end vote" button in front of a Council's Judgment. So the rules vote is its own data, `game.CouncilVote` (council_vote.go). What is shared is the client's drawing of a vote (decision 6).

### 2. One option pick per ballot

Each ballot is an ordinary `PendingChoiceOptionPick` addressed to the voter. `PendingChoice.CouncilVote` makes it a ballot: it carries the vote so far (options, ballots cast, the voters still to vote, the current voter's votes left) and moves to the next voter's prompt when the ballot is cast, so there is one copy of it at a time. Riding the option pick means the choice gate (it blocks the table, which is right: the effect is paused mid-resolution), the enumerator, the departure table, the wire, the client's inline prompt and sheet, and the bot's move list all answer it with no new kind.

- `ResolveOptionPick` hands a ballot's answer to `castBallotLocked` instead of a frame: the ballot is recorded, `EventVoteCast` is emitted, and the next ballot is queued.
- When the last ballot is cast, the vote's continuation runs with a `VoteResult`. The continuation is a registered key (`RegisterVoteThen`, card side `effects.VoteResultThen("vote/<card>", …)`) in the effect-key namespace and ledger (`vote <key>`), and the vote is plain data, so a table waiting on a ballot is a restore point. A restore point naming a vote continuation this binary lacks is refused and kept, like any other key.
- A voter who leaves the game with a ballot open casts nothing (CR 800.4a): the departure table's `dropDefault` for an option pick reaches `castBallotLocked` with no answer, and the vote goes on to the next player. Players who have left are skipped when their turn to vote comes.

### 3. Turn order, and extra votes (CR 701.38d)

`StartVoteForEffect` lists the players still in the game in turn order from the named player ("starting with you": the ability's controller on every shipped card). A player's extra votes are counted at the moment they come to vote, from the permanents they control then:

- `Spec.ExtraVote: game.ExtraVoteYouGet` is "While voting, you get an additional vote" (Brago's Representative): one more ballot they must cast.
- `game.ExtraVoteYouMay` is "While voting, you may vote an additional time" (Ballot Broker, Tivit): after their required votes they are offered one more ballot with a "Don't vote again" option.

All of a player's ballots are asked one after another before the next player votes, which is CR 701.38d's "at the same time the player would otherwise have voted". The static is read through `CatalogAbilityKey`, so a permanent that has lost its abilities gives nothing. `CardDef.ExtraVote` is classed `defFieldOwnRules` for the trigger-order check: its permanent changes how a vote resolves.

### 4. The options

A word is a `ChoiceOption` with a label (CR 701.38b's "words with no rules meaning"). An object is an option carrying its card (`effects.VoteForPermanents`, labelled with the permanent's name), a player an option carrying the seat. The list is fixed when the vote starts and never pruned; each ballot offers the options still there (a permanent still on the battlefield, a player still in the game), and an answer is mapped back to the vote's own index by its label and subject, so a shorter offer never renumbers the tally.

The result's readers are `VoteResult.MoreVotes(i, j)` (will of the council), `Votes(i)` and `VotersFor(i)` (council's dilemma, one entry per vote naming the voter) and `MostVotes()` (an object vote; an option nobody voted for is never among them).

### 5. Public as it is cast

There is no secret in CR 701.38. Each ballot is visible to the whole table the moment it is cast, two ways:

- `EventVoteCast` (Actor voted, Source called the vote, Label is the option as printed) is the public log's new `vote` line: "Alice voted for knowledge (Plea for Power)".
- The open ballot carries the vote on the wire for every seat, `PendingChoiceView.council_vote`: the options, the tally, the ballots, which offered option is which (`offered`, with -1 for "Don't vote again"), and the bot hints.

### 6. The client: the vote request, reused

The voter answers their ballot as any option pick: inline in the dock when every option is a short word, in the sheet when the options are permanents. Each option shows its votes so far (`ballotNote`). Every other seat sees the vote in the dock through the same `voteRequest` the sandbox vote uses, read-only: `councilVoteRequest` passes the vote's tally, names the request "vote" with "<player> is voting", and gives it no cast handler and no "end vote" button, so the option buttons are disabled. It is a `step` request, as the sandbox vote is, so anything a seat owes outranks it. Spectators see the log.

### 7. The bot

The ballot is an option pick, so every policy is already offered it. The heuristic scores it in `aiseat/heuristic/vote.go`:

- **Words.** The card puts two hints on the vote, one number per option: how much the vote's controller wants it to win (`ForController`), and how much an opponent of the controller does (`ForOpponents`). The bot votes for the best option for its side. Plea for Power's controller votes time and an opponent knowledge; Magister of Worth computes its hint from the graveyards as the vote starts. A card with no hint gets the first word.
- **Objects.** Every shipped object vote is removal, so the bot votes for the most valuable permanent it does not control, leaning towards one that already has votes so a table of bots does not split, and never for its own.
- **An extra vote it may decline** is worth half its favourite, so it is always cast, for the option it likes best.

### 8. The cards

All nine ship Full.

- **Galadriel, Elven-Queen.** The beginning-of-combat trigger's "if another Elf entered the battlefield under your control this turn" is an intervening if (CR 603.4) read from the turn's entry tally by subtype, Éowyn, Shieldmaiden's reading (`anotherEnteredWithSubtypeThisTurn`, now shared by both). Dominion is ADR 0114's tempt, then a +1/+1 counter on the creature chosen as Ring-bearer.
- **Council's Judgment.** The options are the nonland permanents the caster doesn't control; every permanent tied for the most votes is exiled at once. Not targeted, so hexproof and protection don't matter.
- **Plea for Power**, **Coercive Portal**, **Magister of Worth.** Will of the council over two words. The Portal's "sacrifice this artifact" and the Magister's "other than this creature" name the object the trigger came from: that is read as the trigger resolves (`sourceIsNewObject`) and carried on the vote, since nothing can move it while the players vote.
- **Expropriate.** An extra turn per time vote; per money vote the caster chooses a permanent the voter owns and gains control of it indefinitely (one prompt per voter, as many permanents as their money votes). "Exile Expropriate" is done as the vote starts rather than after it: the spell is still resolving while the players vote and nothing can see it in between, and the resolution never buries a spell that moved itself (#489).
- **Tivit, Seller of Secrets.** A Clue per evidence vote, a Treasure per bribery vote, and `ExtraVoteYouMay`.
- **Brago's Representative**, **Ballot Broker.** The two extra-vote statics.

## Out of scope

1. **Secret council** (#2926): Mob Verdict, Círdan the Shipwright, Trap the Trespassers, Truth or Consequences, Vault 11: Voter's Dilemma, and Elrond of the White Council. Their ballots must stay hidden until everyone has voted, which is the opposite of decision 5.
2. **"Whenever players finish voting"** (#2927): Grudge Keeper, Model of Unity, Erestor of the Council. No event marks the end of a vote and the finished ballots are not kept for a trigger to read.
3. **The other voting cards** (#2928). Most are card work on this seam. Selvala's Stampede needs "reveal until you reveal N creature cards" and a put from hand of up to N cards at once; Custodi Squire votes for cards in a graveyard, and an option whose card is not on the battlefield is dropped from the offer today; Illusion of Choice makes one player choose every vote for a turn.

## Consequences

**Cards.** Galadriel, Elven-Queen, Council's Judgment, Expropriate, Plea for Power, Coercive Portal, Magister of Worth, Tivit, Seller of Secrets, Brago's Representative and Ballot Broker ship Full.

**Cost.** One `PendingChoice` field (`CouncilVote`, an additive v7 snapshot field `pendingChoices[].councilVote`), one effect-key registry, one event and log kind, one `CardDef`/`Spec` slot, one wire field, and the read-only form of the dock's vote request. No new prompt kind.

**The thing most likely to be got wrong later.** Reading a ballot's index as an index into the vote's options. A ballot offers the options still there, plus "Don't vote again" on an optional vote; the vote's own index is `CouncilVote.OptionIndex(option)`, and on the wire `council_vote.offered[i]`.
