# Mishra — deck automation status

Source: [Archidekt deck 27015788](https://archidekt.com/decks/27015788).
Tracking: S58 — Deck requests, October batch ([#2077](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2077)),
[Mishra #2033](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2033).

## Audit — 2026-10-12

Measured with `deck.ParseArchidekt` and `deckcoverage.Build` against the local
Scryfall bulk index and this branch's live effects catalog. Sideboard and
maybeboard entries are excluded by the existing parser.

| Coverage bucket | Distinct cards | Copies in the 100-card deck |
| --- | ---: | ---: |
| Automated, complete | 89 | 89 |
| Automated with caveats | 1 | 1 |
| No effect to implement (basic lands) | 3 | 10 |
| Manual only | 0 | 0 |
| Unreviewed | 0 | 0 |

All 100 cards are recognized, with no manual-only or unreviewed cards.
**All printed rules in the deck work with recorded mana payments.** The
catalog reports 99 of 100 copies as complete, including the basic lands;
Path of Ancestry retains the required caveat for unrecorded sandbox payments.
A fully funded pool is recorded in permissive mode too. An insufficient
permissive pool or “Cast anyway” deliberately bypasses payment and cannot
truthfully fire a mana-spend rider.

### Completed in this change

- **Goblin Welder:** legal paired targeting and the simultaneous sacrifice
  and return, through entry/exit replacements and undo. Shared target seam
  [#1943](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1943) also enables
  Fiery Annihilation outside this deck.
- **Deadly Rollick:** the existing commander-dependent free-cast alternative.
- **Saheeli, the Gifted:** copies run their entry hooks; haste and delayed
  exile wait for token creation and entry choices to finish. Haste is granted
  rather than becoming copiable printed text.
- **Portal to Phyrexia:** the returned creature becomes a Phyrexian even when
  entry pauses for a choice; the continuation follows undo correctly.

- **Cybermen Squadron:** recipient-owned myriad grants, separate optional
  decisions per opponent, player/planeswalker destinations for each doubled
  copy, simultaneous entries, and copy information for a departed attacker.
- **Notion Thief:** only the first actual draw in each opponent's own draw
  step is exempt; doubled draws and dredge substitutes are handled per card.
- **Path of Ancestry:** funded permissive casts now spend real mana and fire
  the scry rider when its mana funds a matching creature. Paper overrides
  remain explicit. Shared work also completes Alhammarret's Archive,
  Orcish Bowmasters, Legion Loyalty and Teferi's Ageless Insight.

Karlach's Background and Sakashima's Partner support already exist on the
current integration branch and are not counted as new work here.

### Remaining sandbox caveat

| Card | Deliberate sandbox override |
| --- | --- |
| Path of Ancestry | Scry 1 requires recorded spending. It does not fire when an insufficient pool is paid on paper or the player chooses “Cast anyway.” |

This caveat is retained under the payment contract in
[ADR 0068](../decisions/0068-the-mana-spent-on-a-spell.md). It describes a
manual payment override, not an unimplemented card ability. The catalog
has not been relabeled to conceal that distinction.
