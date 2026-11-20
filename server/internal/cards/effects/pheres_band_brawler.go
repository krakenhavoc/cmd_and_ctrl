package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "01499544-6850-49f0-a66e-113df19c4340",
		Name:         "Pheres-Band Brawler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Pheres-Band Brawler — it fights up to one target creature you don't control", itFightsTheFirstTarget),
				TargetCreature("up to one target creature you don't control", OpponentControls()).WithCount(0, 1)),
		},
	})
}
