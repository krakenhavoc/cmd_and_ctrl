package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Banish from Edoras — Sorcery {4}{W} (The Lord of the Rings: Tales of
// Middle-earth):
//
//	"This spell costs {2} less to cast if it targets a tapped creature.
//	 Exile target creature."
//
// Ride's End's reduction (CostsLessIfItTargets) over a tapped creature,
// judged against the announced target as the cost is determined
// (CR 601.2f). The exile is exileFirstLegalCardTarget: a target that is
// no longer legal on resolution is left alone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4508afd-7cc3-4176-8291-cfc2c155ac11",
		Name:         "Banish from Edoras",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a tapped creature.",
				effects.And(effects.Creature(), effects.TappedPermanent())),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
