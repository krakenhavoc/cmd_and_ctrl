package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iceridge Serpent — Creature — Serpent {4}{U}, 3/3 (Tarkir: Dragonstorm,
// #4463):
//
//	"When this creature enters, return target creature an opponent
//	 controls to its owner's hand."
//
// Exclusion Mage's entry trigger. With no creature an opponent controls,
// the trigger has no target and is removed from the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53f07529-c212-43e4-b5b6-086efeadda39",
		Name:         "Iceridge Serpent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Iceridge Serpent — return target creature an opponent controls to its owner's hand",
				bounceTheTargetToHand), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
