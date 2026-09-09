package effects

// Ashnod's Altar — Artifact {3}:
//
//	"Sacrifice a creature: Add {C}{C}."
//
// One of the format's foundational combo pieces, and the reason the
// mana-ability cost model needed a sacrifice-another clause: every
// other sacrifice outlet in the catalog either eats itself (Treasure)
// or uses the stack (Goblin Bombardment). The Altar does neither — it
// is a MANA ability, so it resolves immediately without the stack
// (CR 605.3a) and can be activated any number of times while
// something else is resolving.
//
// That immediacy is exactly what makes it a combo engine: with a
// token producer and a payoff, the Altar converts creatures into mana
// faster than anyone can respond, because there is no window to
// respond in. Paired with Pitiless Plunderer it goes mana-positive.
//
// No tap component, so the Altar works the turn it lands and any
// number of times per turn.
func init() {
	Register(Spec{
		OracleID: "4d18bcba-a346-445e-a182-6cc30b7e066d",
		Name:     "Ashnod's Altar",
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				SacrificeOther: SacrificeACreature().SacrificeOther,
			},
			Produced: "{C}{C}",
			Label:    "Sacrifice a creature: Add {C}{C}",
		}},
	})
}
