package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vivid Revival — Sorcery {4}{G} (Guilds of Ravnica, #5525):
//
//	"Return up to three target multicolored cards from your graveyard to your
//	 hand. Exile Vivid Revival."
//
// Reborn Hope's multicolored test (two or more colours) over up to three
// targets, and All Suns' Dawn's "Exile ~" tail: the spell exiles itself as it
// finishes resolving, so it goes to exile rather than the graveyard. A target
// that has left the graveyard is skipped (CR 608.2b); if every target is
// illegal the spell does not resolve and goes to the graveyard instead.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5d856a3-0305-4c3b-b447-99570d6e8195",
		Name:         "Vivid Revival",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets: effects.TargetCardInGraveyard("up to three target multicolored cards from your graveyard",
			effects.YouOwn(), effects.Multicolored()).WithCount(0, 3),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.ReturnLegalGraveyardTargetsToHand),
			effects.Instruction(func(ctx *effects.Context) error { return effects.ExileTarget{Target: ctx.Source()}.Apply(ctx) }),
		),
	})
}
