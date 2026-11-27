package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Awakening — Enchantment {2}{G}{G} (Stronghold, S82 OLD b750, #5627):
//
//	"At the beginning of each upkeep, untap all creatures and lands."
//
// Baleful Force's every-upkeep trigger (AtEachUpkeep, anyone's), with a body
// that untaps every tapped creature and land on the battlefield, whoever
// controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5563427d-ffdc-4bea-ad7d-1605894b8a6c",
		Name:         "Awakening",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Awakening — untap all creatures and lands", effects.OldB750UntapAllCreaturesAndLands),
		},
	})
}
