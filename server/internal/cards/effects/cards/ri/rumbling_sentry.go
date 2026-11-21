package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rumbling Sentry — Creature — Giant {3}{W}{W}, 3/6 (Theros Beyond Death, #5514):
//
//	"When this creature enters, scry 1."
//
// Chrome Cat's enters scry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7126772a-d586-4859-81ec-46bbe9d1a678",
		Name:         "Rumbling Sentry",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rumbling Sentry — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
