package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warrant // Warden — split card {W/U}{W/U} // {3}{W}{U} (Ravnica
// Remastered):
//
//	Warrant (Instant): "Put target attacking or blocking creature on top of
//	 its owner's library."
//	Warden (Sorcery): "Create a 4/4 white and blue Sphinx creature token with
//	 flying and vigilance."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19053c47-97ce-4a25-a92c-1f4061e85bf3",
		Name:         "Warrant",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		OnResolve:    effects.GapB79PutTargetOnTopOfItsOwnersLibrary,
	})
	effects.Register(effects.Spec{
		OracleID:     "19053c47-97ce-4a25-a92c-1f4061e85bf3#1",
		Name:         "Warden",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("4/4 white and blue Sphinx with flying and vigilance"), N: 1}),
	})
}
