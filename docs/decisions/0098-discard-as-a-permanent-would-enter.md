# ADR 0098 — Discarding a card as a permanent would enter (Mox Diamond)

**Status:** Proposed · 2026-09-30 · Post-S30 — Rolling deck-driven catalog growth
**Issue:** [#1744](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1744). The card is on the ViviVoltron deck request [#1640](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1640), and was noted earlier on batch issue #295 and on #1600.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-09-30. I ran `git fetch --all --prune`, then read every `docs/decisions/` file name on every remote branch (36 heads, and 34 at the re-check before pushing). The highest number anywhere is **0097** (`0097-modes-that-havent-been-chosen.md`).
**Builds on:** [ADR 0013](0013-replacement-effects.md) — §5z (the reveal-from-hand entry choice, #1198), §5o (an entry can pause), §5g and §5af (a discard is an exit, and a cost settles now); [ADR 0061](0061-token-creation-and-discard-are-replaceable-events.md) §4–5 (`RepEventDiscard` and the discard cause) and its 2026-09-23 amendment (the resumable batch entry); [ADR 0062](0062-abilities-and-special-actions-from-the-hand.md) §3 (what a cost discard is, for contrast).
**Owner decisions:** none yet. This ADR is for review. The questions are at the end.

---

## Context

Mox Diamond — Artifact {0}:

> If this artifact would enter, you may discard a land card instead. If you do, put this artifact onto the battlefield. If you don't, put it into its owner's graveyard.
> {T}: Add one mana of any color.

It is one of the three ViviVoltron cards still waiting on engine work. It must not ship until this seam exists: a Mox Diamond that enters for free is stronger than printed, which is the one direction the catalog never errs in.

### What the engine has

`ReplacementEffect.EntryHandReveal` (#1198, ADR 0013 §5z, `server/internal/game/entry_reveal.go`) is the nearest shape. It asks a card choice over the entering permanent's controller's hand, inside the CR 614 entry window, through a `PendingChoiceEntryRevealFromHand` prompt. It carries the choose-cards payload and a `chooseCardsFrame{zone: ZoneHand}`, so `checkChooseCardsPicksLocked` validates the answer for the submit path and the enumerator alike. Its decline runs the effect's `Replace` (the reveal-lands' "enters tapped"); an answer runs `RevealForEffect` and skips `Replace`. §5z Decision 4 named Mox Diamond as the case it deliberately left out.

### What is missing

The issue names two gaps. Reading the code turned up two more, and corrected one premise.

1. **The pick only reveals.** Mox Diamond needs the pick *discarded*, through `discardCardsLocked` (`discard.go`), so `EventDiscardCard` fires, the CR 614 window opens over the discard (`RepEventDiscard`), and discard payoffs see it.

2. **Nothing honours a redirected entry.** "If you don't, put it into its owner's graveyard" is a `Replace` that rewrites `ev.NewZone`. No catalog card does that to an entry today, and every entry finisher is wrong about it in one of two ways:
   - `enterBattlefieldThroughPipelineLocked` (`entry_tail.go`), the resume branch of `applyResolvedReplacementEventLocked` (`pending_choice.go`), and the batch's `recordEntryBatchMemberLocked` (`entry_batch.go`) treat a redirect as a *cancel*: "there is no generic 'put it wherever the pipeline said' helper for these sources". The card stays in its old zone. For a cast Mox Diamond that resumed after its prompt, the old zone is the **stack**, and its `StackMeta` entry is already gone (`resolveTopOfStackLocked` deletes it before the entry). The card would be stranded there.
   - The **stack-resolution** site (`resolveTopOfStackLocked`, `mutations.go`) and the **land-play** site (`CastSpell`'s land branch) check only `out.Canceled`, then call `executeEntryToBattlefieldLocked(out)`, which moves the card to the battlefield **whatever `NewZone` says**. A cast Mox Diamond whose question was answered inline (no pause) would enter the battlefield after being told to go to the graveyard. This is a latent bug: no card has reached it yet.

3. **An un-asked inverted question takes the stronger branch.** `skipQuestionsLocked` and the single-effect `mustSettleNow` branch of `applyReplacementsLocked` (`replacements.go`) mark every effect that `asksItsOwnQuestion` as applied **without running `Replace`**. For an `Optional` "may", not applying is the weaker branch. For `EntryLifeCost` and `EntryHandReveal` it is the opposite: their `Replace` *is* the decline, so skipping it gives the shockland and the reveal-land an untapped entry for free. For Mox Diamond it would be a free Mox. It is unreachable today except in one corner (a CR 616 ordering window whose affected player has left the game), but Mox Diamond turns a small inconsistency into a card that is stronger than printed.

4. **The chosen-order loop cannot ask a hand question.** `ResolveReplacementOrder`'s loop (`pending_choice.go`) has branches for `CopySelector`, `EntryLifeCost` and `Optional`, and none for `EntryHandReveal`. An effect with a hand choice ordered alongside another entry replacement fires its `Replace` blind. For a reveal-land beside a Kismet that is merely unasked; for an opponent's Mox Diamond beside a Kismet it would bin the Mox without asking its controller anything. #847 fixed the same drift for `Optional`.

5. **The premise to correct.** The issue says "`ev.mustSettleNow` windows (a spell putting it onto the battlefield) need a declared answer". No entry sets `mustSettleNow` today: `counter_tail.go`, `life_tail.go` and `produce_mana.go` are the only setters. Since #1322 the "put onto the battlefield" batch is resumable, and the one entry that still cannot pause is the sandbox `move_card` verb. The question still needs a declared answer, for `move_card` and for the chooser-gone window. Decision 5 gives one.

---

## The rules

All numbers checked against the Comprehensive Rules effective August 7, 2026.

- **CR 614.1a.** An effect that uses "instead" is a replacement effect. Mox Diamond's first sentence is one. It is a static ability of the card itself, and **CR 113.6h** makes an ability that modifies how its object enters the battlefield function "as that object is entering the battlefield". So it applies **from every zone**: cast (from the stack), put from a hand, fetched from a library, reanimated from a graveyard, or returned from exile. The 2004 ruling on the same-shaped Heart of Yavimaya says so outright: "no matter how it is put onto the battlefield".
- **CR 614.12 and 614.12a.** Some replacement effects modify how a permanent enters, and may come from the permanent itself. A choice they require "is made before the permanent enters the battlefield."
- **CR 614.13, 614.13a and 614.13b.** "An effect that modifies how a permanent enters the battlefield may cause other objects to change zones." The discarded land is such an object. You "can't choose the object that will become that permanent or any other object entering the battlefield at the same time", and "the same object can't be chosen to change zones more than once" across replacements on one simultaneous entry.
- **CR 614.12b.** When several permanents enter at once, a player may not make choices whose combined costs would not be payable.
- **CR 614.6.** A replaced event never happens; the modified event occurs instead. The only ruling on Mox Diamond (2008-05-01) spells it out: "If you don't discard a land card, Mox Diamond never enters. It won't trigger abilities that look for something entering, and you won't get the opportunity to tap it for mana." Lotus Vale and Scorched Ruins have the same ruling: "it goes directly to your graveyard."
- **CR 616.1, 616.1b, 616.1f and 616.2.** With several replacements on one event the affected player orders them; one that changes under whose control the permanent enters must be chosen first; the process repeats after each; and one effect can make another applicable. So after "put it into its owner's graveyard", an "exile it instead" (Rest in Peace) applies to the same event, and "enters tapped" (Kismet) stops applying, because the event is no longer an entry.
- **CR 701.9a and 701.9b.** "To discard a card, move it from its owner's hand to that player's graveyard." The player chooses which card, by default. (The engine's comments still say 701.8a; the June 2025 edition moved discard to 701.9. The code comments are outside this ADR.)
- **CR 118.12** makes the "[do something]. If you do" action a *cost*, but only for "spells, activated abilities, and triggered abilities". Mox Diamond's is a static ability, so 118.12 does not reach it. Library of Leng's 2004-10-04 ruling draws the line the engine's `DiscardCause` draws: its replacement "applies any time a spell or ability has you discard as part of its effect", and not to a discard "as a cost, because costs aren't effects". Mox Diamond's static ability has you discard as part of its effect. So the cause is **effect**, not cost (Decision 3).
- **CR 608.3e.** A permanent spell whose controller can't put it onto the battlefield goes to its owner's graveyard. Mox Diamond's graveyard branch does not rely on this: the card says where it goes.
- **CR 903.9a and 903.9b.** In the current edition a commander put into a **graveyard** or exile may be moved to the command zone as a *state-based action*; only a hand or library destination is a replacement. The engine models the graveyard and exile half as the replacement `commanderZoneReplacement` (`builtin_replacements.go`) as well, and this ADR does not revisit that. For Mox Diamond it is almost moot: the dump has no commander-legal card whose front face is a land, so a land card in hand is never a commander in a legal game. The path must still not wedge if one is (Decision 3).
- **CR 400.7.** A card moving from the stack, a hand, a library or exile to the graveyard is a new object. A Mox Diamond reanimated without a discard goes from the graveyard "into its owner's graveyard". No rule covers a move to the zone the card is already in (CR 400.8 and 400.10 cover exile and the command zone only). The engine's sandbox mover already treats "a replacement rewrote the destination back to where the card already is" as nothing to do (`mutations.go`), and Decision 4 takes the same reading.
- **CR 116.2a and 305.2.** Playing a land is a special action, and it is the play that uses the land drop. A land whose played entry is redirected (Heart of Yavimaya with no Forest to sacrifice) has still been played.

### The same shape elsewhere

A sweep of the Scryfall dump for "would enter … instead", "as … enters, discard / sacrifice / exile" and "enters … unless you …" found three families.

| Family | Cards | The same seam? |
|---|---|---|
| **614.1a "would enter … instead. If you do, put it onto the battlefield. If you don't, put it into its owner's graveyard"** | Mox Diamond (discard a land card, "you may"); Heart of Yavimaya, Kjeldoran Outpost, Lake of the Dead (sacrifice a basic-typed land); Balduvian Trading Post, Soldevi Excavations (sacrifice an untapped one); Lotus Vale, Scorched Ruins (sacrifice two untapped lands) | **Yes.** The choice, the decline-redirects-to-the-graveyard branch and every finisher are the same. Only the action and its zone differ: a hand discard for the Mox, a battlefield sacrifice for the seven lands. |
| **614.13 "as this enters, discard / sacrifice / exile any number …", no redirect** | Indominus Rex, Alpha (discard creature cards, then counters for their keywords); Shimatsu the Bloodcloaked (sacrifice); Mimeoplasm, Revered One (exile from graveyard); Sheltered Valley (mandatory sacrifice, no graveyard branch) | **Partly.** The in-window action is the same; there is no decline redirect, and the follow-on reads the cards the action moved. Indominus Rex is a hand discard and could ride Decision 1 if its `Then` could write the entering event. |
| **603 "when this enters, sacrifice it unless you discard / sacrifice / return …"** | Fallow Wurm, Thundering Wurm, Hidden Horror, Mercenary Knight, Drekavac, Avatar of Discord, Body Snatcher, the random-discard Hordes; the Karoo bounce lands and Lairs; champion | **No.** These are ETB *triggers* (#578's distinction): the creature enters, triggers, and can be responded to. The "unless" action is a CR 118.12 cost paid at resolution, so its discard is `DiscardCauseCost`. A different seam. |

Chrome Mox is not in the table. It prints "When this artifact enters", so its imprint is an ETB trigger and ships complete already (`chrome_mox.go`).

---

## Options

**A. Generalise `EntryHandReveal` into a hand choice with an action (recommended).** One declaration, `EntryHandChoice`, gains `Action` (`reveal` or `discard`). The offer, its three inline guards, the choose-cards payload and the resolver are shared; the action decides what happens to the picked cards and which prompt kind is queued. The ten reveal-lands keep `reveal`, the zero value.

**B. A sibling shape, `EntryDiscard`.** A fifth branch on the apply-loop with its own offer, prompt and resolver. It reads well at the declaration site, but it copies every guard and the resolver from `entry_reveal.go`. That is the pair §5z warned drifts, and it would need its own arm in the ordering loop and the skip rule — the two places gaps 3 and 4 above already show drifting.

**C. An ETB trigger** ("when this enters, sacrifice it unless you discard a land card"). Rejected. The ruling says the Mox never enters without the discard, so it must not trigger ETB watchers or be tappable in response. That would be observable and stronger than printed.

**D. An additional cost to cast.** Rejected. CR 113.6h applies the ability however the Mox enters, and a countered Mox Diamond would already have taken the land.

**E. A general "entry action" over any zone** — hand discard, battlefield sacrifice, graveyard exile. That covers every row of the first two families at once. It is the right destination, but it is three prompt kinds and two new bot valuations for one waiting card. Option A is written so E is a later widening (`From` beside `Action`) rather than a rewrite. See open question 2.

---

## Decisions (recommended)

### 1. `EntryHandChoice` with an `Action`

`ReplacementEffect.EntryHandReveal *EntryHandReveal` becomes `EntryHandChoice *EntryHandChoice`:

```go
type EntryHandAction string

const (
	EntryHandReveal  EntryHandAction = ""        // CR 701.20 — the ten reveal-lands (zero value)
	EntryHandDiscard EntryHandAction = "discard" // CR 701.9 — Mox Diamond
)

type EntryHandChoice struct {
	Action   EntryHandAction
	Matches  func(c Card) bool // printed characteristics of a card in hand; unchanged rules
	Min, Max int
	Question string
	Then     func(g *Game, picked []uuid.UUID) error
}
```

The contract is §5z's. `Replace` is the **decline**. Answering with a card skips `Replace` and performs the action. The three inline guards (no resume, no matching card in hand, the chooser has left) apply `Replace`. The catalog side is one constructor beside `EntersTappedUnlessYouRevealFromHand`:

```go
// "If this <permanent> would enter, you may discard <clause> instead.
//  If you do, put it onto the battlefield. If you don't, put it into
//  its owner's graveyard."
func EntersOnlyIfYouDiscardFromHand(name, clause string, matches func(game.Card) bool) game.ReplacementEffect
```

Its `Replace` sets `ev.NewZone = ZoneGraveyard` and `ev.NewZoneOwner` to the card's owner. Mox Diamond is then one `Register` with that clause and its existing mana-ability shape (`{W|U|B|R|G}`), `CompletenessFull`.

The rename moves three lines of `server/internal/game/testdata/closure_fields.txt` (`EntryHandReveal.Matches`, `.Then`, `ReplacementEffect.EntryHandReveal`) to their new names in the same class, `census:ChoiceResumeFrames`. The count does not change, so no ceiling moves. `Action` is a string and adds no line.

### 2. One more prompt kind, the fourth card-set pick

`PendingChoiceEntryDiscardFromHand` (`"entry_discard_from_hand"`) joins `isCardSetPickKind`. It carries the same payload as the reveal kind: `ChooseCards`, `ChooseMin`/`ChooseMax`, `chooseCardsFrame{zone: ZoneHand}`, and the `replacementResume` frame. One resolver serves both kinds (`ResolveEntryHandChoice`; `ResolveEntryRevealFromHand` stays as a wrapper if the actions dispatch wants it).

It is a separate *kind* rather than a flag on the reveal kind for §5z's own reason, the third one: **the sign is inverted for a bot**. A revealed card is kept, so any reveal beats none. A discarded card is spent. The client's sentence also differs ("Discard a land card?" against "Reveal an Island or Swamp card?"), and the kind is what the client renders from.

Everything a kind owes in the engine follows §5z's rows. `choiceGateDecisions` blocks the table. `choiceDepartureDecisions` gets `{}`, because the paused entry rides `replacementResume` and `finishDroppedReplacementLocked` settles it. The view redacts it from every other seat as it redacts the reveal kind (`protocol/view.go`), since what is in a hand, and how many cards match, is hidden. `actions.go` routes `{card_ids}` to the resolver.

### 3. The discard is an effect's, through the one discard path, and it may pause

When the answer names a card:

1. The effect is marked applied for this event (CR 614.5), as the reveal path does.
2. The card goes through `discardCardsLocked` with `discardOptions{cause: DiscardCauseEffect, source: <the entering card>}`. So `RepEventDiscard` opens, Library of Leng may offer its top-of-library "may", Rest in Peace exiles the land, and `EventDiscardCard` fires with the Mox as `Source` and cause `effect` (CR 701.9a; Library of Leng's ruling above).
3. "If you do" is `discardOptions.then`'s `landed` list: `discardedThisWayLocked` (#1027), meaning the card left the hand, wherever the window sent it. Non-empty: the Mox's own `Replace` is not run and it will enter. Empty (a replacement cancelled the discard outright, or the prompt was taken away): `Replace` runs and the Mox is redirected. This is the engine's existing reading of "this way", and the weaker one.
4. Then the entry resumes exactly as the reveal path resumes it: re-enter `applyReplacementsLocked(ev)` so CR 616.1f picks up anything newly applicable, and settle through `finishSettledReplacementLocked`.

**Why it may pause.** The discard is not a cost, so it does not set `MustSettleNow`. Two things can stop it: Library of Leng's `Optional` question, and a CR 616 ordering between two discard replacements (Leng and Rest in Peace). A discarded commander is the third in principle. The entry is already paused and resumable (that is why the question was asked), so a second pause costs one action and nothing else.

**How the entry rides across the discard's pause.** The discard's `then` must re-enter the paused entry event. It does not hold a pointer to the live `*ReplacementEvent`: an undo into Library of Leng's prompt would then share it with the snapshot, and the live game's run would mutate what the undone branch replays. §5o's `cloneReplacementResume` exists for exactly this. So the continuation captures a **frozen copy** of the event, taken with the same clone the resume frames use, and clones it again each time it runs. That is undo-safe by construction and adds **no struct field**. The closure rides the existing `zoneRoute.then` route of the closure ratchet, so no new line appears and no ceiling moves. (A `zoneRoute.resumeEntry` field was the obvious alternative. The closure ratchet forbids raising the `census:ChoiceResumeFrames` ceiling, so it is not available.)

The smaller alternative is a discard that settles now: `MustSettleNow` on this one discard. It never pauses, but Library of Leng's "may" is then skipped un-applied under the existing `mustSettleNow` rule. See open question 3.

### 4. A redirected entry is honoured, by one finisher

`finishRedirectedEntryLocked(ev)` is the terminal outcome for a settled entry whose `NewZone` is not the battlefield. It:

- converts the settled event into a settled exit. It attaches `zoneRoute{CardID, Dst: ev.NewZone, DstOwner: ev.NewZoneOwner, Actor: ev.Actor}` and calls `executeZoneRouteLocked(ev)`, the mover every routed exit already uses. It moves from whatever zone the card is in, retires a stack record by source zone (§5ad), and emits the zone move a "put into a graveyard from anywhere" watcher sees. It does **not** open a second window: CR 616.2 already applied Rest in Peace and CR 903.9 to this event while it was being replaced, and running them again would ask the commander question twice.
- is a no-op when the settled destination is the zone the card is already in (a reanimated Mox Diamond, CR 400.7 above);
- runs the entry tail with `uuid.Nil` ("nothing entered"), so a search, a batch or a caller's `Then` is told.

All five entry sites call it:

| Site | Today | After |
|---|---|---|
| `enterBattlefieldThroughPipelineLocked` (search, exile return, reanimation, token) | redirect = cancel | the finisher |
| `applyResolvedReplacementEventLocked`, entry resume branch | redirect = cancel | the finisher |
| `resolveTopOfStackLocked` (a resolving permanent spell) | **enters the battlefield anyway** | the finisher |
| `CastSpell` land branch | **enters the battlefield anyway** | the finisher, and the land drop is still spent (CR 116.2a / 305.2) |
| `recordEntryBatchMemberLocked` / the batch landing | redirect = not entering | a redirected member is moved by the finisher when the batch lands, before the members that enter are announced |

A created token is the one exception. A token redirected to another zone would cease to exist there (CR 111.7, 704.5d), and it was never in a zone to begin with, so it keeps its current behaviour: `dropEnteringTokenLocked` discards it.

The sandbox `move_card` verb already honours a rewritten destination, so it needs nothing.

### 5. When the question cannot be asked, the decline applies

`asksItsOwnQuestion` answers "may this effect be fired blind?" There is a second question it cannot answer: "which branch is the weaker one?" A new predicate answers it:

```go
// declineIsReplace reports that this effect's Replace IS the
// "you didn't" branch — the shockland's enters-tapped, the
// reveal-land's enters-tapped, Mox Diamond's graveyard.
func declineIsReplace(e ReplacementEffect) bool {
	return e.EntryLifeCost > 0 || e.EntryHandChoice != nil
}
```

`skipQuestionsLocked` and the single-effect `mustSettleNow` branch of `applyReplacementsLocked` **run `Replace`** for such an effect instead of marking it applied un-run. `Optional` and `CopySelector` keep the skip, which is already their weaker branch. So in every window that cannot put the question to anybody, the Mox goes to its owner's graveyard, and a shockland or reveal-land enters tapped. Today those windows are the sandbox `move_card` verb, whose entry is not resumable, and an ordering window whose affected player has left. The fix is one predicate and two call sites, and it closes gap 3 for the two land families as well.

### 6. One dispatcher for "an effect that asks its own question"

The apply-loop's single-applicable arm and `ResolveReplacementOrder`'s chosen-order loop each keep their own list of question branches. The second list is already missing `EntryHandReveal` (gap 4). Both call one helper:

```go
// offerOwnQuestionLocked queues the question an effect asks, or applies
// its un-asked branch inline. Reports whether a prompt is now pending.
func (g *Game) offerOwnQuestionLocked(ev *ReplacementEvent, chosen activeReplacement) bool
```

It dispatches `CopySelector`, `EntryLifeCost`, `EntryHandChoice` and `Optional` in that order. After that, adding a question kind means one branch, and the two loops cannot drift again (the #847 lesson).

### 7. CR 614.13 in a simultaneous entry

In the resumable batch (`entry_batch.go`), each member's window is asked in turn, and a hand discard happens when its answer arrives:

- **CR 614.13a.** The candidate list excludes every card that is itself a member of the same batch. A Mox Diamond and a land put from a hand together cannot discard that land.
- **CR 614.13b and 614.12b** hold by construction: a land discarded for one member has left the hand before the next member's question is built.
- **Declared simplification.** In paper every choice is made before the event, so a land could be revealed for one member and discarded for another. Here a card discarded for an earlier member cannot be revealed for a later one. This is weaker, never stronger, and needs a reveal-land and a Mox Diamond put from the same hand at once.

### 8. What the bot and the enumerator need

- **Enumerator** (`legal/choices.go`): the new kind joins the card-set arm beside `entry_reveal_from_hand`, with the verb `": discard"`. The floor is zero, so "discard nothing" is the `AlwaysLegal` answer, offered first. `ChooseCardsPickLegalLocked` is the only validity check.
- **Heuristic** (`aiseat/heuristic/choices.go`): a new case with the opposite sign to the reveal case. Naming a card spends it. The seat prices each candidate with `fuelValue` (#1028, the pricer for a card a cost eats) and discards the cheapest land. It declines only when that land is worth more than the permanent: for Mox Diamond, when the land is the seat's only land card in hand and it has not yet made its land drop this turn. This is a first cut; the position suite can label it.
- **The bot reads nothing new about the catalog.** The prompt, its candidates and its bounds come through the filtered view, as §5z's did. `aiseat/` still imports nothing from `internal/game`.

### 9. What the client prompt needs

`ChoicePromptModal.svelte` renders the new kind next to the reveal branch. The header is the card's sentence (`Question`, falling back to the effect's `PromptQuestion`: "Mox Diamond — discard a land card so it enters?"). A single-select grid shows the matching hand cards, with two actions: **Discard** (enabled once a card is picked) and **Don't discard — put Mox Diamond into its owner's graveyard**. The decline names its consequence, because it destroys the card. `protocol.ts` gains the kind string. No other wire change: the payload is the reveal kind's.

### 10. Snapshot impact: none

- No new field on `Card`, `Game`, `ReplacementEvent` or `zoneRoute`. The shape guard (`TestSnapshotShapeIsRecorded`) does not move, and `SnapshotSchemaVersion` does not bump.
- A table with either prompt open, or with the discard paused inside one, holds resume frames and so writes no restore point. That is true of every continuation-bearing prompt today, counted in `ContinuationCensus.ChoiceResumeFrames`. The new kind is a new *value* of the persisted `kind` string, never written in a restorable snapshot.
- Closure ratchet: three renamed lines, same class and same count (Decision 1). No new route (Decision 3).
- The fixture corpus is untouched. No existing restore point can hold a Mox Diamond or an entry prompt.

---

## Cards

**Covered by this ADR:** Mox Diamond. It ships `CompletenessFull`.

**Made one constructor away, not in scope** (open question 2):

- The seven "sacrifice … instead" lands: Heart of Yavimaya, Kjeldoran Outpost, Lake of the Dead, Balduvian Trading Post, Soldevi Excavations, Lotus Vale and Scorched Ruins. They need Decisions 4–6 unchanged, plus a battlefield-sacrifice action (`From: ZoneBattlefield`, the `own_permanents` pick shape, `SacrificeAllThenForEffect` for the move), and "untapped" for three of them.
- Indominus Rex, Alpha: a hand discard with no redirect. It needs `Then` to be handed the entering event so it can add counters (`ev.AddCounterAtETB`).

**Not covered:** the ETB "sacrifice it unless you …" triggers (a CR 118.12 cost at resolution, a separate seam); champion; Shimatsu, Mimeoplasm and Sheltered Valley (other zones, no redirect).

---

## Consequences

- Mox Diamond ships complete, and the roadmap row "Discarding a card as a permanent would enter" (`roadmap/registry.go`, `discard-as-it-would-enter`) closes with a fragment in `docs/engine-seams/closed/`.
- Two latent bugs close before a card can reach them: a redirected spell or land entry landing on the battlefield anyway, and an un-asked shockland or reveal-land entering untapped.
- The two question loops share one dispatcher.
- One more card-set-pick kind means four readers of `isCardSetPickKind`, which is the direction §5z chose over a copy of the validation.
- Out of scope, and noticed on the way: `resolveTopOfStackLocked` leaves a permanent spell whose entry is *cancelled* on the stack ("canceled permanents stay on the stack"). CR 608.3e would put it into its owner's graveyard. No catalog card cancels an entry, so it is recorded here rather than fixed.

---

## Open questions for the owner

1. **Is the recommended shape right?** That is, `EntryHandReveal` renamed to `EntryHandChoice` with an `Action`, and a second prompt kind (Option A). Or do you prefer a separate sibling declaration (Option B), accepting the duplicated guards and resolver?

2. **Scope: Mox Diamond alone, or the whole "would enter … instead" family?** The recommendation builds the shared machinery (the redirect finisher, the decline rule, the one dispatcher) and ships Mox Diamond only. The seven sacrifice lands would follow in their own PR with a battlefield-sacrifice action (Option E's first widening). Should they be in this change instead? And should Indominus Rex, Alpha ride along by handing `Then` the entering event?

3. **May the discard pause?** The recommendation lets it pause (effect cause, no `MustSettleNow`), so Library of Leng's "may" and a Leng-plus-Rest-in-Peace ordering are asked, with the paused entry carried in a frozen copy. The alternative settles the discard at once: simpler, but Leng's "may" is skipped (weaker). Which do you want?

4. **Is the discard an effect's (`DiscardCauseEffect`) or a cost's (`DiscardCauseCost`)?** The recommendation is effect: CR 118.12 covers only spells, activated abilities and triggered abilities, and Library of Leng's ruling applies to "any time a spell or ability has you discard as part of its effect". The practical difference is whether Library of Leng can put the land back on top of the library.

5. **What counts as "if you do"?** Whether the card left the hand (`discardedThisWayLocked`, recommended, weaker), or whether the player *chose* to discard (the CR 118.12 reading for spells and abilities, under which a discard a replacement cancelled would still let the Mox enter)? No catalog card cancels a discard today.

6. **Should the sandbox `move_card` verb bin a Mox Diamond?** It cannot pause, so under Decision 5 a Mox Diamond dragged onto the battlefield by hand goes to its owner's graveyard, just as a reveal-land dragged there enters tapped today. The alternative is that a manual move skips replacements that ask questions entirely. That would be a sandbox exception, and it would also change the reveal-lands and shocklands.

7. **Is a land-play redirect a spent land drop?** Not reached by Mox Diamond. Heart of Yavimaya with no Forest is played and goes to the graveyard. The recommendation, from CR 116.2a, is that the land drop is spent, because the play happened. Confirm, so the finisher's land-play arm is settled once.
