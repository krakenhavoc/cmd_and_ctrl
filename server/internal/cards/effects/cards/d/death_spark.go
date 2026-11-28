package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Death Spark — Instant {R} (Alliances, S82 OLD batch 819):
//
//	"Death Spark deals 1 damage to any target.
//	 At the beginning of your upkeep, if this card is in your graveyard with a
//	 creature card directly above it, you may pay {1}. If you do, return this
//	 card to your hand."
//
// The graveyard is a pile, and "directly above" is the card put on top of it
// next (OldB819CreatureCardDirectlyAbove). The condition is an intervening
// "if" (CR 603.4), so it is checked as the ability triggers and as it
// resolves; the {1} is paid on resolution.
func init() {
	const label = "Death Spark — you may pay {1}; if you do, return this card to your hand"
	upkeep := effects.InterveningIf(effects.OldB819GraveyardCreatureAboveCondition,
		effects.On(game.EventBeginUpkeep, effects.AllOf(effects.ByYou, effects.OldB819InGraveyardUnderACreature), label,
			func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{Chooser: item.Controller, Cost: "{1}", Question: label, Then: deathSparkReturn}.Apply(effects.NewContext(g, item))
			}))
	upkeep.Zones = []game.ZoneKind{game.ZoneGraveyard}
	effects.Register(effects.Spec{
		OracleID:     "4b23a89c-89ba-41ad-b894-ec02221689eb",
		Name:         "Death Spark",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 1)),
		Targets:      effects.TargetAny(),
		OnResolve:    effects.DamageToFirstTargetStep(1),
		Triggered:    []game.TriggeredAbility{upkeep},
	})
}

var deathSparkReturn = effects.PayUnlessBranches("pay-unless/death-spark-return", func(ctx *effects.Context, r game.PayUnlessAnswer) error {
	return effects.OldB819ReturnThisFromGraveyardToHand(ctx, r.Source)
}, nil)
