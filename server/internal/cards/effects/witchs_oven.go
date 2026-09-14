package effects

// Witch's Oven — Artifact {1} (EDHREC rank 3105):
//
//	"{T}, Sacrifice a creature: Create a Food token. If the
//	 sacrificed creature's toughness was 4 or greater, create two
//	 Food tokens instead. (They're artifacts with "{2}, {T},
//	 Sacrifice this token: You gain 3 life.")"
//
// The one-mana sacrifice outlet that pays in Food. A CR 602 ability
// with a tap and a sacrifice-a-creature cost (Goblin Bombardment's
// clause — the activator's own creatures, the source excluded by
// type); the creature is gone by the time the ability resolves, so
// which one it was is read back off the event log
// (b17PermanentSacrificedToPay, Jarad's read) and its toughness as
// it last stood is its printed toughness plus its counters
// (b29LastKnownToughnessOffBattlefield).
//
// Sandbox simplification, declared, weaker than printed: a static
// toughness bonus from another permanent (an anthem, an Equipment)
// is not in that read, so a 3/3 wearing a +1/+1 anthem makes one
// Food where printed it makes two. Counters are counted.
func init() {
	Register(Spec{
		OracleID:     "8fa0fe02-2452-4386-8e0c-165757b0f0a3",
		Name:         "Witch's Oven",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The sacrificed creature's toughness is read from its printed value and its counters only — a bonus from another permanent isn't counted toward the two-Food threshold."},
		Activated: []ActivatedAbility{{
			Label:  "{T}, Sacrifice a creature: Create a Food token. If the sacrificed creature's toughness was 4 or greater, create two Food tokens instead.",
			Cost:   Plus(TapCost(), SacrificeACreature()),
			Effect: b29FoodForSacrificedCreature,
		}},
	})
}
