package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boltwing Marauder — Creature — Dragon {3}{B}{R}, 5/4 (Dragons of Tarkir, #5551):
//
//	"Flying
//	 Whenever another creature you control enters, target creature gets +2/+0 until end of turn."
//
// Mighty Mutanimals' other-creature-entered trigger with a pump on any
// creature.
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:        "1a0162b9-7a7a-4ed9-9a70-189433932449",
		Name:            "Boltwing Marauder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverAnotherCreatureEntersUnderYourControl("Boltwing Marauder — target creature gets +2/+0 until end of turn",
					effects.BoostTargetUntilEOT("Boltwing Marauder — +2/+0", 2, 0)),
				effects.TargetCreature("target creature")),
		},
	})
}
