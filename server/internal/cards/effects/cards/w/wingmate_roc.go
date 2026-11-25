package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wingmate Roc — Creature — Bird {3}{W}{W}, 3/4 (Commander 2019):
//
//	"Flying
//	 Raid — When this creature enters, if you attacked this turn, create a 3/4
//	 white Bird creature token with flying.
//	 Whenever this creature attacks, you gain 1 life for each attacking
//	 creature."
//
// The raid trigger is Rose, Cutthroat Raider's: a CR 603.4 intervening "if"
// read from the turn's attack tally as the creature enters and again as the
// ability resolves, so a Roc cast before combat makes nothing. The life is read
// as the attack trigger resolves: every creature on the battlefield attacking
// at that moment, whoever controls it, the Roc included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5badc2e8-a7d6-48c7-b594-65db953d9d57",
		Name:            "Wingmate Roc",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you attacked this turn",
				Test: func(ctx *effects.Context) bool { return effects.B18AttackedThisTurn(ctx.Game, ctx.Controller()) },
			}, effects.On(game.EventETB, effects.AllOf(effects.Self, effects.TmcB6YouAttackedThisTurn),
				"Wingmate Roc — raid: create a 3/4 white Bird creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/4 white Bird with flying"), N: 1}))),
			effects.WheneverThisAttacks("Wingmate Roc — you gain 1 life for each attacking creature",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCreature() && c.AttackingTarget != uuid.Nil {
							n++
						}
					}
					return effects.GainLife{Amount: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
