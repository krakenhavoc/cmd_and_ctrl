package effects

// Thought Vessel — "You have no maximum hand size. {T}: Add {C}."
//
// Aang deck (Azorius flash/blink) mana rock.
//
// Sandbox simplification: the engine does not enforce a maximum
// hand size at cleanup, so the static half is a no-op today and is
// deliberately NOT declared as a Static ability — there is no
// characteristic to modify. If hand-size enforcement lands, this
// card needs a player-scoped static, which the layer engine does
// not model (it is card-scoped, CR 613).
func init() {
	Register(Spec{
		OracleID: "9965d9c5-2ebf-4a6c-930e-55c5890979be",
		Name:     "Thought Vessel",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
	})
}
