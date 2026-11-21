package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pheres-Band Brawler — Creature — Centaur Warrior {4}{G}{G}, 4/4 (Theros Beyond Death, #5514):
//
//	"When this creature enters, it fights up to one target creature
//	 you don't control. (Each deals damage equal to its power to the
//	 other.)"
//
// Apex Altisaur's enters fight: "up to one" target, so the trigger
// resolves with no target too, and the Brawler fights only a creature
// still legal as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01499544-6850-49f0-a66e-113df19c4340",
		Name:         "Pheres-Band Brawler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Pheres-Band Brawler — it fights up to one target creature you don't control", effects.ItFightsTheFirstTarget),
				effects.TargetCreature("up to one target creature you don't control", effects.OpponentControls()).WithCount(0, 1)),
		},
	})
}
