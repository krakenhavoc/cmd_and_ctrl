package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mine Collapse — Instant {3}{R} (Modern Horizons 2, #5501):
//
//	"If it's your turn, you may sacrifice a Mountain rather than pay this
//	 spell's mana cost.
//	 Mine Collapse deals 5 damage to target creature or planeswalker."
//
// Crash's sacrifice alternative cost (SacrificeInstead, #1727) with the
// offer gated on its Condition, as Snuff Out's Swamp gates its life
// payment: only the active player is offered it, at announce and in the
// view. The Mountain is sacrificed as a cost with the spell already on
// the stack (CR 601.2h).
//
// No simplification.
func init() {
	sacrificeAMountain := effects.SacrificeInstead(1, "a Mountain", 0, effects.HasSubtype("Mountain"))
	sacrificeAMountain.Condition = effects.IsActivePlayer
	effects.Register(effects.Spec{
		OracleID:         "83df0f44-86ca-4adb-9f3a-da1ba6dd7a6d",
		Name:             "Mine Collapse",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:          game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		AlternativeCosts: []game.AlternativeCost{sacrificeAMountain},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DealDamageToTheTarget(5)(ctx.Game, item)
		},
	})
}
