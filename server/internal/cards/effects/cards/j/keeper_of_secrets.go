package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keeper of Secrets — Creature — Demon {5}{R}, 6/4 (Warhammer 40,000
// Commander, S82 GAP batch 194, #5948):
//
//	"First strike, haste
//	 Symphony of Pain — Whenever you cast a spell from anywhere other than your
//	 hand, this creature deals damage equal to that spell's mana value to target
//	 opponent."
//
// Vega's cast-from-not-hand reading (the cast event's origin zone), with the
// opponent chosen as the trigger goes on the stack. The damage is the spell's
// mana value as it was cast (an X spell counts its X, CR 202.3e), read from
// last-known information if the spell has already left the stack, and is the
// Keeper's own damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "23f86ee2-b1e3-45c2-bfd5-044919033922",
		Name:            "Keeper of Secrets",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventCast, effects.GapB194CastFromAnywhereButYourHand,
					"Keeper of Secrets — damage equal to that spell's mana value to target opponent",
					effects.GapB194DamageEqualToTheCastSpellsManaValue),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
