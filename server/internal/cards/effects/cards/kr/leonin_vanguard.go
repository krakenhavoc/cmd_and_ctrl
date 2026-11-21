package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leonin Vanguard — Creature — Cat Soldier {W}, 1/1 (Foundations, #4465):
//
//	"At the beginning of combat on your turn, if you control three or
//	 more creatures, this creature gets +1/+1 until end of turn and you
//	 gain 1 life."
//
// Territorial Witchstalker's beginning-of-combat trigger with Dundoolin
// Weaver's count as the intervening if (CR 603.4): checked as combat
// begins and again as the trigger resolves, the Vanguard counting itself.
// A Vanguard that has left the battlefield gets no pump, and the life is
// gained either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de1e1352-adb2-4ac1-bb2a-ee654d5c727d",
		Name:         "Leonin Vanguard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control three or more creatures",
				Test: func(ctx *effects.Context) bool {
					return effects.DundoolinWeaverThreeCreatures(ctx.Game, ctx.Controller())
				},
			}, effects.On(game.EventStepBegan, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.StepBegan(game.StepBeginCombat, true)(ev, source, lki, g) &&
					effects.DundoolinWeaverThreeCreatures(g, source.Controller)
			}, "Leonin Vanguard — +1/+1 until end of turn and you gain 1 life",
				func(g *game.Game, item *game.StackItem) error {
					if effects.B09SourceStillOnBattlefield(g, item) {
						if err := effects.ThisGetsUntilEndOfTurn(1, 1, "Leonin Vanguard — +1/+1 until end of turn")(g, item); err != nil {
							return err
						}
					}
					return effects.GainLife{Player: item.Controller, Amount: 1}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
