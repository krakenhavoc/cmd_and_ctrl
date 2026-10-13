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
| Automated, complete | 66 | 66 |
| Automated with caveats | 2 | 2 |
| No effect to implement (Swamp) | 1 | 27 |
| Manual only | 5 | 5 |
| Unreviewed | 0 | 0 |

**93 of 100 copies are complete on this branch.** The previously merged
19-card batches and two review stamps are present. This change adds
Forsaken Miner through the corrected crime-announcement event (#2944).
Copies and retargeting still cause ordinary target events, but do not commit
new crimes. One qualifying announcement commits one crime, even if it has
several qualifying targets, including opponents' stack abilities.

## Remaining work

| Card | Tracked gap |
| --- | --- |
| Gisa, the Hellraiser | One ward payment combining mana and life (#2945). |
| Vraska, Betrayal's Sting | Compleated entry loyalty from exact Phyrexian-symbol payments (#2946). |
| Kaervek, the Punisher | Cast a copy of an exiled card during resolution, then continue (#2947). |
| Liliana, Dreadhorde General | Gather every affected player's choices before simultaneous sacrifices (#2948). |
| The Darkness Crystal | A move replacement with resumable additional life gain and linked reanimation (#2950). |
| Library of Leng | Choose the relative order of several cards discarded to the library (#2949). |
| Nether Traitor | Distinguish separate death instructions in one resolution from simultaneous deaths. |

The deck request remains open. No remaining caveat or manual card has been
relabeled complete to increase the reported percentage.
