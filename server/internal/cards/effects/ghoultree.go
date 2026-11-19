package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghoultree — Creature — Zombie Treefolk {7}{G}, 10/10 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"This spell costs {1} less to cast for each creature card in your
//	 graveyard."
//
// Cyan, Vengeful Samurai's self cost modifier: it counts the creature
// cards in the caster's graveyard as the spell is cast and applies only
// while Ghoultree is a spell; the generic part of the cost can't go
// below zero.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f2628e3-e29a-4d77-bc27-6e1ce240d192",
		Name:         "Ghoultree",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(creatureCardsInYourGraveyard,
				"This spell costs {1} less to cast for each creature card in your graveyard."),
		},
	})
}
