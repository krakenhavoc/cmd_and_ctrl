package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "db849660-1ad4-4af4-8822-40ba6c0b6d73",
		Name:         "Defiant Ogre",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{defiantOgreEnters()},
	})
}

// defiantOgreEnters is the modal entry trigger: a +1/+1 counter on the
// Ogre, or destroy target artifact.
func defiantOgreEnters() game.TriggeredAbility {
	enters := effects.WhenThisEnters("Defiant Ogre — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeDoing("Put a +1/+1 counter on this creature.", nil, effects.PutAPlusOneCounterOnThisCreatureMode),
		effects.ModeDoing("Destroy target artifact.", effects.TargetPermanent("target artifact", effects.Artifact()), effects.TitanOfIndustryDestroy),
	)
	return enters
}
