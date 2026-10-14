# All Will Be Black — deck automation status

Source: [Archidekt deck 15043016](https://archidekt.com/decks/15043016).
Commander: Yawgmoth, Thran Physician.
Tracking: [#2910](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2910),
S58 — Deck requests, October batch follow-up.

## Audit — 2026-10-13

Measured with `deck.ParseArchidekt` and `deckcoverage.Build` against a fresh
Archidekt export, the real Scryfall index and the catalog on this branch,
including merged Gisa PR #3329 and this Nether Traitor fix. Sideboard and
maybeboard cards are excluded. There are no unresolved names or violations.

| Coverage bucket | Distinct cards | Copies in the 100-card deck |
| --- | ---: | ---: |
| Declared automated, full | 70 | 70 |
| Automated with caveats | 1 | 1 |
| No effect to implement (Swamp) | 1 | 27 |
| Manual only | 2 | 2 |
| Unreviewed | 0 | 0 |

**The live catalog reports 97 of 100 copies complete on this branch.**
Gisa's compound ward, crime trigger, tapped two-color tokens and tribal buffs
are merged (#2945, PR #3329). This change completes Nether Traitor (#3417):
separate death instructions in one resolution trigger it when it is already
in the graveyard, while an actual shared wipe or sacrifice does not.

Newer deck batches also register Liliana and The Darkness Crystal as Full.
Their tracked behavior gaps still reproduce in focused tests: Liliana's −4
moves creatures before every player chooses (#2948); the Crystal's life
gain can pause while the original effect continues (#2950). Those issues
remain open. Catalog declarations alone do not close the deck request.

## Remaining work

| Card | Tracked gap |
| --- | --- |
| Vraska, Betrayal's Sting | Compleated entry loyalty from precise Phyrexian-symbol payment provenance (#2946). |
| Kaervek, the Punisher | Cast a copy of an exiled card during resolution, then continue (#2947). |
| Library of Leng | Choose the relative order of several cards discarded to the library (#2949). |
| Liliana, Dreadhorde General (cataloged Full) | Gather all −4 choices before one simultaneous sacrifice; preserve protection and choice order (#2948). |
| The Darkness Crystal (cataloged Full) | Wait for additional life gain before the enclosing effect resumes (#2950). |

The deck request remains open until these printed behaviors are complete.
