package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// March of Wretched Sorrow — Instant {X}{B} (Kamigawa: Neon Dynasty):
//
//	"As an additional cost to cast this spell, you may exile any number
//	 of black cards from your hand. This spell costs {2} less to cast for
//	 each card exiled this way.
//	 March of Wretched Sorrow deals X damage to target creature or
//	 planeswalker and you gain X life."
//
// The cost is March of Swirling Mist's: ExileAnyNumberFromHandCost plus
// CostsLessPerExiledFromHand ({2} off per card at CR 601.2f; the announced X
// stays X). The life is the announced X, not the damage dealt, so it is
// gained in full when the damage is prevented; a target that became illegal
// stops the whole spell (CR 608.2b), so no life is gained then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "6e638332-fba1-4e91-9237-fdff23826640",
		Name:           "March of Wretched Sorrow",
		Completeness:   effects.CompletenessFull,
		XMatters:       true,
		AdditionalCost: effects.ExileAnyNumberFromHandCost("black cards", "B"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessPerExiledFromHand("{2}", "This spell costs {2} less to cast for each card exiled this way"),
		},
		Targets: effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose: effects.ForTargets(effects.DamageXToTarget(0)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DamageXToTheTargetAndGainX(item, ctx, ctx.X())
		},
	})
}
