package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Timbermare — Creature — Elemental Horse, {3}{G}, 5/5:
//
//	"Haste
//	 Echo {5}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, tap all other creatures."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8e25e4f5-5676-480f-a52b-43ebb6f31537",
		Name:            "Timbermare",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.Echo("Timbermare", "{5}{G}"),
			effects.WhenThisEnters("Timbermare — tap all other creatures", effects.TapAllOtherCreatures),
		},
	})
}
