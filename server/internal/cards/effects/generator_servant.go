package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Generator Servant — Creature — Elemental {1}{R}, 2/1:
//
//	"{T}, Sacrifice this creature: Add {C}{C}. If any of that mana is
//	 spent on a creature spell, it gains haste until end of turn.
//	 (That creature can attack and {T} as soon as it comes under your
//	 control.)"
//
// A mana ability with a spend rider that gives the SPELL haste until end
// of turn (ADR 0109 §11 decision 3, #1552): the grant is made where the
// mana is spent, rides the spell onto the permanent (CR 400.7a) and ends
// in the cleanup step like any "until end of turn" effect. Both mana
// come from one activation, so spending one or both on a creature spell
// is one grant. A {T} mana ability of a creature, so the Servant can't
// use it the turn it arrives (CR 302.6).
//
// No simplification is left (#3844, ADR 0068 amendment 2026-10-15). A
// waived payment (strict mana off with too little in the pool, or Cast
// anyway) still spends the share of the cost the pool can pay, so this
// mana pays whenever it can and its rider fires. Mana that never
// entered the pool was not this card's mana.
func init() {
	Register(Spec{
		OracleID:     "68ac061a-e4a3-46df-9169-62f6a8481584",
		Name:         "Generator Servant",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
			Produced: "{C}{C}",
			Label:    "{T}, Sacrifice this creature: Add {C}{C}",
			SpendRiders: []game.ManaSpendRider{
				SpentSpellGains([]string{"haste"}, true, ManaRestrictCast, ManaRestrictType("Creature")),
			},
		}},
	})
}
