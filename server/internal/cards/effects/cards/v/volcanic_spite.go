package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volcanic Spite — Instant {1}{R} (March of the Machine):
//
//	"Volcanic Spite deals 3 damage to target creature, planeswalker, or
//	 battle. You may put a card from your hand on the bottom of your
//	 library. If you do, draw a card."
//
// The rummage is Manhole Missile's (manholeMissileTucked: the pick goes
// to the bottom and the draw is told whether it got there). A spell whose
// only target is illegal on resolution does nothing at all (CR 608.2b), the
// rummage included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10426070-2e6f-47a5-bc8d-16dcfb058b26",
		Name:         "Volcanic Spite",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.S82MomB12CreaturePlaneswalkerOrBattle("target creature, planeswalker, or battle"),
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
					Question: "Volcanic Spite — you may put a card from your hand on the bottom of your library. If you do, draw a card",
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
