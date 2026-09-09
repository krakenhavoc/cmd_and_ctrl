package effects

// Palladium Myr — Artifact Creature — Myr {3}, 2/2:
//
//	"{T}: Add {C}{C}."
//
// A mana rock with a body, which makes it a sacrifice-outlet target
// and an artifact-synergy piece as well as ramp. The tap cost is on a
// CREATURE, so summoning sickness applies — unlike Thran Dynamo, this
// does nothing the turn it lands. ActivateManaAbility enforces that
// (CR 302.1) from the source's type line; the spec just declares the
// ability. Haste on the Myr lifts it, same as any tap ability.
func init() {
	Register(Spec{
		OracleID: "7b0767b8-b504-456e-93bd-218502f73b3d",
		Name:     "Palladium Myr",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
	})
}
