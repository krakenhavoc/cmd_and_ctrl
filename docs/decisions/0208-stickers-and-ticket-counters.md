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
