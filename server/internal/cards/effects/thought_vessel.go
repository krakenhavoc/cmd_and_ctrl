package effects

// Thought Vessel — "You have no maximum hand size. {T}: Add {C}."
//
// Aang deck (Azorius flash/blink) mana rock.
//
// Both halves are implemented. The static half is declared through
// Spec.NoMaxHandSize rather than Spec.Static because it is
// player-scoped: the CR 613 layer engine modifies characteristics of
// objects, and a player's maximum hand size is not one. The engine
// derives the answer from the battlefield at cleanup instead of
// writing to Player.MaxHandSize, so two copies compose and one
// leaving while another remains does not strand the controller back
// at a maximum. See Spec.NoMaxHandSize and
// game.Game.EffectiveMaxHandSizeLocked.
//
// Issue #338: this shipped with a "the engine does not enforce a
// maximum hand size at cleanup" note that was already stale —
// enforcement landed in S13.4 — and a player was made to discard at
// their end step with a Thought Vessel on the battlefield.
func init() {
	Register(Spec{
		OracleID:      "9965d9c5-2ebf-4a6c-930e-55c5890979be",
		Name:          "Thought Vessel",
		NoMaxHandSize: true,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
	})
}
