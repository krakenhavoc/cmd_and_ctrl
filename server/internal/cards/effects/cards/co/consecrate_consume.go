package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Consecrate // Consume — split card {1}{W/B} // {2}{W}{B} (Ravnica Allegiance):
//
//	Consecrate (Instant): "Exile target card from a graveyard. Draw a card."
//	Consume (Sorcery): "Target player sacrifices a creature with the greatest power
//	 among creatures they control. You gain life equal to its power."
//
// Consecrate does nothing at all if its target is gone (CR 608.2b), the draw
// included. Consume lets the target player pick among their creatures tied for
// the greatest power; the life follows the sacrifice, equal to that creature's
// power as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20e7a93f-77ce-466b-8586-35d390689d0c",
		Name:         "Consecrate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target card from a graveyard"),
		Purpose:      game.Purpose{Answers: game.AnswerGraveyard, Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok {
				return nil
			}
			if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "20e7a93f-77ce-466b-8586-35d390689d0c#1",
		Name:         "Consume",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.OldB178EdictGreatestPowerGainPower(ctx, "Consume")
		},
	})
}
