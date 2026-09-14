package effects

// Oswald Fiddlebender — Legendary Creature — Gnome Artificer {1}{W},
// 2/2 (EDHREC rank 2891):
//
//	"Magical Tinkering — {W}, {T}, Sacrifice an artifact: Search your
//	 library for an artifact card with mana value equal to 1 plus the
//	 sacrificed artifact's mana value, put it onto the battlefield,
//	 then shuffle. Activate only as a sorcery."
//
// Birthing Pod on a Gnome, for artifacts. The cost is mana, tap and
// "sacrifice an artifact" (b10SacrificeAnArtifact — any artifact the
// activator controls, a Treasure or a Food included), paid at
// announce; the search reads the sacrificed artifact's mana value
// back off the event log (b17PermanentSacrificedToPay), accepts only
// artifact cards at exactly that value plus one, and puts the pick
// onto the battlefield untapped. The tap is a {T} on a creature, so
// summoning sickness applies (CR 302.6) — the engine enforces it.
// Sorcery speed, as printed.
//
// Rides the catalog-wide deterministic search pick when only one
// card qualifies — the S22 chooser asks only when there is a choice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dbc9ea19-cf68-41d3-88a8-b5ab8df75c5a",
		Name:         "Oswald Fiddlebender",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{W}, {T}, Sacrifice an artifact: Search your library for an artifact card with mana value equal to 1 plus the sacrificed artifact's mana value, put it onto the battlefield, then shuffle. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{W}"), TapCost(), b10SacrificeAnArtifact()),
			SorcerySpeed: true,
			Effect:       b27OswaldSearch,
		}},
	})
}
