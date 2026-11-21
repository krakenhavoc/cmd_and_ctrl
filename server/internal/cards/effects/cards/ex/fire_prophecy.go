package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Prophecy — Instant {1}{R}:
//
//	"Fire Prophecy deals 3 damage to target creature. You may put a card
//	 from your hand on the bottom of your library. If you do, draw a
//	 card."
//
// Volcanic Spite's shape with a creature-only target: the rummage is
// Manhole Missile's (manholeMissileTucked). A spell whose only target
// is illegal on resolution does nothing at all (CR 608.2b), the rummage
// included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5e51171-2e36-4b09-b43d-e5d437a19305",
		Name:         "Fire Prophecy",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DealDamageToTheTarget(3)(ctx.Game, ctx.Item)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				player := ctx.Controller()
				hand := effects.AllHandCardIDs(ctx.Game, player)
				if len(hand) == 0 {
					return nil
				}
				ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:  player,
					Source:   ctx.Source(),
					Question: "Fire Prophecy — you may put a card from your hand on the bottom of your library. If you do, draw a card",
					Cards:    hand,
					Min:      0,
					Max:      1,
					LiveZone: game.ZoneHand,
					ThenKey:  effects.ManholeMissileTucked,
				})
				return nil
			}),
		),
	})
}
