package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Webspinner Cuff — Artifact Creature — Equipment Spider {2}{G}, 1/4:
//
//	"Reach
//	 Equipped creature gets +1/+4 and has reach.
//	 Reconfigure {4}"
//
// Reconfigure is reconfigure.go (#2639).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7826f3f5-c8f2-43f5-a837-ff016da625e6",
		Name:            "Webspinner Cuff",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 4),
			effects.GrantToAttached("reach"),
		},
		Activated: effects.Reconfigure("{4}"),
	})
}
