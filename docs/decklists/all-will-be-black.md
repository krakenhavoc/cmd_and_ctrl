# All Will Be Black — deck automation status

Source: [Archidekt deck 15043016](https://archidekt.com/decks/15043016).
Commander: Yawgmoth, Thran Physician.
Tracking: [#2910](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2910),
S58 — Deck requests, October batch follow-up.

## Audit — 2026-10-14

Measured with `deck.ParseArchidekt` and `deckcoverage.Build` against a fresh
Archidekt export, the real Scryfall index and the catalog on this branch,
including merged Gisa PR #3329, Nether Traitor PR #3418 and this Liliana fix. Sideboard and
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
are merged (#2945, PR #3329), as is Nether Traitor's distinction between
successive and simultaneous deaths (#3417, PR #3418). This change completes
Liliana's public APNAP choices and one shared sacrifice for both loyalty
abilities (#2948), retaining forced-sacrifice protection and multitype choices.

The Darkness Crystal's tracked behavior gap still reproduces: its life gain
can pause while the original effect continues (#2950), despite its Full stamp.
That issue remains open. Catalog declarations alone do not close the request.

## Remaining work

| Card | Tracked gap |
| --- | --- |
| Vraska, Betrayal's Sting | Compleated entry loyalty from precise Phyrexian-symbol payment provenance (#2946). |
| Kaervek, the Punisher | Cast a copy of an exiled card during resolution, then continue (#2947). |
| Library of Leng | Choose the relative order of several cards discarded to the library (#2949). |
| The Darkness Crystal (cataloged Full) | Wait for additional life gain before the enclosing effect resumes (#2950). |

The deck request remains open until these printed behaviors are complete.
