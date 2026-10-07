package effects

// Orazca Relic — Artifact {3}:
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 {T}: Add {C}.
//	 {T}, Sacrifice this artifact: You gain 3 life and draw a card.
//	 Activate only if you have the city's blessing."
//
// DECLARED SIMPLIFICATION (weaker than printed): the engine has no
// "city's blessing" player designation, so the gate reads "you control
// ten or more permanents" live, exactly as Arch of Orazca does. The
// blessing is permanent once earned; this read is not, so the Relic
// stops being activatable if permanents drop below ten again. Never
// stronger than printed. The Relic itself counts toward the ten, as it
// does on the real card (it is a permanent until the cost sacrifices it).
func init() {
	Register(Spec{
		OracleID:     "48b84b58-1a06-4bb4-be1f-ad3ca69e66dc",
		Name:         "Orazca Relic",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The city's blessing isn't kept once earned — the sacrifice ability works only while you control ten or more permanents."},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:     "{T}, Sacrifice this artifact: You gain 3 life and draw a card. Activate only if you have the city's blessing.",
			Cost:      Plus(TapCost(), SacrificeThis()),
			Condition: archOfOrazcaCitysBlessing,
			Effect:    Do(GainLife{Amount: 3}, DrawCards{N: 1}),
		}},
	})
}
