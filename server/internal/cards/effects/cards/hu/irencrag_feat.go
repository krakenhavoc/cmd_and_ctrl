package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Irencrag Feat — Sorcery {1}{R}{R}{R} (Throne of Eldraine, S82 old sets):
//
//	"Add seven {R}. You can cast only one more spell this turn."
//
// A sorcery, so it uses the stack (CR 605.3a). The mana is added as it
// resolves; the second sentence is a granted cast ban for the rest of the
// turn (CR 611.2, the CastBanMaxPerTurn seam): its cap is the number of
// spells its controller has cast this turn, Irencrag Feat included, plus
// one, so exactly one more spell can be cast. Only its controller is held to
// it, and a later copy of the ban can only lower the cap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d550e17-7340-4b11-8b25-1d4f1a58f82f",
		Name:         "Irencrag Feat",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: effects.ResolveSteps(
			effects.AddMana{Amount: effects.ManaN(7, "R")},
			effects.Instruction(func(ctx *effects.Context) error {
				you := ctx.Controller()
				ctx.Game.GrantCastBanForEffect(you, game.CastBanRule{
					Kind:       game.CastBanMaxPerTurn,
					MaxPerTurn: ctx.Game.CastTallyFor(you).Total + 1,
				}, "Irencrag Feat — you can cast only one more spell this turn", ctx.Source(), game.Duration{})
				return nil
			}),
		),
	})
}
