package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Cosmozoan — Creature — Fish {5}{R} (Star Trek, #4973), 3/3:
//
//	"This spell costs {1} less to cast for each noncreature, nonland card
//	 in your graveyard.
//	 Flying, haste"
//
// The discount (CR 601.2f) counts the caster's own graveyard as the
// cost is determined; it reduces only the generic part, so {R} is
// always paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5cbe6066-07f8-45de-b3b4-b8590d111b0d",
		Name:            "Ravenous Cosmozoan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.TrekB17NoncreatureNonlandCardsInYourGraveyard,
				"This spell costs {1} less to cast for each noncreature, nonland card in your graveyard."),
		},
	})
}
