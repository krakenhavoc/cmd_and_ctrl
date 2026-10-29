---
title: "Putting a card onto the battlefield as a different type"
date: 2026-10-29
issues: [4506]
---
**Putting a card onto the battlefield as a different type** (#4506, [ADR 0210](decisions/0210-putting-a-card-onto-the-battlefield-as-a-different-type.md), CR 614.1c, CR 614.12, CR 205.1a) — the putting effect can set what a card enters as. `game.EntersAs` (types, subtypes, granted bundle keys) rides `ZoneEntryOptions.EntersAs`, and `effects.ReturnFromGraveyardTogether` exposes it. As the permanent lands, before the zone move and `EventETB`, the landing registers one `ScopedEffect`: `SetTypesMod` (layer 4) and `GrantAbilitiesMod` (layer 6), pinned to the object by its epoch.
- **The effect.** The card is never a creature on the battlefield, so "whenever a creature enters" does not trigger. It lasts as long as that object is on the battlefield: a flicker, a reanimation or a copy (CR 707.2) does not have it. The entry window judges the card as it will be (CR 614.12): the look-ahead dry run applies the overwrite, and `LookupCardForEffect` answers with that dry run for the span of a replacement gather (`Game.entryView`, never set between actions).
- **Snapshot:** additive in v9. `entersAs` is written on a paused entry's event and on its batch's options, and the shape is recorded. **Closure ratchet:** unchanged.
- **Cards (2, both Full):** Supper for Spiders (opponents' creature cards put into a graveyard from the battlefield this turn come back as Food artifacts under your control) and Tom, Bert, and William (they return once, as a legendary noncreature artifact).
- **Tests:** `cards/effects/enter_as_different_type_test.go`: what Supper takes; the Foods with Soul Warden; a legend; a 0/0; the Food ability; Authority of the Consuls; a token copy; a flicker; a restore; a paused Clone entry restored with a second batch member; Tom, Bert, and William's single return, Wrath of God, the second death, the reanimation and the sacrifice ability.
- **Back-outs:**
  - With no registration at the landing, eight tests fail.
  - Registering after `EventETB` fails the Soul Warden, Authority and Tom, Bert, and William tests.
  - Without the entry-window view, the Authority test fails.
  - An unpinned record fails the flicker sweep and the reanimation.
  - Dropping either snapshot mirror fails the paused Clone restore.
