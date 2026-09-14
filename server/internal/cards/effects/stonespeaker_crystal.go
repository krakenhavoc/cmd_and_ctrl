package effects

// Stonespeaker Crystal — Artifact {4} (EDHREC rank 3147):
//
//	"{T}: Add {C}{C}.
//	 {2}, {T}, Sacrifice this artifact: Exile any number of target
//	 players' graveyards. Draw a card."
//
// A Hedron Archive that answers graveyards instead of drawing two.
// The mana ability is a plain Sol Ring shape; the sacrifice ability
// is Tormod's Crypt widened to "any number of target players" — a
// zero-to-unbounded player clause, so the activator may name nobody,
// everyone, or just the reanimator — with the draw after the exile,
// in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a494fcee-6885-434c-aad5-6f83640c4472",
		Name:         "Stonespeaker Crystal",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice Stonespeaker Crystal: Exile any number of target players' graveyards. Draw a card.",
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Targets: TargetPlayer("any number of target players").WithCount(0, 0),
			Effect:  b30ExileTargetGraveyardsThenDraw,
		}},
	})
}
