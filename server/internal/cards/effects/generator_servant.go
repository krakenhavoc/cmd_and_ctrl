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
// One declared simplification, the rider's: with strict mana off the
// engine never spends the pool, so no rider fires (ADR 0068 §3).
func init() {
	Register(Spec{
		OracleID:     "68ac061a-e4a3-46df-9169-62f6a8481584",
		Name:         "Generator Servant",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't see which mana you spent, so the creature never gains haste."},
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
