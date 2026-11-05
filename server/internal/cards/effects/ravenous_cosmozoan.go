package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "5cbe6066-07f8-45de-b3b4-b8590d111b0d",
		Name:            "Ravenous Cosmozoan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(trekB17NoncreatureNonlandCardsInYourGraveyard,
				"This spell costs {1} less to cast for each noncreature, nonland card in your graveyard."),
		},
	})
}
