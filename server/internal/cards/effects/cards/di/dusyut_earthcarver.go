package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dusyut Earthcarver — Creature — Elephant Druid {5}{G}, 4/4 (Tarkir:
// Dragonstorm, #4463):
//
//	"Reach
//	 When this creature enters, it endures 3. (Put three +1/+1 counters
//	 on it or create a 3/3 white Spirit creature token.)"
//
// Endure is endure_n.go (CR 701.63a), Sandskitter Outrider's entry
// trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "387b12b1-d8f5-487b-972f-5303b23eed5a",
		Name:            "Dusyut Earthcarver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dusyut Earthcarver — it endures 3", effects.ThisEndures("Dusyut Earthcarver", 3)),
		},
	})
}
