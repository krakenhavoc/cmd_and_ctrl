package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6ff995f9-080d-4316-8fbe-39bf195db91b",
		Name:         "Bola Slinger",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       bolaSlingerGrant,
			Triggered: []game.TriggeredAbility{bolaSlingerAttack()},
			Text:      "Whenever this creature attacks, tap target artifact or creature an opponent controls.",
		}},
		Triggered: []game.TriggeredAbility{
			Backup("Bola Slinger", 1, bolaSlingerGrant),
			bolaSlingerAttack(),
		},
	})
}

func bolaSlingerAttack() game.TriggeredAbility {
	return Targeting(
		On(game.EventAttack, ThisAttacked,
			"Bola Slinger — tap target artifact or creature an opponent controls",
			b36TapChosenCreature),
		TargetPermanent("target artifact or creature an opponent controls", Or(Artifact(), Creature()), OpponentControls()))
}
