package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bola Slinger — Creature — Cat Soldier {3}{W}, 2/2 (March of the Machine,
// #5477):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target
//	 creature. If that's another creature, it gains the following ability
//	 until end of turn.)
//	 Whenever this creature attacks, tap target artifact or creature an
//	 opponent controls."
//
// Backup (CR 702.165) grants the attack trigger, a bundle in Spec.Grants
// that the Slinger also has itself. "An opponent" is an opponent of the
// controller of the creature that has the ability.
//
// No simplification.
const bolaSlingerGrant = "bola-slinger/attack-tap"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ff995f9-080d-4316-8fbe-39bf195db91b",
		Name:         "Bola Slinger",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       bolaSlingerGrant,
			Triggered: []game.TriggeredAbility{bolaSlingerAttack()},
			Text:      "Whenever this creature attacks, tap target artifact or creature an opponent controls.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Backup("Bola Slinger", 1, bolaSlingerGrant),
			bolaSlingerAttack(),
		},
	})
}

func bolaSlingerAttack() game.TriggeredAbility {
	return effects.Targeting(
		effects.On(game.EventAttack, effects.ThisAttacked,
			"Bola Slinger — tap target artifact or creature an opponent controls",
			effects.B36TapChosenCreature),
		effects.TargetPermanent("target artifact or creature an opponent controls", effects.Or(effects.Artifact(), effects.Creature()), effects.OpponentControls()))
}
