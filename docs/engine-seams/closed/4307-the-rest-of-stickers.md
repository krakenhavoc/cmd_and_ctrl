---
title: "The rest of stickers"
date: 2026-10-26
issues: [4307]
---
**The rest of stickers** (#4307, CR 123, CR 702.33h, [ADR 0208](../../decisions/0208-stickers-and-ticket-counters.md)'s 2026-10-26 amendment) — #2138 shipped the sticker seam and two sheets; the other legal sheets, sticker kicker, the borrowed abilities, name-sticker counts and stickers in a graveyard were named as follow-ups. All are built now but two corners.
- **Sheets.** 47 of the 48 Commander-legal sheets carry `Spec.StickerSheet` entries. New vocabulary the abilities needed: `Afflict` and `BattleCry` rows, layer-6 statics inside a sticker's bundle (`noteAppendedGrants`), a granted `DependsOnHandSize` static (Hellbent), four protection qualities (odd and even mana values, two or more creature types, noncreature permanents), the pairwise block rule `ModCantBeBlockedByObject`, and a cast permission derived from a sticker (`StickerAbilityDef.GraveyardCastLife`, claimed as the offer `sticker-graveyard`).
- **Sticker kicker** (CR 702.33h): `StickerKicker`, its own key `game.StickerKickerKey`; `CastSpell` runs the ticket and the sticker before the cast event (Wicker Picker).
- **"Has all abilities of ability stickers"**: `HasAllAbilitiesOfAbilityStickers` and `game.GiveAbilityOfSticker` (CR 123.7a); Pin Collection's free, cost-limited placement rides the prompts' carry word (Pin Collection, Clandestine Chameleon).
- **Name-sticker counts** (CR 123.6d/e): a step record per placement (`StickerPlacedRecordKey`), `Event.Words` on `EventStickerPlaced`, and the eleven `_____` cards.
- **Stickers on cards in a graveyard**: `PutASticker{On: StickerOnCreatureCardInYourGraveyard}` (Scampire), Roxi, Ambassador Blorpityblorpboop, Unlawful Entry, A Good Day to Pie; Done for the Day; Chicken Troupe.
- **Not built.** Squid Fire Knight's sheet (its dice-roll replacement with a second player's choice needs every dice-rolling card to become a continuation), and CR 123.3d (moving a sticker): no Commander-legal card moves one.
