package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "19053c47-97ce-4a25-a92c-1f4061e85bf3",
		Name:         "Warrant",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		OnResolve:    gapB79PutTargetOnTopOfItsOwnersLibrary,
	})
	Register(Spec{
		OracleID:     "19053c47-97ce-4a25-a92c-1f4061e85bf3#1",
		Name:         "Warden",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve:    ResolveSteps(CreateToken{Template: TokenCard("4/4 white and blue Sphinx with flying and vigilance"), N: 1}),
	})
}
