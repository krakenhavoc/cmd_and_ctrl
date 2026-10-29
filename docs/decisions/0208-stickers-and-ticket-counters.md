# ADR 0208 — Stickers and ticket counters

**Status:** Accepted · 2026-10-24 · S76 — Cards: incubate, face-down and outside the game.
**Issue:** [#2138](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2138) (found landing the undying and persist pool, [#2075](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2075)). Related: [#2121](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2121) (Attractions, the other Unfinity seam).
**Numbering:** checked on 2026-10-24. `develop` ends at 0207, and `git log --remotes=origin --name-only -- docs/decisions` over every remote head shows no file above 0207. The 19 heads checked are listed in the PR. No open issue or pull request mentions 0208.
**Builds on:** [ADR 0207](0207-attractions.md) (its Consequences named the attachment points this uses: tickets on `Player.Counters`, a sticker as data on its card), [ADR 0191](0191-a-sideboard-is-the-outside-the-game-pool.md) (a deck section that is said and not seated when it breaks a rule), [ADR 0129](0129-energy-getting-and-paying-it.md) (a resource counter on the player and the one path that pays it), [ADR 0093](0093-abilities-granted-to-other-permanents.md) (a grant bundle on another object), [ADR 0101](0101-keyword-counters.md) (a layer source that is data on the permanent, not an ability of it), [ADR 0041](0041-game-persistence.md) (the additive snapshot rule and the closure ratchet), [ADR 0149](0149-privacy-notice-and-account-deletion.md).

---

## Context

### What the issue named

Night Brushwagg Ringmaster and Phyrexian Midway Bamboozle are not cards. They are two of Unfinity's 48 **sticker sheets**, which Scryfall lists as records typed `Stickers` (set `sunf`), and 46 of them are `legal` in Commander there. Each printed sheet carries three name stickers (its name, split into words: "Night", "Brushwagg", "Ringmaster"), three art stickers, two ability stickers and two power and toughness stickers, each of the last four with a ticket cost. Scryfall's text lists only the costed ones; the name and art stickers, and the fact that they cost nothing, were read off the printed sheets (the Scryfall images of four sheets, 2026-10-24). One legal sheet has a two-word name sticker: Ancestral Hot Dog Minotaur's are "Ancestral", "Hot Dog" and "Minotaur".

A sheet does nothing alone. A sticker goes on a permanent only when an effect says so, and those effects are on ordinary cards: Ticketomaton, Command Performance, Costume Shop and some thirty more Unfinity cards that are legal in Commander.

### The rules

Checked against `MagicCompRules 20260925.txt`.

- **107.17 / 107.17a.** `{TK}` is one ticket counter. A number in a ticket symbol is a ticket cost: "To pay that cost, a player removes that many ticket counters from themselves."
- **122.1.** A counter may be on a player.
- **123.1.** A sticker is a marker on an object, not an object, not a counter, not a token; "Changes to an object from stickers are not part of its copiable values." Four kinds: name, ability, power and toughness, art.
- **123.2 / 123.2a / 123.2c / 103.2d.** Sticker sheets are not cards and have no characteristics. In constructed play a player who plays with stickers starts with at least ten sheets, each unique, reveals them all and chooses three at random; they have access to the stickers on those three only, and those stay revealed. Commander (CR 903) has no rule of its own about stickers.
- **123.3 / 123.3a-d.** Putting a sticker on: the player chooses a sticker that is not on any object they own; each sticker is distinct; they can't put one on an object they don't own; a sticker with a ticket cost needs the object's owner to pay that much `{TK}`, and can't go on otherwise; moving one already on an object costs nothing again.
- **123.4.** "Stickered" is having any sticker on it now.
- **123.5 / 400.7m.** Stickers come off on a move to a hidden zone and stay on through a move to a public zone, applying to the new object (an exception to 400.7).
- **123.6 / 123.6a-c / 612.9.** A name sticker adds its word to the name, a text-changing effect (layer 3, 613.1c). Its controller chooses where, at the start or after any number of words; a blank line is not a word; with fewer words later it goes at the end.
- **123.7.** An ability sticker gives its ability (layer 6, 613.1f).
- **123.8.** A power and toughness sticker sets the creature's power and toughness (layer 7b, 613.4b), and the later of two wins (613.7). It also applies to a creature or Vehicle card in another zone.
- **123.9.** An art sticker only marks the object.
- **613.7k.** A sticker gets a timestamp each time it is put on, and a new one immediately after the object's own new timestamp.
- **702.33h.** Sticker kicker (not built here).

## Options considered

**Where the sheets live.**

- **A. As cards in a container outside the game, like the sideboard (ADR 0191) or the Attraction deck (ADR 0207).** Rejected: a sheet is not a card (CR 123.2), nothing ever moves it, and a new per-player container of cards is a new route the closure census would have to classify (`census:IntrinsicAbilityCards`, a ceiling raise the owner approves).
- **B. As plain data on the player (chosen).** `Player.StickerSheetPool` and `Player.StickerSheets` are slices of `StickerSheet`, a struct of strings and numbers. No route to a card or a function, so the closure ceilings do not move.

**Where a placed sticker lives.**

- **A. A `ScopedEffect` per sticker.** Rejected: a scoped effect ends with its object (CR 400.7), and a sticker must follow the card to its next public zone (CR 123.5); it would also need its own lifecycle for hidden-zone moves.
- **B. Data on the card, read by the layer pass (chosen).** `Card.Stickers`, the way keyword counters are data the pass reads (ADR 0101). `MoveCard` keeps or drops it by destination.

**What an ability sticker's ability is.**

- **A. Parse it from Scryfall's text.** Rejected: a triggered or activated ability needs code, and parsing keywords alone would make the uncatalogued sheets half-work in a way no card file declared.
- **B. The catalog's (chosen).** `Spec.StickerSheet` says, per ability sticker in printed order, which keywords or which grant bundle (ADR 0093) it gives. An uncatalogued sheet's name, power and toughness, and art stickers still work (the rules define them); its ability stickers go on and do nothing, weaker than printed, and the deck checker says the sheet is not automated.

## Decision

### 1. Tickets

`game.CounterTicket` ("ticket") is a player counter in `KnownPlayerCounters`, beside energy and rad. "You get `{TK}`" is `effects.GetTickets{N}`, through `AddPlayerCounterByForEffect`, so a CR 614 window on player counters and `EventPlayerCounterPlaced` see it. A ticket cost is paid only by `payTicketsLocked` (CR 107.17a): it refuses, changing nothing, when the payer has fewer (CR 118.3), and announces a negative counter delta. No printed card that is legal in Commander has `{TK}` in an activated ability's cost: every ticket cost is a sticker's, paid as that sticker goes on. So `{TK}` is not an `AbilityCost` component; if a card ever needs one, it joins `AbilityCost` beside `Energy` (ADR 0129 §2).

### 2. The sheets: a deck section, a revealed pool, three chosen

- `deck.PrintsStickerSheet` is a record whose type line is `Stickers`. `Resolve` puts every such record in `List.StickerSheets`, wherever the list named it (a `Stickers:` or `Sticker sheets` heading, Moxfield's `stickers` board, a sideboard section, or none): never one of the 100, so count, singleton and colour identity never see it. `List.Text()` writes the section back, so a saved deck reseats it.
- **CR 123.2a**, the constructed rule, because Commander has none of its own: fewer than ten sheets, or one twice, is warned (`sticker_sheets_illegal`, non-fatal, beside the sideboard and Attraction warnings in `lobby`) and `SeatedStickerSheets()` seats none, ADR 0191 §5's posture. Each sheet is held to the format's legality like a card.
- `deck.StickerSheetOf` reads a record into a `game.StickerSheet`: the name stickers (the name's words, or `stickerNameWords` for a sheet whose name sticker has two), the ability stickers and power and toughness stickers with their ticket costs in printed order, and three art stickers. `ToGameCards` hands each sheet to the game on a pseudo-card (`Card.StickerSheet`), and `seatDeckCard` turns it into `Player.StickerSheetPool`; the pseudo-card goes into no zone.
- **CR 103.2d.** `dealAndBeginLocked` chooses three of the pool at random into `Player.StickerSheets` on a stream of its own (`sticker_sheets`), so a deck with sheets shuffles its library exactly as it would without them. A restart (CR 727.2) carries the pool and chooses again.

### 3. A sticker on an object is data

`Card.Stickers` is `[]PlacedSticker`: the sheet's oracle ID and the sticker's index (together, which sticker this is, CR 123.3a), what it says, its CR 613.7k timestamp, and for a name sticker the number of words before it (CR 123.6c). `MoveCard` drops them on a move to a library, a hand or out of the game and keeps them on every other move (CR 123.5); a minted token, a copy of a spell and the develop spawner never inherit them (CR 123.1, 707.2). A sticker is available while it is on one of the player's three sheets and on no object they own in any public zone (`AvailableStickersLocked`, CR 123.3).

### 4. Putting one on: three keyed questions

`Game.PutStickerForEffect(StickerRequest)` is CR 123.3 for an instruction that names the objects ("a nonland permanent you own") and, optionally, a kind ("a name sticker"). It asks nothing when no object qualifies or no sticker of an allowed kind is available and affordable. Otherwise:

1. **Which object, or none** (`sticker/choose-object`, a keyed `choose_cards` of zero or one: every printed instruction says "you may"). The allowed kinds ride as the prompt's word.
2. **Which sticker** (`sticker/choose-sticker`, a keyed option pick): "Don't put a sticker", then every available sticker the owner can pay for (CR 123.3c), the object and the stickers' stable IDs (`StickerID`) riding as `Carry`.
3. **Where a name sticker goes** (`sticker/name-position`), asked of the object's controller (CR 123.6b), one option per resulting name.

Each is registered under an on-disk key in the effect-key ledger, so a table waiting on any of them is a restore point. `placeStickerLocked` re-checks the object, the owner and the sticker, pays the ticket cost (CR 123.3c), stamps the timestamp and announces `EventStickerPlaced` (its label is the kind and the text, read with `StickerEventKindText`). A `Do` or `ResolveSteps` walk waits for the questions (the step walk, #4012), so Finishing Move's damage is dealt after a power and toughness sticker has gone on its biter.

### 5. What a sticker does: one more layer source

`stickerContinuousEffectsLocked` is one more source list in the gather, beside the keyword counters, with each sticker at `max(PlacedAt, object timestamp + 1)` (CR 613.7k):

- a **name** sticker in layer 3 inserts its word (`InsertNameWord`), so the legend rule, every "named" check and the wire read the new name;
- an **ability** sticker in layer 6 adds the catalog's keywords and grant (`CatalogStickerAbility`, from `Spec.StickerSheet`). Undying and persist are read off the dying permanent's last-known abilities (ADR 0113 §4), so a sticker's persist returns the creature, and the sticker comes back with it;
- a **power and toughness** sticker in layer 7b on a creature (CR 123.8);
- an **art** sticker does nothing (CR 123.9).

A sticker has no source permanent, so CR 613.6's silencing never touches it: a "loses all abilities" with a later timestamp takes an ability sticker's ability away and an earlier one does not. Off the battlefield, `effectiveOf` folds the name words, the ability stickers' keywords and the latest power and toughness sticker on a creature or Vehicle card onto the baseline, as it does keyword counters.

### 6. The wire and the table

- `PlayerView.sticker_sheets`: the three sheets, public (CR 123.2c), each sticker with `in_use` while it is on an object. Tickets ride `PlayerView.counters.ticket`.
- `CardView.stickers`: the stickers on an object, redacted with the rest of a hidden card's identity. The card's name, abilities and power and toughness already show what they changed.
- The log says each placement: "Ian put the Menace sticker on Bear (2 tickets)" (`sticker_placed`).
- The client draws a ticket marker on the seat (a count; only the rules give or spend tickets, so no stepper), a sticker-sheet marker whose number is the stickers still free and whose hover lists them, a sticker badge on the card's left edge with the stickers on hover, and the stickers in the hover panel. The questions are the ordinary `choose_cards` and option-pick prompts.

### 7. Persistence and the closure ratchet

- **Snapshot (additive within v9):** `seats[].stickerSheetPool`, `seats[].stickerSheets` and the card's `stickers`, plain JSON. An older file has none, which is right for every game before this; an older binary drops them and the objects lose what the stickers gave them, weaker than printed. Recorded with `-update-shape`; no fixture edited, no version bump.
- **Clone** deep-copies the sheets and the stickers. `Card.StickerSheet` is classified `dropped` ("always empty on a card in a zone").
- **Closure ratchet:** nothing here is a route from `Game` to a function or a container of cards. The ceilings do not move.

### 8. The deck checker

`deckcoverage` routes a sheet the way `Resolve` does and reports it in `Report.sticker_sheets`, bucketed by the catalog (a catalogued sheet is automated, another is manual), outside the deck's own rows, counts and copies.

### 9. Privacy

A deck's sticker sheets are deck data, stored where its cards are (the deck's source text, ADR 0051), and say nothing about the person beyond the deck itself. Nothing new is stored about a person and nothing is kept longer, so ADR 0149's inventory does not change.

### 10. Cards

The vocabulary (`cards/effects/stickers.go`): `GetTickets`, `PutASticker{On, Kinds}`, `GetTicketsThenSticker`, `WhenThisEntersGetTicketsThenSticker`, `Stickered`, `WithSticker`, `WheneverYouPlaceASticker`, `PlacedStickerKind`, and `Spec.StickerSheet`. Twenty-nine entries ship, all `full`: the two sheets; Ticketomaton, Aerialephant, Minotaur de Force, Stiltstrider, Carnival Carnivore and Glitterflitter (enter, get tickets, then a sticker); Grabby Tabby, Scared Stiff, Sanguine Sipper, Croakid Amphibonaut and Big Winner (a keyword while you control a stickered permanent); Ticket Turbotubes, Blorbian Buddy and Prize Wall (tickets for mana); Bioluminary; Costume Shop (a visit); Wee Champion, Goblin Airbrusher and Tusk and Whiskers (whenever you put a sticker on); Robo-Piñata and Command Performance (modes); Finishing Move; Baaallerina, Sword-Swallowing Seraph and Proficient Pyrodancer (a name or art sticker, and an ability that reaches a creature with one); Lineprancers; and Park Bleater.

## Consequences

- **The other 46 legal sheets** work at once for their name, power and toughness, and art stickers. Each one's ability stickers need a catalog entry (`Spec.StickerSheet`), most of them a keyword or a short grant bundle; until then the deck checker lists the sheet as not automated.
- **Still open, as card work on this seam:** sticker kicker (CR 702.33h, Wicker Picker), "has all abilities of ability stickers on …" (Pin Collection, Clandestine Chameleon), the name-sticker word and vowel counts of the `_____` cards, a sticker put on a card in a graveyard (Scampire) and counting stickers there (Roxi, Ambassador Blorpityblorpboop), moving a sticker (CR 123.3d), and "Done for the Day". A blank in a name is kept where it is; nothing yet fills it.
- **Not built:** the limited rule (CR 123.2b) and Moxfield's board name for sheets is assumed to be `stickers`; a list under no heading seats the same way.


---

## Amendment 2026-10-26 — the rest of stickers (#4307)

**Issue:** [#4307](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4307), the follow-up the Consequences above named. Checked against `MagicCompRules 20260925.txt`: CR 123.1-123.9, 702.33h, 123.3d, 123.6d/e. No new ADR number is taken.

### What shipped, in the order the issue listed it

**1. The other sheets' ability stickers.** `Spec.StickerSheet` entries for 45 more of the 48 Commander-legal sheets (47 in all), each ability sticker naming keywords, a grant bundle (ADR 0093), or `StickerAbilitySpec.GraveyardCastLife`. Pieces the abilities needed:

- *Keywords as rows.* `Afflict(n, name)` and `BattleCry(name)` are triggered rows beside `Bushido`, `Mentor`, `Ward`, `Provoke` and `Monstrosity`, which existed. Outlast is an activated bundle.
- *Layer-6 statics inside a sticker's bundle.* A sticker writes its grant as `stickerEffect.Apply` does, not as a static's declared `GrantAbilities`, so the layer pass never recorded it (`noteGrants`). `applyOneEffectLocked` now notes any grant an effect appended while it applied (`noteAppendedGrants`, by the length of the host's grant list before and after), which also covers a borrow. Mystic Doom Sandwich's "must be blocked if able" and Trained Blessed Mind's threshold trample are such statics. A layer-7 static of a bundle was already gathered from the final grant list.
- *Hellbent.* A static in a granted bundle may now declare `DependsOnHandSize` (`GrantedStaticProblem` allowed it; the others are still refused), and the hint walk (`staticOnBattlefieldLocked`) reads the statics of the grants a permanent holds as well as its own, so a draw or a discard re-reads Vampire Champion Fury.
- *Four protection qualities* (`game/protection.go`, closed grammar): "odd mana values", "even mana values", "creatures with two or more creature types" (a changeling is every type) and "noncreature permanents". The last needs a fact about the source the snapshot did not carry, `Characteristic.SourceIsPermanent` (a layered characteristic means it is on the battlefield; a spell has none; a last-known source was a permanent). Wire kinds `mana_value_odd`, `mana_value_even`, `multityped_creatures`, `noncreature_permanents`.
- *A pairwise block rule.* `ModCantBeBlockedByObject` ("target creature can't block this creature this turn", Wrinkly Monkey Shenanigans): a block-rule scoped mod naming the one barred blocker as the object it is (ID and epoch, CR 400.7). `effects.CantBlockThisCreatureThisTurn`.
- *A cast permission from a sticker.* Eldrazi Guacamole Tightrope's "You may cast this card from your graveyard by paying 2 life in addition to paying its other costs" is derived on every query from the sticker on a card in its owner's graveyard (`stickerGraveyardPermissionLocked`, CR 123.7 and 123.5), never stored. It is claimed as the offer `sticker-graveyard`, which charges the printed mana cost and the life (`CastPermission.AlternativeCostFor`). Weaker than printed in one corner only: it is an offer, so it cannot be combined with another alternative cost the card prints; a card with no other way to be cast from a graveyard (every one the sticker is for) is unaffected.
- *Casting a card from the graveyard this turn* for Elemental Time Flamingo is a per-card permission (`effects.CastTargetFromYourGraveyardThisTurn`), unpriced beyond the card's own cost.
- Token row `1/1 white Clown Robot artifact`; a `Hellbent`/`Metalcraft`/`Threshold` static is written as the bundle's own `Static`.

**2. Sticker kicker (CR 702.33h).** `effects.StickerKicker(mana)` is an optional cost keyed `game.StickerKickerKey`. It is its own key, not `KickerKey`: an "if it was kicked" ability is linked to the kicker printed on its own object (CR 702.33e), and a granted copy must not be dropped as a duplicate of a printed kicker in the offer. Wicker Picker grants it with `Spec.GrantedOptionalCosts`. The second half of 702.33h, "you get a ticket counter and you may put a sticker on this spell", is an additional cost: `CastSpell` runs it (`payStickerKickerLocked`) after every cost is paid and before `EventCast`, so the sticker is on the spell card on the stack when the cast triggers look, and before anyone holds priority. The questions are the ordinary three keyed ones, with the spell as the object; the sticker stays on the permanent (CR 123.5).

**3. "Has all abilities of ability stickers on …".** `effects.HasAllAbilitiesOfAbilityStickers` is a layer-6 static that writes, for each ability sticker its lenders hold, what that sticker grants (`game.GiveAbilityOfSticker`, CR 123.7a: the ability the sticker grants to the object it is on, whether or not that object has it now). The recipient is "this creature" in each bundle. Pin Collection's lender is its own stickers; Clandestine Chameleon's are the other permanents its owner owns and the cards in its owner's graveyard. Pin Collection's "an ability sticker with ticket cost X or less on it without paying that sticker's ticket cost" is `PutASticker{MaxCostIsCastX, Free}`, which rides the questions as a limit: the first prompt's word becomes `kinds;max=N;free`, and the second and third questions' Carry leads with a marker ID, so a prompt asked before this existed reads as it always did.

**4. Name-sticker counts (CR 123.6d/e).** A placement files a step record (`StickerPlacedRecordKey`, #4163) under the asking card, so a later step of the same resolution reads "that sticker" from the game and survives a restore between the questions (`placedNameSticker`). `StickerUniqueVowels` counts the different vowels (A, E, I, O, U, Y, case-insensitive), `StickerLetterCount` a letter, `StickerLetters` the letters of a word (a hyphen or an apostrophe is not one), `StickerBeginsWith` the first letter. `EventStickerPlaced` carries the words of the object's name stickers (`Event.Words`), so a trigger whose target count depends on them reads it from its trigger context (Make a _____ Splash's "up to X targets", X the u's). The eleven `_____` cards are Full: _____ Goblin, _____ Bird Gets the Worm, _____-o-saurus, Wizards of the _____, Wolf in _____ Clothing (a reflexive "When you do", `up to X` targets as `Params.Amount`), _____ Balls of Fire, Make a _____ Splash, _____ _____ _____ Trespasser, _____ _____ Rocketship (a keyed "choose a letter" of A to Z), Fight the _____ Fight and Last Voyage of the _____. Last Voyage "becomes an Aura with enchant creature" through the Aura form of ADR 0154's 2026-10-23 amendment (`game.BecomeAuraForEffect` sets `Card.ReturnedAura`; `Spec.ReturnedAura` is "enchant creature"), returns a creature card chosen by a keyed `choose_cards`, attaches to what entered, and sacrifices the enchanted creature when it leaves from the last-known `PermanentInfo.AttachedObject`. A blank in a name is kept where it is, not a word (CR 123.6a).

**5. Stickers on cards in a graveyard.** `PutASticker{On: StickerOnCreatureCardInYourGraveyard}` (Scampire); a name sticker applies there (CR 123.6), an ability sticker's keywords through `applyStickersOffBattlefield` and its cast permission as above. Roxi counts permanents you control and cards in your graveyard with an art sticker as a characteristic-defining ability; Ambassador Blorpityblorpboop totals the printed power and toughness of the stickers on permanents you control (CR 123.8a); Unlawful Entry and A Good Day to Pie watch from the graveyard for an art or name sticker put on a creature (a creature is a permanent: a creature card in a graveyard is not "a creature").

**6. Moving a sticker (CR 123.3d) and Done for the Day.** Done for the Day is Full. **No Commander-legal card moves a sticker** (the dump was searched for "move"/"attached"/"transfer" beside "sticker": the only hit is Solaflora, Intergalactic Icon, which is not Commander-legal and does not move one), so CR 123.3d's rule — a moved sticker's ticket cost is not paid again — has no card to drive it and nothing is built for it.

Also found and added: **Chicken Troupe**, a legal sticker card the issue did not list (ward {2}; a ticket, then a sticker).

### Not built

- **Squid Fire Knight.** Its first ability sticker says "The next time target player would roll one or more dice this turn, instead they roll that many dice plus one, then you choose one of those rolls to ignore." Every dice-rolling card (`random_effects.go`: the Ancient dragons, Vexing Puzzlebox, Berserker's Frenzy, Hoarding Ogre, Maddening Hex, Reckless Endeavor, Component Pouch, Diviner's Portent, the Attractions visit) reads its rolls on the next line, and the choice is a second player's, made after seeing the dice. A replacement that waits for that answer means making each of those cards a continuation. The sheet is left out whole (it is `stkSheetsNotCatalogued` in the tests): its stickers go on and do nothing, and the deck checker lists it as not automated.

### Snapshots and the ratchet

Additive within v9, recorded with `-update-shape`: `Event.Words`. `Characteristic.SourceIsPermanent` rides the last-known snapshots (`lastKnownBattlefield`, `lastKnownPermanents`), recorded with the same flag, additive and read as false when absent (a source snapshot, like `SourceManaValue`). `CastPermission` is unchanged: the sticker permission is derived. The closure ceilings do not move: every prompt is keyed (`sticker/…` unchanged, `choose-cards/last-voyage-of-the-blank-return`, `option-pick/blank-blank-rocketship-letter`, `option-pick/done-for-the-day-reward`) and the two reflexive bodies are registered (`wolf-in-blank-clothing/shrink`, `fight-the-blank-fight/fight`); all are in the effect-key ledger.

## Amendment 2026-10-26 (second) — Squid Fire Knight (#4307)

The sheet the first amendment left out ("Not built" above) ships, Full. Its first ability sticker gives "{T}: The next
time target player would roll one or more dice this turn, instead they roll that many dice plus one, then you choose
one of those rolls to ignore", through ADR 0054's amendment of the same date: a dice roll is now a replaceable event
(`RepEventRollDice`) that can pause for the choice, every dice-rolling card's follow-up is a registered tail, and the
ability writes a one-use `ModRollExtraDieIgnoreOne` record (`game.RollExtraDieNextTimeForEffect`) whose controller is
the ability's. Its second gives protection from odd mana values (the quality the first amendment added). All 48
Commander-legal sheets are catalogued; `stkSheetsNotCatalogued` is empty and the deck checker reports Squid Fire Knight
as automated. Snapshots: `pendingChoices[].diceRoll`, additive within v9, recorded with `-update-shape`. The closure
ceilings do not move: the two prompts are keyed (`option dice/ignore-roll`, `option dice/order-replacements`) and every
follow-up is a ledgered tail.
