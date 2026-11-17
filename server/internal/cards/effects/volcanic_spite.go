package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "10426070-2e6f-47a5-bc8d-16dcfb058b26",
		Name:         "Volcanic Spite",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      s82MomB12CreaturePlaneswalkerOrBattle("target creature, planeswalker, or battle"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return DealDamageToTheTarget(3)(ctx.Game, ctx.Item)
			}),
			Instruction(func(ctx *Context) error {
				player := ctx.Controller()
				hand := allHandCardIDs(ctx.Game, player)
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
					ThenKey:  manholeMissileTucked,
				})
				return nil
			}),
		),
	})
}
