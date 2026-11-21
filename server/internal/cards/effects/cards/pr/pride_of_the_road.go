package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pride of the Road — Creature — Zombie Cat Warrior {3}{W}, 2/5:
//
//	"Vigilance
//	 Start your engines!
//	 Max speed — At the beginning of combat on your turn, target creature
//	 or Vehicle you control gains double strike until end of turn."
//
// ADR 0138 (#2711). A real target clause (Perilous Snare's "creature or
// Vehicle you control"), so the trigger goes on the stack and can be
// answered; with no legal target it is removed (CR 603.3d). A Vehicle
// that is not a creature gains the keyword and it matters only if it is
// crewed before damage.
//
// No simplification.
func init() {
	const label = "Pride of the Road — target creature or Vehicle you control gains double strike until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "2af236b0-6b30-4ddc-8735-864d2ffad68e",
		Name:            "Pride of the Road",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", effects.StartYourEngines},
		Triggered: []game.TriggeredAbility{
			effects.MaxSpeedTrigger(effects.Targeting(
				effects.AtBeginningOfYourCombat(label, effects.EbTargetCreatureGainsUntilEOT("double strike", label)),
				effects.TargetPermanent("target creature or Vehicle you control", effects.Or(effects.Creature(), effects.OfSubtype("Vehicle")), effects.YouControl()))),
		},
	})
}
