package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armguard Familiar — Artifact Creature — Equipment Beast {1}{U}, 2/1:
//
//	"Ward {2}
//	 Equipped creature gets +2/+1 and has ward {2}.
//	 Reconfigure {4}"
//
// Two wards: the Familiar's own (Ward), which protects it whether or not
// it is a creature, and the one it gives its host (WardAttached, Lavaspur
// Boots' shape). Reconfigure is reconfigure.go (#2639).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7955f10-f502-4893-88ad-f36e7ff0af4f",
		Name:         "Armguard Familiar",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Armguard Familiar — ward {2}"),
			effects.WardAttached(effects.WardMana("{2}"), "Armguard Familiar — equipped creature's ward {2}"),
		},
		Activated: effects.Reconfigure("{4}"),
	})
}
