package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aura Mutation — Instant {G}{W} (Invasion, #5616):
//
//	"Destroy target enchantment. Create X 1/1 green Saproling creature
//	 tokens, where X is that enchantment's mana value."
//
// Artifact Mutation's body (destroyTargetThenSaprolingsForItsManaValue)
// on an enchantment, and without the regeneration clause. X is the
// enchantment's mana value just before it is destroyed (CR 608.2h); the
// Saprolings are the caster's, and an illegal target means no tokens
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "19bb51dd-2b93-45f0-927d-bbafddadd0a3",
		Name:         "Aura Mutation",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target enchantment", Enchantment()),
		OnResolve:    destroyTargetThenSaprolingsForItsManaValue(false),
	})
}
