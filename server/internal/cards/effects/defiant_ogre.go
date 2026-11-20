package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Defiant Ogre — Creature — Ogre Warrior {5}{R}, 3/5 (Fate Reforged,
// #5552):
//
//	"When this creature enters, choose one —
//	 • Put a +1/+1 counter on this creature.
//	 • Destroy target artifact."
//
// Knight of Autumn's modal entry trigger (CR 700.2): the mode is chosen as
// the trigger goes on the stack, and the artifact bullet is offered only
// when there is an artifact to target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "db849660-1ad4-4af4-8822-40ba6c0b6d73",
		Name:         "Defiant Ogre",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{defiantOgreEnters()},
	})
}

// defiantOgreEnters is the modal entry trigger: a +1/+1 counter on the
// Ogre, or destroy target artifact.
func defiantOgreEnters() game.TriggeredAbility {
	enters := WhenThisEnters("Defiant Ogre — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = ChooseOne(
		ModeDoing("Put a +1/+1 counter on this creature.", nil, putAPlusOneCounterOnThisCreatureMode),
		ModeDoing("Destroy target artifact.", TargetPermanent("target artifact", Artifact()), titanOfIndustryDestroy),
	)
	return enters
}
