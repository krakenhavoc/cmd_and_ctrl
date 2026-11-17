package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Writhing Necromass — Creature — Zombie Giant {6}{B}, 5/5 (Dominaria
// United, S82 DMU batch 4):
//
//	"This spell costs {1} less to cast for each creature card in your
//	 graveyard.
//	 Deathtouch"
//
// A self cost modifier counting the creature cards in the caster's
// graveyard (CountCardsInYourGraveyard), reducing only the generic part;
// deathtouch is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7f6505a1-ab8b-468b-a5de-33451598b464",
		Name:            "Writhing Necromass",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(CountCardsInYourGraveyard(Creature()),
				"This spell costs {1} less to cast for each creature card in your graveyard."),
		},
	})
}
