# ADR 0114 — The Ring tempts you

**Status:** Proposed · 2026-10-03 · S58 — Deck requests, October batch (tracker [#2077](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2077))
**Owner decisions:** pending. The four questions are under [Open questions for the owner](#open-questions-for-the-owner). Nothing is built until they are answered.
**Issues:** [#2076](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2076) (the seam), [#2062](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2062) (Night - Sauron The Slayer, the deck that asked for it).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-03. I ran `git fetch --all --prune` and listed `docs/decisions/` on all 36 remote heads (`origin/develop`, `origin/main`, `origin/feat/s58-pr1-stamps-mdfc` and the other chore, docs, feat, fix, repro and wip branches). The highest number on any of them is 0112. `docs/s58-small-seams-adr` is being written at the same time and takes the first free number, 0113, so this ADR takes **0114**.
**Builds on:** [ADR 0064](0064-emblems.md) (emblems), [ADR 0071](0071-designations-that-switch-abilities-on.md) (designations and the `ActiveWhen` gate), [ADR 0045](0045-combat-restrictions.md) addendum (block rules), [ADR 0087](0087-amass.md) (amass and its forced-answer shortcut), [ADR 0060](0060-leaving-the-game.md) (CR 800.4 and the choice departure table), [ADR 0041](0041-game-persistence.md) (snapshots), [ADR 0033](0033-ai-bot-seat.md) §1 (the legal-move enumerator) and [ADR 0106](0106-five-small-seams-from-the-s50-rechecks.md) owner decision 6 (a PR lands every card its seam unblocks).

This ADR was written plan-first. No engine, client or card code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

The deck in #2062 is led by Sauron, the Dark Lord, whose last two abilities are "Whenever an Army you control deals combat damage to a player, the Ring tempts you" and "Whenever the Ring tempts you, you may discard your hand. If you do, draw four cards." Four more of its cards need the same keyword action: Call of the Ring, Nazgûl (nine copies), Ringsight and Sauron, Lord of the Rings. The engine has no Ring, no Ring-bearer and no "the Ring tempts you".

### The rules

Every number below was read in the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`). Rulings are the Scryfall rulings of 2023-06-16, which every Ring card repeats.

- **CR 701.54a.** "Each time the Ring tempts you, choose a creature you control. That creature becomes your Ring-bearer until another creature becomes your Ring-bearer or another player gains control of it."
- **CR 701.54b.** "Ring-bearer is a designation a permanent can have. Being a Ring-bearer is not a copiable value."
- **CR 701.54c.** If a player has no emblem named The Ring when the Ring tempts them, they get one "before choosing a creature to be their Ring-bearer". The emblem has "Your Ring-bearer is legendary and can't be blocked by creatures with greater power." As long as the Ring has tempted that player two or more times it also has "Whenever your Ring-bearer attacks, draw a card, then discard a card"; three or more, "Whenever your Ring-bearer becomes blocked by a creature, the blocking creature's controller sacrifices it at end of combat"; four or more, "Whenever your Ring-bearer deals combat damage to a player, each opponent loses 3 life."
- **CR 701.54d.** "The Ring tempts a player whenever they complete the actions in 701.54a, even if some or all of those actions were impossible."
- **CR 701.54e.** A creature "is your Ring-bearer" if it "is on the battlefield under your control and has the Ring-bearer designation."
- **The rulings.** "Each time the Ring tempts you, you must choose a creature if you control one." "Each player can have only one emblem named The Ring and only one Ring-bearer at a time." "The Ring gains its abilities in order from top to bottom. Once it gains an ability, it has that ability for the rest of the game." "If the creature you choose as your Ring-bearer was already your Ring-bearer, that still counts as choosing that creature" (Call of the Ring triggers again). "If the Ring tempts you but you can't choose a creature … Call of the Ring's last ability won't trigger", while "whenever the Ring tempts you" abilities still do. The reminder card orders the steps: get the emblem if you have none, then the emblem gains its next ability, then you choose.

The surrounding rules:

- **Emblems.** CR 114.2 to 114.5: an emblem is in its owner's command zone, is neither a card nor a permanent, and its abilities function there (also CR 113.6p). The source of the emblem's triggered abilities is the emblem (CR 113.7).
- **Legendary.** "Is legendary" adds a supertype, a layer 4 effect (CR 613.1d). A legendary permanent is subject to the legend rule (CR 205.4d, 704.5j), which groups by controller and name.
- **The evasion.** "Can't be blocked by creatures with greater power" is a CR 509.1b restriction, checked as blockers are declared.
- **Becomes blocked by a creature.** CR 509.3d: it triggers once for each creature that blocks the Ring-bearer, including one put onto the battlefield blocking it.
- **At end of combat.** The level 3 ability creates a delayed triggered ability (CR 603.7) that triggers as the end of combat step begins (CR 511.2). The creature's controller is read when it resolves.
- **Control.** A control change ends the designation (CR 701.54a) and removes the creature from combat (CR 506.4).
- **Zones and phasing.** A Ring-bearer that leaves the battlefield is a new object with no designation (CR 400.7). Phasing is not a zone or control change (CR 702.26d), so a phased-out Ring-bearer keeps the designation, but while phased out it is treated as though it does not exist (CR 702.26b), so it is nobody's Ring-bearer for CR 701.54e.
- **Multiplayer.** Every player has their own emblem and their own Ring-bearer. "Each opponent" in the level 4 ability is every opponent still in the game, not only the one dealt damage. A player who leaves the game takes their emblem with them (CR 800.4a). A prompt they owe is handled by CR 800.4g and 800.4h.
- **Commander.** A commander can be a Ring-bearer, and is already legendary. CR 903.9a makes a commander that goes to a graveyard or exile a state-based-action choice *after* the move. That matters for Sauron, Lord of the Rings ([question 4](#4-sauron-lord-of-the-rings-and-a-commander-that-dies)).

### What exists, what is missing

| Piece | On `origin/develop` at `f6444416` | Missing |
|---|---|---|
| Emblems | `Player.Emblems`, a second command-zone slice; `CreateEmblemForEffect` (`game/emblem.go`) files the emblem under `"emblem:" + <the creating card's key>`. Five walks read emblems: the layer pass (`emblemContinuousEffectsLocked`), the trigger harvest (`harvestFromEmblemsLocked`), the rule gates (`forEachEmblemLocked`), the untap and draw steps (`untap.go`, `draw_step.go`) and activation timing. Wire: `PlayerView.emblems[]` (`protocol/view.go:1507`), drawn as chips by `PlayerIdentity.svelte:493`. | An emblem no card creates. The Ring is a rules object (CR 701.54c), so it has no creating card to key it on. |
| A gate that switches abilities on | `Designation` and `ActiveWhen` (`game/designations.go:154`), read only through `StaticAbilitiesForCard`, `TriggersForCard` and the other accessors. The emblem walks already go through them (`emblem.go:226`). `Active(c Card)` reads the object only. | A designation kind for "the Ring has tempted that player N or more times". |
| A designation on a permanent | `Card.Monstrous`, `Card.Harnessed` (`game/card.go`), cleared at both CR 400.7 sites (`zone.go:288`, `entry_tail.go:324`), carried by the snapshot (`snapshot.go:820`), sent as `CardView.monstrous` (`view.go:2240`) and drawn in `Card.svelte`'s one designation-badge slot (`Card.svelte:385`). | A designation that also ends on a control change. The control step is `materialiseControlLocked` (`game/layers.go:1004`). |
| Block rules | `BlockRule.Pair` and its constructors (`game/block_rules.go`, `cards/effects/block_rules.go`); Locke, Treasure Hunter already has "can't be blocked by creatures with greater power" (`locke_treasure_hunter.go`). | `forEachBlockRuleLocked` (`block_rules.go:150`) walks the battlefield and the scoped records. It does not walk emblems. |
| Combat triggers | `EventAttack` (`events.go:505`), `EventBlock` per blocker–attacker pair (`events.go:944`, Lim-Dûl's Cohort's shape), combat `EventDealDamage`; `ScheduleDelayedTrigger{At: game.StepEndCombat}` (`disturb_batch_c_helpers.go:52`). | Nothing. |
| A resolution-time "choose a creature you control" | Amass asks it with `QueueChooseCardsForEffect`, Min 1 / Max 1, and skips the prompt when there is only one Army (`game/amass.go:244`). `own_permanents` (`game/resolution_pick.go:137`) is the same question as its own kind. | A kind whose bot ordering points the right way. `own_permanents` ranks a seat's own permanents by what it would miss least (`Options.OrderCostFuel`), which is the opposite of picking a Ring-bearer. |
| Deck construction | `deck.Validate` refuses a second copy of any non-basic card (`deck/validate.go:154`), and the upload path treats that as fatal (`lobby/http.go:2573`). | CR 113.6n. #2062's list has nine Nazgûl, so the deck cannot be uploaded at all today. |
| Commander deaths | CR 903.9 is the built-in replacement `commanderZoneReplacement` (`game/builtin_replacements.go:33`), applied to graveyard and exile as well as hand and library. A commander whose owner takes the command zone never reaches the graveyard, so it never dies. | The pinned CR 903.9a. See question 4. |
| The bot | `internal/legal` answers every card-set pick kind (`legal/choices.go:591`). The model prompt (`aiseat/model`) renders no emblem. | An ordering for the new pick, and the Ring in the model's board text. |

Nothing in the tree mentions the Ring today except a test comment about The One Ring. No registry row covers it (`roadmap/registry.go` has no "tempt").

### Sizing

Scryfall's `o:"the ring tempts you"` returns 54 cards. Two more mention a Ring-bearer without tempting: Lord of the Nazgûl ("protection from Ring-bearers") and Sauron, the Necromancer ("unless Sauron is your Ring-bearer"). None of the 56 is in the catalog (compared against `server/internal/cards/coverage/testdata/oracle/` at `f6444416`).

I read each one's full text and asked whether everything except the Ring has a shape in the engine. As in ADR 0108, ✓ is a card I found every other clause for and ? is one I was unsure of. It is a planning number; each PR verifies every card again.

- **✓ (38):** Bilbo, Retired Burglar; Birthday Escape; Bombadil's Song; Breaking of the Fellowship; Call of the Ring; Claim the Precious; Dreadful as the Storm; Dúnedain Rangers; Enraged Huorn; Faramir, Field Commander; Fiery Inscription; Galadriel of Lothlórien; Glorious Gale; Gollum, Patient Plotter; Gollum's Bite; Horses of the Bruinen; Inherited Envelope; Mirrormere Guardian; Nazgûl; Now for Wrath, Now for Ruin!; One Ring to Rule Them All; Ranger's Firebrand; Relentless Rohirrim; Ringsight; Ringwraiths; Rohirrim Lancer; Sam's Desperate Rescue; Sauron, the Dark Lord; Shortcut to Mushrooms; Slip On the Ring; Sméagol, Helpful Guide; Soothing of Sméagol; Stalwarts of Osgiliath; The Black Breath; The Ring Goes South; Took Reaper; Uruk-hai Berserker; War of the Last Alliance.
- **? (16):** Aragorn, Company Leader (a counter kind of your choice, and "whenever you put one or more counters on Aragorn"); Boromir, Warden of the Tower ("if no mana was spent to cast it"); Elrond, Lord of Rivendell ("the second time this ability has resolved this turn"); Frodo, Adventurous Hobbit (partner with); Frodo Baggins (a conditional "must be blocked if able"); Frodo, Sauron's Bane (becomes a new creature type with a granted ability that ends the game); Gandalf, Friend of the Shire (sorceries as though they had flash); In the Darkness Bind Them (one theft per opponent); Rangers of Ithilien; Samwise the Stouthearted ("put there from the battlefield this turn"); Sauron's Ransom (a face-down and a face-up pile); Scroll of Isildur; There and Back Again; Witch-king of Angmar; Lord of the Nazgûl (protection from a designation); Sauron, the Necromancer.
- **Not alone (2):** Galadriel, Elven-Queen needs voting (Will of the council, an open seam on the S58 tracker). Sauron, Lord of the Rings is question 4.

About 46 of 56.

---

## Decision

### 1. The tempt is one keyword action

`(*Game).RingTemptsForEffect(player uuid.UUID, then func(*Game, uuid.UUID) error)` is the whole of CR 701.54a and 701.54c, in the reminder card's order:

1. **The emblem.** If the player has no Ring emblem, they get one (§2).
2. **The count.** The emblem's count goes up by one. The layer version is bumped, because a static or a trigger may just have switched on.
3. **The choice.** The candidates are the creatures the player controls that are on the battlefield and phased in. With none, nothing is chosen. With two or more, the player is asked (§4). With exactly one, the answer is forced, and question 2 decides whether it is still asked.
4. **The designation.** The chosen creature becomes the player's Ring-bearer (§3). The player's previous Ring-bearer stops being one, even if it is phased out. Choosing the creature that already is the Ring-bearer changes nothing on the board, but it still counts as choosing it.
5. **The event.** `EventRingTempted` is emitted with `Actor` the player, `CardID` the creature chosen (or `uuid.Nil`) and `Amount` the new count. This is the "complete the actions" of CR 701.54d, so it is emitted even when nothing could be chosen.
6. **The rest of the sentence.** `then` runs with the chosen creature. Ringsight's search and One Ring to Rule Them All's mill come after the tempt, and Galadriel, Elven-Queen puts a counter on "your Ring-bearer". This is amass's continuation shape (`amass.go`).

The card side is one primitive, `effects.TheRingTemptsYou{Then: …}`, whose player is the controller. Trigger constructors:

- `WheneverTheRingTemptsYou(label, effect)` watches `EventRingTempted` for the source's controller.
- `WheneverYouChooseARingBearer(label, effect)` is the same event with `CardID != uuid.Nil` (Call of the Ring and its rulings).
- `IfYouChoseAnotherRingBearer` is the intervening "if" (CR 603.4) for "if you chose a creature other than ~ as your Ring-bearer" (Aragorn, Faramir, Galadriel of Lothlórien, Gandalf).

The readers are `game.RingBearerOf(g, player)` (CR 701.54e: on the battlefield, phased in, controlled by that player, designated) and `game.RingTemptCount(g, player)` ("the Ring has tempted you N or more times this game", Frodo, Adventurous Hobbit and Frodo, Sauron's Bane).

There is no CR 614 window on the tempt. Nothing printed replaces one, and amass's reason for opening one (a "double the counters" replacement) has no Ring equivalent.

### 2. The Ring is an emblem with a fixed key

The Ring is an ordinary ADR 0064 emblem in `Player.Emblems`. Only its key differs. ADR 0064 derives the key from the card that makes the emblem, and no card makes this one, so it gets a fixed key, `game.RingEmblemKey = "emblem:the-ring"`. The key is an on-disk identity, like a token slug: never renamed, never reused.

- **Where it is declared.** `cards/effects/the_ring.go` files one emblem `CardDef` under that key, through a small `registerRulesEmblem` beside `Register`. It goes into `defs`, not `registry`, so the card census never counts it (ADR 0064 Decision 3).
- **Its four abilities** are written in the vocabulary every emblem uses. Each carries `ActiveWhen: game.RingTempted(n)`:
  - n = 1: a `Layer4Type` static that adds the Legendary supertype to the emblem owner's Ring-bearer, and a block rule (§5).
  - n = 2: `On(EventAttack, …)` for the owner's Ring-bearer. The effect is draw a card, then discard a card.
  - n = 3: `On(EventBlock, …)` with the owner's Ring-bearer as the attacker, once per blocker (CR 509.3d). The effect schedules a delayed trigger at the end of combat step for that blocker. Its registered body has the blocker's controller sacrifice it, if it is still the same object on the battlefield.
  - n = 4: `On(EventDealDamage, …)` for combat damage from the owner's Ring-bearer to a player. The effect is that each opponent of the emblem's owner loses 3 life.
- **The count.** The count lives on the emblem object, in a new field `Card.RingTemptations int`, and `DesignationRingTempted` reads it: `Active` is `c.RingTemptations >= d.N`. The count is a fact about the player, but the emblem is created at the first temptation and never leaves while its owner is in the game, so the count on the emblem *is* the player's count. That makes it one source of truth, read by the object-only gate ADR 0071 requires. `RingTemptCount` reads it, and answers 0 for a player with no emblem.
- **Why an emblem and not the monarch's listener.** CR 725.2 says the monarch's two triggers have no source, which is why `monarch.go` is a `Listener`. The Ring's abilities belong to the emblem (CR 701.54c, 113.7), so they are harvested, ordered, put on the stack and answerable like any emblem trigger. This also gives the wire, the persistence and the CR 800.4a exit for free.
- **What the emblem chip says.** `EmblemView` gains `level` (the count). Its `text` for the Ring is the lines the emblem has right now, derived on every projection, because the emblem has no other abilities (CR 114.1). Question 1 asks how the table shows the rest.

### 3. Ring-bearer is a designation on the permanent

`Card.RingBearer bool`, in the struct's bool block beside `Monstrous`:

- **Set** only by step 4 of §1, which first clears every other permanent the chooser controls. That one write site keeps "only one Ring-bearer at a time" per player.
- **Cleared** at both CR 400.7 sites (`zone.go`, `entry_tail.go`) and in `materialiseControlLocked`, for a control change (CR 701.54a). It is not cleared by phasing (CR 702.26d), by turning face down or up, or by ceasing to be a creature. CR 701.54a ends it for two reasons only, and the readers ask for a creature where the text does.
- **Not copiable.** `CopiableValuesOf` never reads it (CR 701.54b, 707.2), so a Clone of a Ring-bearer is not one.
- **Read** only through `RingBearerOf` and the emblem's scope helper, both of which apply CR 701.54e. A phased-out Ring-bearer is therefore nobody's, and it is back when it phases in, unless another creature was chosen meanwhile.
- **Legendary.** The emblem's layer 4 static makes the Ring-bearer legendary with the emblem's timestamp (CR 613.7c). `IsLegendary` reads the effective supertypes (`legend_rule.go:54`), so the legend rule, Ringsight's "a legendary creature you control" and Sauron, the Dark Lord's "Ward—Sacrifice a legendary artifact or legendary creature" all see it with no change. A Ring-bearer that shares a name with another legendary permanent its controller has goes to the legend rule (CR 704.5j). That is the printed outcome.

### 4. The choice: a `ring_bearer` prompt

A new pending-choice kind, `ring_bearer`, with the card-set payload: the candidates, Min 1, Max 1, and a re-check against the live battlefield on submit, as amass's prompt does. It is a kind of its own and not `own_permanents` or `choose_cards` for the reason `resolution_pick.go` gives for keeping those apart: a bot orders the answer by what the question is for. A Ring-bearer is the creature you most want to be legendary, evasive and looting, which is the opposite of "the permanent you would miss least".

The three rows docs/adding-cards.md requires for a new kind:

1. **It blocks the table** (`choiceGateDecisions`). It is asked mid-resolution, and the rest of the spell waits on it.
2. **It is enumerated** (`choiceMoves`). It joins the card-set pick case at `legal/choices.go:591` with the verb ": choose Ring-bearer", and the pool is walked in the order a new `Options.OrderRingBearer` hook gives.
3. **On departure it is dropped**, with `dropDefault` (`choiceDepartureDecisions`). The candidates are the departing player's own creatures, which CR 800.4a has taken with them. The drop runs the continuation with nothing chosen, so the tempt still completes (CR 701.54d) and no frame is stranded. A trigger it would cause is controlled by a player who has left, so CR 800.4d keeps it off the stack.

The prompt holds a Go continuation, as amass's and every card-set pick's do, so a table sitting on an open `ring_bearer` prompt is counted under `ChoiceResumeFrames` and is not a restore point for that moment. That is the existing posture for every resolution-time pick and is not changed here.

Choosing a Ring-bearer is not targeting. Hexproof, shroud, protection and ward do not apply, and nothing can respond to the choice.

### 5. The evasion: emblems join the block-rule walk

`forEachBlockRuleLocked` gains the sixth emblem walk, after the battlefield and before the scoped records. It reads `EmblemDef.BlockRules`, a new slot on the emblem def, mirroring `Spec.BlockRules`. The Ring's rule is `CantBeBlockedBy(YourRingBearer(), GreaterPowerThanAttacker(), "creatures with greater power")`:

- `YourRingBearer()` is a `BlockScope` that accepts an attacker that is `RingBearerOf` the source emblem's controller.
- `GreaterPowerThanAttacker()` compares `PowerForComparison`, as Locke's hand-written rule does. Locke moves onto the shared predicate in the same PR.

Because the rule is read through `BlockPairRefusalLocked`, the enumerator never offers the illegal block and the refusal sentence names the source. That source is an emblem, not a card on the battlefield, so the client's refusal text has to fall back to the emblem's label rather than look up a board card. A test pins it.

### 6. Deck construction (CR 113.6n)

`deck.Validate` honours "A deck can have any number of cards named ~" and "A deck can have up to N cards named ~" by reading the card's own oracle text. Thirteen Commander-legal cards print one: Nazgûl (nine), Seven Dwarves (seven), and eleven "any number" cards (Cid, Timeless Artificer; Dragon's Approach; Hare Apparent; Persistent Petitioners; Rat Colony; Relentless Rats; Shadowborn Apostle; Slime Against Humanity; Sphinx's Approach; Tempest Hawk; Templar Knight). Ten Nazgûl is still a `singleton_violation`, and its message names the limit. This is independent of the Ring and unblocks uploading #2062 at all.

### 7. The bot

- **The enumerator** offers each candidate as one move (§4). The prompt is always answerable: its floor is 1, it is only asked when there are candidates, and the submit re-check refuses a creature that has gone, leaving the prompt open with the rest.
- **The heuristic** orders the candidates by the power the creature attacks with, then by whether it can attack this turn, then prefers the current Ring-bearer, which costs nothing to keep. It ranks last a creature whose new legendary status would put it into the legend rule against a same-named legendary permanent it controls.
- **The emblem's effects** need nothing new. Legendary and the evasion reach the evaluator through effective characteristics and block refusals. The loot's discard is the ordinary discard prompt.
- **The model tiers** see `ring_bearer` on the filtered view's card and `emblems[].level`. `aiseat/model`'s board text gains one line per emblem ("The Ring (tempted 3 times): …") and a "Ring-bearer" tag on the card. `aiseat/` still imports nothing from `internal/game`.

### 8. Persistence

- **Two additive card fields** under schema v7: `ringBearer` and `ringTemptations` in `cardSnapshot`, recorded in `testdata/snapshot_shape/v7.txt` with `-update-shape`. Clone copies them with the struct.
- **One new on-disk identity**, `emblem:the-ring`, the Ring emblem's `OracleID`.
- **One new registered body**, the level 3 delayed sacrifice, so a table holding that delayed trigger stays a restore point. The emblem's triggered rows are catalog rows. `registerRulesEmblem` files them through the same `game.IdentifyCatalogRows` call `carddef.go` makes for every def, so their stack items are stamped `catalog/triggered` and restore like a card's. PR 3 pins that with a round-trip test.
- **Closure ratchet:** no new route. A bool and an int on `Card`, and `EmblemDef.BlockRules`, which is catalog data, not game state.
- **The fixture corpus** gets one new board in `testdata/snapshots/v7/`: two players with the Ring at levels 1 and 4, a Ring-bearer each, and the level 3 delayed sacrifice queued. The writer adds a file and never touches an existing one.
- **Rollback.** An older v7 binary ignores the two fields and finds no def under `emblem:the-ring`. The emblem comes back with no abilities, which the restore ability check already flags and logs. This is the posture `Card.Monstrous` and `Card.ControlledSinceUpkeep` took within v7. It is not a bump, because every change here is additive (`SnapshotSchemaVersion`'s own rule).

### 9. The wire and the log

- `CardView.ring_bearer` (bool, public). A face-down Ring-bearer shows it too: the designation was chosen in public and says nothing about the hidden card.
- `EmblemView.level`, and the Ring's `text` as in §2.
- `EventRingTempted` gets a log line: "The Ring tempts Alice (2) — Alice chooses Nazgûl as her Ring-bearer", or "— Alice controls no creature".
- The prompt is a `pending_choices[]` entry of kind `ring_bearer`. Its reason, and so its dialog name in the action dock (ADR 0111 §10), is "choose your Ring-bearer". That name joins the labels contract once it ships.
- `docs/protocol.md` documents all four.

---

## Delivery

Each PR lands test first. Each verifies every card it lands against its full oracle text, moves any card that needs more onto the row's `Waiting` list with the reason, and flips the registry row when the seam ships (with a closed-seam fragment under `docs/engine-seams/closed/`).

| PR | What | Cards (Completeness) | Needs |
|---|---|---|---|
| 1 | §6: CR 113.6n in `deck.Validate` | none in the catalog; #2062 becomes uploadable | — |
| 2 | §1–§5, §7–§9 for level 1: the tempt, the emblem and its key, `RingTemptations`, `RingBearer`, `ring_bearer` and its three rows, `EventRingTempted`, `DesignationRingTempted`, the emblem block-rule walk, the wire fields, the log line, a minimal client answer, the enumerator and heuristic, the shape and the fixture | none; it is a seam PR, like ADR 0108's PR 0 | — |
| 3 | The emblem's levels 2 to 4, the registered delayed sacrifice, and the five deck cards | Call of the Ring, Nazgûl, Ringsight, Sauron, the Dark Lord: Full. Sauron, Lord of the Rings: per question 4 | 1 (Nazgûl's deck rule), 2 |
| 4 | Client display, per question 1 | none | 2 |
| 5 | Every other Ring card that needs nothing else, per question 3 | about 41 of the remaining 51 | 3 |

PR 2 lands no cards on purpose. A card that tempts while the emblem has only its first line would be weaker than printed for everyone who reaches level 2, so no card ships until PR 3 has all four.

### PR 1 — deck construction

- **Change:** `deck.Validate` reads an exemption from the card's oracle text, with "any number" as no limit and "up to N" (written as a word, "nine" or "seven") as N. The upload path is unchanged: a violation is still fatal.
- **Snapshot:** none. **Wire:** none. **Bot:** none.
- **Tests:** nine Nazgûl pass, ten fail with the limit in the message, thirty Relentless Rats pass, two copies of a card with no such line still fail, and a pasted #2062 list validates.

### PR 2 — the tempt and the Ring-bearer

- **Data model:** `Card.RingTemptations`, `Card.RingBearer`, `game.RingEmblemKey`, `DesignationRingTempted` and `game.RingTempted(n)`, `EmblemDef.BlockRules`, `PendingChoiceRingBearer`, `EventRingTempted`, and `Options.OrderRingBearer`.
- **Engine:** `RingTemptsForEffect`, `RingBearerOf`, `RingTemptCount`, the three clear sites, and the emblem walk in `forEachBlockRuleLocked`. The layer version is bumped on the count and on the designation.
- **Cards side:** `the_ring.go` with the level 1 static and block rule. Also `TheRingTemptsYou`, the three trigger constructors, `YourRingBearer` and `GreaterPowerThanAttacker`, and Locke moved onto the shared predicate.
- **Snapshot:** §8's two fields, the shape record, and the v7 fixture board (at level 1 for now; PR 3 adds the level 4 board).
- **Wire and client:** `CardView.ring_bearer`, `EmblemView.level`, the log line, and `ring_bearer` added to `ChoicePromptModal`'s card-set kinds so a human can answer. A plain text badge is enough until PR 4.
- **Bot:** §7.
- **Registry:** a new `the-ring-tempts-you` row (`KindMechanic`, rules 701.54, issue 2076, this ADR), `StatusMissing`, with the 56 cards on `Waiting`.
- **Tests:**
  - The emblem is created once, at the first temptation, before the choice.
  - The count rises on every temptation, including one with no creature.
  - The event fires with and without a choice.
  - Re-choosing the same creature fires it again with that creature.
  - A second Ring-bearer clears the first.
  - A control change (Act of Treason) ends the designation. The thief's own later tempt can choose the creature.
  - Bounce and flicker end it.
  - A phased-out Ring-bearer is nobody's and is back on phasing in.
  - A Clone of a Ring-bearer is not one.
  - Legendary applies through layer 4, and two same-named legendaries under one controller go to the legend rule.
  - The evasion refuses a greater-power blocker, allows equal power, is withheld by `legal.EnumerateFor`, and names the emblem in the refusal.
  - Four players each get their own emblem and Ring-bearer.
  - A player who leaves takes their emblem, and their open `ring_bearer` prompt is dropped with the tempt still completed.
  - `TestEveryChoiceKindIsClassifiedAndEnumerated` and `TestEveryChoiceKindHasAReassignmentDecision` pass.
  - Snapshot shape, the corpus and the closure ratchet pass.

### PR 3 — levels 2 to 4, and the deck's five cards

- **Engine:** the emblem's three triggered rows and the registered end-of-combat sacrifice body.
- **Tests:**
  - Level 2 loots once per attack, in each player's own combat, and not before the second temptation.
  - Level 3 triggers once per blocker (two blockers, two sacrifices). The blocker's controller at end of combat sacrifices it. A blocker that left and came back is not sacrificed. It does not trigger when the Ring-bearer blocks.
  - Level 4 makes every opponent lose 3, not only the one damaged, and triggers on a Ring-bearer that dies to the blocker's damage in the same step.
  - A table with each trigger on the stack round-trips through `RestoreStrict`.
- **Cards:**
  - **Call of the Ring:** an upkeep trigger that tempts, and `WheneverYouChooseARingBearer` with `MayChoice{LifeCost: 2}` that draws. It does not trigger with no creature, and does trigger on re-choosing. Full.
  - **Nazgûl:** deathtouch, an ETB tempt, and `WheneverTheRingTemptsYou` with a +1/+1 counter on each Wraith you control. The deck rule is PR 1's. Full.
  - **Ringsight:** tempt, then a library search for a card that shares a color with a legendary creature you control, read after the tempt. Per the rulings, the new Ring-bearer counts and a colorless legend adds nothing. Full.
  - **Sauron, the Dark Lord:** `WardSacrifice` over legendary artifacts and legendary creatures (an opponent's Ring-bearer qualifies); amass Orcs 1 when an opponent casts a spell; a tempt when an Army you control deals combat damage to a player; and `MayChoice` to discard the hand and draw four. Full.
  - **Sauron, Lord of the Rings:** the cast trigger (amass Orcs 5, mill five, then return a creature card of your choice from your graveyard, which need not be one you milled), trample, and "Whenever a commander an opponent controls dies, the Ring tempts you". Caveats, or not in this PR: see question 4.
- **Registry:** the row goes to `StatusPartial` with the landed cards as `Examples`.

### PR 4 — what the table shows

Per question 1: the Ring chip with its level and lines, the Ring-bearer marker on the card, its hover title, and the log line's rendering. The labels it adds (the marker's `aria-label` "Ring-bearer", and the "choose your Ring-bearer" dialog) are new, not renamed. The PR still runs the nightly Playwright job on its branch, as AGENTS.md's labels contract asks of a change to the dock.

### PR 5 — the rest of the Ring cards

Every one of the ✓ and ? cards in [Sizing](#sizing) that turns out to need nothing else, each verified against its full text. Any that needs more goes on the row's `Waiting` list with the reason. Lord of the Nazgûl lands only if protection from a designation turns out to be one predicate on the existing protection vocabulary. Galadriel, Elven-Queen waits on voting. The row goes to `StatusImplemented` when nothing but other seams' cards is left waiting.

### Order

PR 1 has no dependencies and goes first, so #2062 can be uploaded and played in the sandbox while the rest is built. PR 2 can start at the same time. PR 3 follows PR 2. PR 4 follows PR 2 and can run beside PR 3. PR 5 follows PR 3. S58's card PR 7 (Sauron, Orcs and amass) does not wait for this ADR: any Ring card it meets goes on the new row's `Waiting` list, and if it merges before PR 2, it adds the row as `StatusMissing` itself.

---

## Consequences

- The engine can tempt a player, and every one of 56 printed cards has the action, the designation and the four-level emblem it needs.
- An emblem no card creates is possible. Its key is fixed, and it reaches the five existing emblem walks and the new block-rule one with no special case.
- Emblems can carry a block rule. The next emblem that restricts blocking is one slot entry.
- An emblem can switch its own abilities on with a count it carries. The count is one int on `Card`, used only by the Ring.
- A designation can end on a control change as well as on a zone change, at the one control step.
- Commander decks may run the 13 cards whose text changes deck construction.
- Rollback to an older v7 binary loses the Ring's abilities and the Ring-bearers on a restored table, and the restore check reports it. This is the same trade the other in-version designations made.

## Alternatives considered

- **The count on `Player`** (`Player.RingTemptations`), the literal subject of CR 701.54c. Rejected: ADR 0071's gate reads the object and nothing else, and a gate that needs `*Game` would have to be special-cased in each of the six emblem walks. Since the emblem exists exactly when the count is at least 1, putting the count on it loses nothing.
- **The count in `Card.ClassLevel`.** It would need no new field. Rejected: a Class reader that one day walks the command zone would read the Ring as a level 3 Class, and `ClassLevelOf` floors at 1.
- **The Ring-bearer as `Player.RingBearer ObjectRef`.** It would make "one per player" structural. Rejected: every reader and the wire want the mark on the card, as `Monstrous` and `Harnessed` have it, and a reference needs its epoch checked everywhere it is read. The single write site in §1 keeps the invariant.
- **The four abilities as a monarch-style `Listener`** with sourceless items. Rejected: CR 701.54c gives the abilities to the emblem, so they have a source, and a listener would also need its own wire, persistence and CR 800.4a handling.
- **Four emblem keys, one per level**, swapping the emblem object as the count rises. Rejected: it changes the object's identity, and its CR 613.7c timestamp, on every temptation.
- **Reuse `own_permanents` or `choose_cards`.** Rejected in §4: the bot would order the candidates backwards.
- **A CR 614 window on the tempt**, as amass has. Rejected: nothing printed replaces a tempt.

## Out of scope

- CR 903.9a as a state-based action (question 4 proposes its own issue and ADR).
- Voting (Galadriel, Elven-Queen).
- Making an open `ring_bearer` prompt a restore point. That is ADR 0041 phase 3's general work on resume frames, not this seam's.

---

## Open questions for the owner

Each has options and a recommended answer. Nothing else in this ADR is a choice: the rules and the earlier owner decisions settle it.

### 1. How the table shows the Ring and the Ring-bearer

Every seat may have its own Ring at its own level, and its own Ring-bearer.

- **(a) Recommended:** the Ring is the existing emblem chip on the player panel, with its level as a pip ("The Ring · 3"). Its hover lists all four lines, the gained ones in full and the rest dimmed with "after the 4th temptation", so a player can see what is coming. The Ring-bearer gets its own small ring marker on the card, in a corner that is not the designation badge slot, with the title "Alice's Ring-bearer". A Ring-bearer can also be monstrous or harnessed, and the one badge slot shows only one of them (`Card.svelte:385`).
- (b) The same chip, but its hover shows only the gained lines (the emblem's actual abilities), and the Ring-bearer takes the designation badge slot as "RING-BEARER", after the existing tenants in the priority chain.
- (c) A marker in the panel's marker column beside the monarch crown, with the level, instead of an emblem chip.

### 2. A forced choice: ask, or choose for the player?

With exactly one creature, the rules leave no choice (the ruling: "you must choose a creature if you control one").

- **(a) Recommended:** choose it without a prompt, and say so in the log line ("Alice's only creature, Nazgûl, becomes her Ring-bearer"). This is amass's precedent (`amass.go`: "the prompt queue is not the place to make the player click through a forced answer"). Nine Nazgûl and an upkeep trigger make it a frequent case, and the bot loses nothing.
- (b) Always ask, even with one candidate, so a player always sees the moment the Ring-bearer is set. It costs a click per tempt and one more table-blocking prompt.

### 3. Which cards land, and when

ADR 0106 owner decision 6 says a PR lands every card its seam unblocks. Applied here, that is about 46 cards.

- **(a) Recommended:** PR 3 lands the deck's five, and PR 5 lands every other Ring card that needs nothing else, in its own PR so the deck is not held up by the breadth.
- (b) PR 3 lands all of them at once. One PR, but the deck waits for about 46 card files and their tests.
- (c) The five only. The other 51 wait on the row's `Waiting` list until a deck asks for them.

### 4. Sauron, Lord of the Rings and a commander that dies

"Whenever a commander an opponent controls dies" triggers "even if the owner of the commander that died chooses to return it to the command zone after it dies" (ruling), because CR 903.9a puts the commander in the graveyard first and moves it as a state-based action. The engine applies CR 903.9 as a replacement to graveyard and exile moves too (`builtin_replacements.go:33`), so a commander whose owner takes the command zone never dies, and the trigger never sees it. The same gap affects every "dies" trigger in the catalog that can see a commander.

- **(a) Recommended:** open a separate issue and ADR to bring graveyard and exile under CR 903.9a (move first, then the state-based choice), keeping the replacement for hand and library (CR 903.9b). Sauron, Lord of the Rings lands in PR 3 with one caveat ("Doesn't trigger when an opponent's commander dies and its owner moves it to the command zone"), which comes off when that ADR ships. It is the most rules-faithful end state, and it is not held up by this seam.
- (b) The same separate ADR, but Sauron, Lord of the Rings waits for it on the row's `Waiting` list rather than shipping with a caveat.
- (c) A narrow fix in this seam: Sauron's trigger also watches a commander redirected from the battlefield to the command zone. It is smaller, but it gives "dies" a second meaning for one card while every other dies trigger keeps the gap.
