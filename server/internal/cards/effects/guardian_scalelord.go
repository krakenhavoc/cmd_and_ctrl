package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guardian Scalelord — Creature — Dragon {4}{W}, 3/4 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target
//	 creature. If that's another creature, it gains the following
//	 abilities until end of turn.)
//	 Flying
//	 Whenever this creature attacks, return target nonland permanent card
//	 with mana value X or less from your graveyard to the battlefield,
//	 where X is this creature's power."
//
// Backup (CR 702.165) grants the two abilities printed below it: flying,
// and the attack trigger, a bundle in Spec.Grants that the Scalelord also
// has itself. The trigger is Winter Soldier's: the mana-value bound is
// data on the clause (RelativeToSource, ManaValueNoGreaterThanPower),
// judged against the power of the creature whose ability it is as the
// target is chosen and again as it resolves (CR 608.2b), and its
// last-known power once it has left (CR 608.2h). "Your graveyard" is the
// graveyard of that creature's controller, so a creature that gained the
// ability returns from its own controller's graveyard.
//
// No simplification.
const guardianScalelordGrant = "guardian-scalelord/attack-return"

func init() {
	Register(Spec{
		OracleID:        "d6c7d7a6-2d38-4c06-ab83-bee9cc18de52",
		Name:            "Guardian Scalelord",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []AbilityGrant{{
			Key:       guardianScalelordGrant,
			Triggered: []game.TriggeredAbility{guardianScalelordAttack()},
			Text:      "Whenever this creature attacks, return target nonland permanent card with mana value X or less from your graveyard to the battlefield, where X is this creature's power.",
		}},
		Triggered: []game.TriggeredAbility{
			Backup("Guardian Scalelord", 1, guardianScalelordGrant, "flying"),
			guardianScalelordAttack(),
		},
	})
}

// guardianScalelordAttack is the attack trigger, built fresh for the card
// and for the bundle backup grants.
func guardianScalelordAttack() game.TriggeredAbility {
	return Targeting(
		On(game.EventAttack, ThisAttacked,
			"Guardian Scalelord — return target nonland permanent card with mana value X or less from your graveyard to the battlefield, where X is this creature's power",
			returnFirstLegalGraveyardTargetToBattlefield),
		RelativeToSource(
			TargetCardInGraveyard("target nonland permanent card with mana value X or less from your graveyard, where X is this creature's power",
				YouOwn(), NonlandPermanentCard()),
			ManaValueNoGreaterThanPower()))
}
