# All Will Be Black — deck automation status

Source: [Archidekt deck 15043016](https://archidekt.com/decks/15043016).
Commander: Yawgmoth, Thran Physician.
Tracking: [#2910](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2910),
S58 — Deck requests, October batch follow-up.

## Audit — 2026-10-13

Measured with `deck.ParseArchidekt` and `deckcoverage.Build` using the fresh
deck export, local Scryfall index, and this branch's live catalog. Sideboard
and maybeboard cards are excluded. The report has no unresolved names.

| Coverage bucket | Distinct cards | Copies in the 100-card deck |
| --- | ---: | ---: |
| Declared automated, full | 69 | 69 |
| Automated with caveats | 1 | 1 |
| No effect to implement (Swamp) | 1 | 27 |
| Manual only | 3 | 3 |
| Unreviewed | 0 | 0 |

**The live catalog reports 96 of 100 copies complete on this branch.** The previously merged
19-card batches, review stamps and Forsaken Miner fix are present. Newer
merged deck batches also register Liliana and The Darkness Crystal as Full.
Their tracked simultaneous-sacrifice and paused-life-replacement cases
(#2948 and #2950) still need verification before closing the deck request. This change
completes Nether Traitor (#3417): a separate later death instruction in one
resolution wakes a Traitor already in the graveyard, while an actual shared
wipe or sacrifice does not. It uses existing simultaneous-exit data and
retains the graveyard object and payment guards.

Gisa's implementation is separately tracked in PR #3329. It is not yet on
this branch's audited base; with both changes merged the deck gains one
further complete card. Only the measured branch result is counted here.

## Remaining work

| Card | Tracked gap |
| --- | --- |
| Gisa, the Hellraiser | One ward payment combining mana and life (#2945). |
| Vraska, Betrayal's Sting | Compleated entry loyalty from exact Phyrexian-symbol payments (#2946). |
| Kaervek, the Punisher | Cast a copy of an exiled card during resolution, then continue (#2947). |
| Liliana, Dreadhorde General (cataloged Full) | Verify that all −4 choices are gathered before simultaneous sacrifices (#2948). |
| The Darkness Crystal (cataloged Full) | Verify that additional life gain pauses the original move/effect correctly (#2950). |
| Library of Leng | Choose the relative order of several cards discarded to the library (#2949). |

The deck request remains open. No remaining caveat or manual card has been
relabeled complete to increase the reported percentage.
