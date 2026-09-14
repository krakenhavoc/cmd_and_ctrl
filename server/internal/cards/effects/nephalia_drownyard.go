package effects

// Nephalia Drownyard — Land (EDHREC rank 3318):
//
//	"{T}: Add {C}.
//	 {1}{U}{B}, {T}: Target player mills three cards."
//
// The Innistrad mill land. The mana ability is a plain colorless
// tap; the mill is a CR 602 activation with a mana and a tap
// component, the player picked at announce and re-checked at
// resolution. Either player is a legal target — milling yourself is
// the graveyard deck's use of it.
//
// The mill is bounded by the target's library (b31MillAtMost): a
// mill does not lose the game, only a draw from an empty library
// does (CR 704.5b), and the engine's mill would otherwise flag the
// loss when it runs a library out.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6429b4ed-1845-4643-9a3d-85f7c12f2bba",
		Name:         "Nephalia Drownyard",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}{B}, {T}: Target player mills three cards.",
			Cost:    Plus(ManaCost("{1}{U}{B}"), TapCost()),
			Targets: TargetPlayer("target player"),
			Effect:  b31MillChosenPlayer(3),
		}},
	})
}
