package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Militia's Pride — Kindred Enchantment — Kithkin {1}{W} (Lorwyn, S82 OLD,
// #5593):
//
//	"Whenever a nontoken creature you control attacks, you may pay {W}. If you
//	 do, create a 1/1 white Kithkin Soldier creature token that's tapped and
//	 attacking."
//
// Mardu Ascendancy's per-attacker trigger over nontoken creatures, with the
// payment asked as it resolves. The token enters tapped attacking the player the
// creature attacked and was never declared as an attacker (CR 508.4), so it
// triggers nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d311154b-4742-49e5-9b65-4de2d3e0d50b",
		Name:         "Militia's Pride",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ANontokenCreatureYouControlAttacked, "Militia's Pride — you may pay {W} to create a tapped and attacking 1/1 white Kithkin Soldier",
				effects.S82OldB554MayPayForAnAttackingKithkin),
		},
	})
}
