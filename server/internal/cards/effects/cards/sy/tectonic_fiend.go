package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tectonic Fiend — Creature — Elemental, {4}{R}{R}, 7/7:
//
//	"Echo {4}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 This creature attacks each combat if able."
//
// "Attacks each combat if able" is AttacksEachCombat() (CR 508.1d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ead1bcc4-6bf7-4ace-b705-e3095e6a716c",
		Name:         "Tectonic Fiend",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			effects.Echo("Tectonic Fiend", "{4}{R}{R}"),
		},
	})
}
