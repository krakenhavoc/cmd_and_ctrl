package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overwhelming Remorse — Instant {4}{B} (The Brothers' War):
//
//	"This spell costs {1} less to cast for each creature card in your
//	 graveyard.
//	 Exile target creature or planeswalker."
//
// Cyan, Vengeful Samurai's discount (CR 601.2f), counted as the cost is
// determined, on Devout Decree's exile of a creature or planeswalker.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bb4507c1-1e5d-48b5-8fc0-1d26bd1c9e72",
		Name:         "Overwhelming Remorse",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(creatureCardsInYourGraveyard,
				"This spell costs {1} less to cast for each creature card in your graveyard."),
		},
		Targets:   TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: ResolveSteps(ExileChosenTarget{}),
	})
}
