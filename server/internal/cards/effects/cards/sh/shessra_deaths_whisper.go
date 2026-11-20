package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shessra, Death's Whisper — Legendary Creature — Human Elf Warlock
// {2}{B}{G}, 1/3 (Adventures in the Forgotten Realms, S82):
//
//	"Bewitching Whispers — When Shessra enters, target creature blocks this
//	 turn if able.
//	 Whispers of the Grave — At the beginning of your end step, if a creature
//	 died this turn, you may pay 2 life. If you do, draw a card."
//
// The enters trigger is Culvert Ambusher's block requirement (CR 509.1c).
// The end-step trigger is Bulette's intervening "if" (CR 603.4); the life is
// paid, and the card drawn, only on the yes answer, and only if the player
// can pay it (CR 119.4). A player a draw limit stops drawing is not asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb29737c-2dfa-4139-988b-29dc8b829c08",
		Name:         "Shessra, Death's Whisper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Shessra, Death's Whisper — target creature blocks this turn if able",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.BlockRequirementUntilEOT{
							Target: effects.FirstLegalBattlefieldTarget(ctx),
							Kind:   game.BlockRequirementBlocks,
							Label:  "Shessra, Death's Whisper — blocks this turn if able",
						}.Apply(ctx)
					}),
				effects.TargetCreature("target creature")),
			effects.IfVentureCond(effects.CondACreatureDiedThisTurn,
				effects.AtYourEndStep("Shessra, Death's Whisper — you may pay 2 life to draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return effects.MayChoice{
							Question: "Shessra, Death's Whisper — pay 2 life to draw a card?",
							LifeCost: 2,
							Draws:    1,
							Then:     shessraPayLifeAndDraw,
						}.Apply(effects.NewContext(g, item))
					})),
		},
	})
}

// shessraPayLifeAndDraw is "you may pay 2 life. If you do, draw a card",
// run on the yes answer. An on-disk key.
var shessraPayLifeAndDraw = effects.MayYes("confirm/shessra-pay-life-draw", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	controller := ctx.Controller()
	p := ctx.Game.PlayerByIDForEffect(controller)
	if p == nil || p.Eliminated || p.Life < 2 {
		return nil
	}
	if err := ctx.Game.PayLifeForEffect(ctx.Source(), controller, 2); err != nil {
		return err
	}
	return effects.DrawCards{Player: controller, N: 1}.Apply(ctx)
})
