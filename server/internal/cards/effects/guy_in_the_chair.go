package effects

// Guy in the Chair — Creature — Human Advisor {2}{G}, 2/3:
//
//	"{T}: Add one mana of any color.
//	 Web Support — {2}{G}, {T}: Put a +1/+1 counter on target Spider.
//	 Activate only as a sorcery."
//
// The mana ability is the Birds shape (any colour). Web Support is an
// ability word, so the activation is an ordinary sorcery-speed
// ability with a mana-and-tap cost whose target is any Spider
// (Hall of Oracles' counter body); a Spider that has left is skipped
// at resolution. No simplification.
func init() {
	Register(Spec{
		OracleID:     "83eb74df-1c77-48b0-83be-55d7aa4d3618",
		Name:         "Guy in the Chair",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:        "Web Support — {2}{G}, {T}: Put a +1/+1 counter on target Spider. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{G}"), TapCost()),
			Targets:      TargetCreature("target Spider", OfCreatureType("Spider")),
			SorcerySpeed: true,
			Effect:       b36CounterOnChosenAnimal,
		}},
	})
}
