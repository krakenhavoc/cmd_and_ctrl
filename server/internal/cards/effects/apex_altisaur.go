package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Apex Altisaur — Creature — Dinosaur {7}{G}{G}, 10/10 (Lost Caverns of
// Ixalan Commander):
//
//	"When this creature enters, it fights up to one target creature you
//	 don't control.
//	 Enrage — Whenever this creature is dealt damage, it fights up to one
//	 target creature you don't control."
//
// Two triggers sharing one clause (CR 701.14). The enrage one fires once
// per batch of damage, and a fight that has lost either fighter does nothing.
func init() {
	target := func() *game.TargetSpec {
		return TargetCreature("up to one target creature you don't control", OpponentControls()).WithCount(0, 1)
	}
	Register(Spec{
		OracleID:     "fc467fe5-85ab-4c95-8d21-49096778b107",
		Name:         "Apex Altisaur",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Apex Altisaur — it fights up to one target creature you don't control", itFightsTheFirstTarget), target()),
			Targeting(cmdr2B08EnrageTrigger("Apex Altisaur — Enrage: it fights up to one target creature you don't control", itFightsTheFirstTarget), target()),
		},
	})
}
