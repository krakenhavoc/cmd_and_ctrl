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
| Automated, complete | 67 | 67 |
| Automated with caveats | 2 | 2 |
| No effect to implement (Swamp) | 1 | 27 |
| Manual only | 4 | 4 |
| Unreviewed | 0 | 0 |

**94 of 100 copies are complete on this branch.** The previously merged
19-card batches, review stamps, and Forsaken Miner crime fix are present.
This change adds Gisa, the Hellraiser with one payment for both ward costs,
her once-per-turn crime trigger, two tapped blue and black Zombie Rogue
tokens, and the Skeleton/Zombie anthem and menace grant (#2945).
The shared ward also removes Captain Howler's last caveat outside this deck.

## Remaining work

| Card | Tracked gap |
| --- | --- |
| Vraska, Betrayal's Sting | Compleated entry loyalty from exact Phyrexian-symbol payments (#2946). |
| Kaervek, the Punisher | Cast a copy of an exiled card during resolution, then continue (#2947). |
| Liliana, Dreadhorde General | Gather every affected player's choices before simultaneous sacrifices (#2948). |
| The Darkness Crystal | A move replacement with resumable additional life gain and linked reanimation (#2950). |
| Library of Leng | Choose the relative order of several cards discarded to the library (#2949). |
| Nether Traitor | Distinguish separate death instructions in one resolution from simultaneous deaths. |

The deck request remains open. No remaining caveat or manual card has been
relabeled complete to increase the reported percentage.
