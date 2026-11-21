package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Imposing Grandeur — Sorcery {4}{R} (Innistrad: Crimson Vow):
//
//	"Each player may discard their hand and draw cards equal to the
//	 greatest mana value of a commander they own on the battlefield or in
//	 the command zone."
//
// Every player chooses before anything happens, in APNAP order (the shared
// each-player-may wheel, CR 101.4). Then every yes-sayer discards their
// hand, and only then does each draw their own number of cards: the
// greatest mana value among the commanders THEY own that are on the
// battlefield or in the command zone, read as the draws happen (a commander
// in hand, graveyard or exile is not counted; a commander that is a copy on
// the battlefield counts at its copied mana value; a token never is one).
// A player with no such commander may still discard their hand and draws
// zero.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9ef02d6-717d-4f2a-aa10-0af717d6360f",
		Name:         "Imposing Grandeur",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return imposingGrandeurWheel.Start(ctx)
		},
	})
}

var imposingGrandeurWheel = effects.NewMayWheelDrawing("confirm/imposing-grandeur-may-wheel",
	"Imposing Grandeur — discard your hand and draw cards equal to the greatest mana value of a commander you own on the battlefield or in the command zone?",
	0, effects.RegisterThenSteps("then-steps/imposing-grandeur-wheel",
		effects.Instruction(func(ctx *effects.Context) error {
			for _, id := range ctx.StepRecord(effects.MayWheelYesKey).IDs {
				if _, err := effects.DiscardWholeHand(ctx.Game, id); err != nil {
					return err
				}
			}
			return nil
		}),
		effects.Instruction(func(ctx *effects.Context) error {
			for _, id := range ctx.StepRecord(effects.MayWheelYesKey).IDs {
				p := ctx.Game.PlayerByIDForEffect(id)
				if p == nil || p.Eliminated {
					continue
				}
				n := imposingGrandeurCommanderValue(ctx.Game, id)
				if n <= 0 {
					continue
				}
				if err := ctx.Game.DrawNForEffect(id, n); err != nil {
					return err
				}
			}
			return nil
		}),
	))

// imposingGrandeurCommanderValue is the greatest mana value of a commander
// `player` owns on the battlefield or in their command zone.
func imposingGrandeurCommanderValue(g *game.Game, player uuid.UUID) int {
	best := 0
	consider := func(c game.Card) {
		if c.IsCommander && c.Owner == player && !c.IsToken() {
			if mv := c.ManaValue(); mv > best {
				best = mv
			}
		}
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		consider(c)
	}
	if p := g.PlayerByIDForEffect(player); p != nil {
		for _, c := range p.Command.Cards {
			consider(c)
		}
	}
	return best
}
