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
2. **"Whenever players finish voting"** (#2927): Grudge Keeper, Model of Unity, Erestor of the Council. No event marked the end of a vote and the finished ballots were not kept for a trigger to read. **Shipped by the amendment of 2026-10-13 below.**
3. **The other voting cards** (#2928). Most are card work on this seam. Selvala's Stampede needs "reveal until you reveal N creature cards" and a put from hand of up to N cards at once; Custodi Squire votes for cards in a graveyard, and an option whose card is not on the battlefield is dropped from the offer today; Illusion of Choice makes one player choose every vote for a turn. **The first twelve shipped in #3107; the last five (Selvala's Stampede, Custodi Squire, Trial of a Time Lord, Illusion of Choice, The Valeyard) by the second amendment of 2026-10-13 below.**

## Consequences

**Cards.** Galadriel, Elven-Queen, Council's Judgment, Expropriate, Plea for Power, Coercive Portal, Magister of Worth, Tivit, Seller of Secrets, Brago's Representative and Ballot Broker ship Full.

**Cost.** One `PendingChoice` field (`CouncilVote`, an additive v7 snapshot field `pendingChoices[].councilVote`), one effect-key registry, one event and log kind, one `CardDef`/`Spec` slot, one wire field, and the read-only form of the dock's vote request. No new prompt kind.

**The thing most likely to be got wrong later.** Reading a ballot's index as an index into the vote's options. A ballot offers the options still there, plus "Don't vote again" on an optional vote; the vote's own index is `CouncilVote.OptionIndex(option)`, and on the wire `council_vote.offered[i]`.

---

## Amendment 2026-10-13 ([#2927](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2927)): "Whenever players finish voting"

Grudge Keeper, Model of Unity and Erestor of the Council trigger when a vote ends and read who voted for what in the vote that just ended. Decision 5's `EventVoteCast` is the ballot as it is cast; nothing marked the end, and the vote itself is gone from the game once its continuation runs (`PendingChoice.CouncilVote` lives on the open ballot only).

### 9. One event, carrying the ballots

`finishVoteLocked` emits `EventVotesFinished` (`votes_finished`) when the last ballot is cast: `Actor` is the vote's controller, `Source` the card that called it, `Amount` how many ballots were cast, and the new `Event.Ballots` lists every ballot in the order cast. It is emitted **before** the vote's continuation runs, so the log reads the last ballot, the end of the vote, then what the ability did with it. A "finish voting" trigger goes on the stack once the ability that called the vote has finished resolving either way (CR 603.2); it is not a trigger of the vote's own resolution.

The ballots ride the event rather than a record on the game, for three reasons:

- A trigger's stack item keeps its event (`ctx.Trigger().Event`), so the ability reads the ballots **as it resolves**, not as they stand then. A second vote started in response, or another vote later in the turn, cannot change what this one read.
- The event is plain data and so is copied when emitted; nothing aliases the vote it came from. The effect never captures a pointer, which is the rule for every triggered ability's `Effect`.
- The event log and every queued trigger's event are already in the snapshot. The change is one additive key, `ballots`, on the `Event` paths of the v8 shape file (`events[]`, `pendingTriggers[].trigger.Event`, `stackMeta[].trigger.Event`, the last-known stack and the paused mode walk), absent from older files and from every other event, so no schema bump: a file without it reads as a vote with no ballots, and a binary before it that drops the key restores a game whose waiting trigger sees an empty vote, which is the worst case of the unchanged behaviour (nothing happens) and not a corruption.

A vote that never began does not emit it: `RunVoteThenForEffect`, the path for a card with nothing to vote for (no permanent qualified), runs the continuation without announcing a finish, because nobody voted and nobody "finished voting". A vote that began and in which every ballot was dropped (each voter left, CR 800.4a) is a finished vote with no ballots, and does emit it.

The public log does not narrate the event (`silentEventKinds`): every ballot is a `vote` line and the trigger is a `trigger` line when it is queued.

### 10. Reading "a choice you voted for"

`game.VotedFor(ballots, voter)` is the distinct options a player voted for. The two predicates the cards need are read against the same ballots:

- `VotedForAChoiceVotedForBy(ballots, voter, you)`: some option the voter voted for is also one you voted for.
- `VotedForAChoiceNotVotedForBy(ballots, voter, you)`: some option the voter voted for is not one you voted for.

Three cases decide them, and the tests pin each:

- **A player with an extra vote (CR 701.38d)** names every option either of their votes went to. A controller who voted for two options has both in "a choice you voted for"; an opponent who split two votes across one option the controller took and one it did not is in **both** groups, and Erestor gives them a Treasure *and* counts them in X. They did vote for each.
- **A voter who cast no ballot** voted for nothing, so they are in neither group. Erestor's X does not count them.
- **A controller who cast no ballot** has voted for no choice, so no opponent shares one (Model of Unity: only you scry; Erestor: no Treasures), and every choice an opponent voted for is "a choice you didn't vote for" (Grudge Keeper drains each opponent who voted at all).

"Each opponent" is each opponent still in the game when the ability resolves (`Context.Opponents`), whoever called the vote: the abilities trigger on every vote at the table.

### 11. The card side

`effects.WheneverPlayersFinishVoting(label, effect)` is the trigger (always applies; one per vote, so extra ballots do not multiply it), `FinishedVoteBallots(ctx)` reads the event off the resolving item, and `OpponentsWhoVotedForAChoiceYouVotedFor` / `OpponentsWhoVotedForAChoiceYouDidntVoteFor` return the opponents in seat order. Model of Unity's "you and each opponent ... may scry 2" is a confirm and a scry per player in turn, you first, the next player asked once the previous scry is answered; each prompt is an ordinary confirm or scry addressed to its player, so the enumerator and the bot already answer it. Erestor puts its draw in the scry's `Then`, as Preordain does, so the draw sees the order the player chose.

### Consequences

Grudge Keeper, Model of Unity and Erestor of the Council ship Full. The `players-finish-voting` seam row is closed. A vote's continuation cannot see the event (it runs with the `VoteResult`, which already holds the ballots), and a future ability that wants "whenever you vote" wants `EventVoteCast`, not this.

---

## Amendment 2026-10-13 ([#2928](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2928), second slice): votes over graveyard cards, choosing how everyone votes, and the five cards

Selvala's Stampede, Custodi Squire, Trial of a Time Lord, Illusion of Choice and The Valeyard each wanted one thing the vote did not have. This amendment adds exactly those, and nothing else about the ballot changes: still one option pick per ballot, still public, still a registered continuation.

### 12. An object vote knows its zone

Decision 4 said each ballot offers "a permanent still on the battlefield". Custodi Squire votes for cards in a graveyard, which that rule dropped from every ballot, so the vote now carries where its cards are.

- `CouncilVote.Zone` (`VotePrompt.OptionZone`, `effects.Vote.Zone`) is the zone kind the options' cards are in. Empty means the battlefield, which is what every earlier vote meant, so no shipped card changes. `game.ZoneGraveyard` offers an option while its card is in a graveyard.
- A card that leaves the zone while the players vote is not offered to the next ballot, and the continuation re-checks it before it acts (Custodi Squire does not return a card that was exiled meanwhile). It is a new object once it has left (CR 400.7).
- The vote's controller's graveyard is read once, as the trigger resolves, as the options; the list is fixed then like every vote's.
- Wire: `council_vote.zone` (`"graveyard"`), omitted for a vote over permanents, words or players. Snapshot: `pendingChoices[].councilVote.zone`, one additive key in v8 (shape file and drift test), absent from older files, which read as the battlefield.
- The bot: a vote with a zone is not removal. The controller's bot votes for the best card, an opponent's for the worst, both leaning towards a card that already has votes (`aiseat/heuristic/vote.go`).

### 13. "You choose how each player votes this turn"

Illusion of Choice puts every ballot of every vote until end of turn in one player's hands, so it is turn-scoped state that must survive Clone and a snapshot. It is a stored scoped effect, the shape Jared Carthalion's "can't become the monarch" uses (ADR 0096 amendment, #2039), not a flag on the game:

- `ModVoteChooser` (`"voteChooser"`), scope `ScopeGame`, `Player` the chooser, swept at cleanup (CR 514.2). `Game.ChooseHowEachPlayerVotesThisTurnForEffect` registers it. An older binary refuses a file naming a kind it does not know rather than letting every voter vote for themselves.
- `voteChooserLocked(voter)` is read when each ballot is asked, so a vote that begins later the same turn is covered, and a chooser who leaves the game stops choosing (CR 800.4a: their effect has nothing to apply it). Where several records stand the newest wins (CR 613.7), so the last Illusion cast chooses.
- The ballot is still the voter's: `CouncilVote.Voter`, the tally, the ballots and the log's `vote` line (Actor) all name the voter. What changes is whom the prompt is owed to: `PendingChoice.Chooser` and `FromPlayer` are the chooser, and the header says "you choose <name>'s vote". Order and extra votes (CR 701.38d) are untouched: the chooser is asked each of the voter's ballots in turn, optional ones included.
- Departure. A chooser who leaves with a ballot open drops it like any chooser's, and the ballot goes back to the voter (the drop path re-asks it). A voter who leaves while the chooser holds their ballot has it withdrawn from the chooser (`dropChoicesForPlayerLocked`) and casts nothing; `castBallotLocked` also refuses an answer for a voter no longer in the game.
- The wire needs nothing new: `council_vote.voter` is the voter and `chooser` on the prompt the chooser, so the dock reads "<chooser> is choosing <voter>'s vote" from fields that were already there. The enumerator offers the ballot to its chooser, one answer per option, and the bot answers it for its side: hints are read relative to the bot, whoever the voter is.

### 14. Up to N from hand, and reveal until N

Two primitives grew a count; neither needed a new one.

- `effects.PutFromHandOntoBattlefield.Max` is "up to N" (exactly N when not optional), and N cards enter as one batch through the new `game.PutCardsFromHandOntoBattlefieldThenForEffect`, the hand twin of the library batch on the same entry machinery. Every CR 614 window sees the board as it stood before any of them arrived; every ETB fires after all have landed. `PutFromHandResult.EnteredAll` lists them.
- `revealUntilCount` is `revealUntil` stopping at the Nth hit. `revealUntil` is its N = 1 case.

### 15. The cards

- **Selvala's Stampede** (Full). Wild reveals until a creature card for each wild vote and puts them all onto the battlefield, then shuffles the library; free puts up to that many permanent cards from hand at once, after the wild half has settled. With no wild vote nothing is revealed and the library is not shuffled, since there is no "rest" to shuffle into it.
- **Custodi Squire** (Full). An object vote over the controller's artifact, creature and enchantment cards in their graveyard (decision 12); every card tied for the most votes and still there returns to hand.
- **Trial of a Time Lord** (Full). Needs nothing new: a Saga is an ordinary chapter trigger and chapters I to III are the "until" exile (`ExileUntil`). Chapter IV is a will-of-the-council vote; guilty tucks "the cards exiled with this Saga" (read per chapter, as Skyclave Apparition reads its card, CR 607.2a) on the bottom of their owners' libraries as one exit, an owner of several ordering them (CR 401.4), through the same helper as Terminus. A tie is innocent. The Saga is not sacrificed while the vote is open, so the exile has not ended when the cards are tucked.
- **Illusion of Choice** (Full). Decision 13, and a card.
- **The Valeyard** (caveat). Its extra vote is `ExtraVoteYouMay`, as Ballot Broker's. "Face that choice an additional time" is not built: villainous choice (CR 701.55) is a keyword action nothing in the engine asks, and no card in the catalog could make an opponent face one, so the clause has no choice to double. That is weaker than printed, the safe direction, and the card says so. Building the keyword is its own seam.

### Consequences

Selvala's Stampede, Custodi Squire, Trial of a Time Lord and Illusion of Choice ship Full and The Valeyard ships with a caveat. Every cost is additive: one `CouncilVote` field, one scoped-effect kind, one wire key, one engine door (the hand batch), and two counts on existing primitives. A vote over cards in a zone other than the battlefield or a graveyard, or a chooser who answers some ballots and not others, is not built; the first card that wants either extends the same two fields.

**Back-outs.** Each shape has a test that fails without it: `TestAVoteOverGraveyardCardsOffersWhatIsStillThere` (the zone), `TestAChosenVoteIsAnsweredByTheChooserForEachVoter` and `TestAVotersChosenBallotIsDroppedWhenTheyLeave` (the chooser), `TestSelvalasStampedePutsUpToOneCardPerFreeVote` (the count on the hand put) and `TestSelvalasStampedeRevealsUntilACreatureForEachWildVote` (the count on the reveal).
