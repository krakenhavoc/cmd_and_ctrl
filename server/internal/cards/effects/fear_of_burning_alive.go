package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fear of Burning Alive — Enchantment Creature — Nightmare {4}{R}{R}, 4/4
// (Duskmourn: House of Horror, #4466):
//
//	"When this creature enters, it deals 4 damage to each opponent.
//	 Delirium — Whenever a source you control deals noncombat damage to
//	 an opponent, if there are four or more card types among cards in
//	 your graveyard, this creature deals that amount of damage to target
//	 creature that player controls."
//
// The enters trigger is the shared "deals N damage to each opponent"
// primitive. The delirium trigger is Niv-Mizzet, Visionary's watch (one
// trigger per damage event, so noncombat damage dealt to two opponents at
// once is two triggers, as the ruling says) behind an intervening "if"
// (CR 603.4): the four card types are checked as it triggers and again as
// it resolves (b24CondDelirium). "That player" is the opponent the damage
// event names, so the target is chosen from creatures that player controls
// (TargetsFrom, read off the triggering event), and "that amount" is the
// event's damage, read as the ability resolves. The damage is Fear's, not
// the original source's.
//
// No simplification.
func init() {
	const label = "Fear of Burning Alive — it deals that amount of damage to target creature that player controls"
	delirium := InterveningIf(b24CondDelirium{}, game.TriggeredAbility{
		Watches:     []game.EventKind{game.EventDealDamage},
		AppliesTo:   AllOf(ASourceYouControlDealtNoncombatDamageToAnOpponent, youHaveDeliriumWhen),
		TargetsFrom: b24CreatureThatPlayerControls,
		Key:         label,
		Effect:      b24DamageToTargetEqualToEvent,
	})
	Register(Spec{
		OracleID:     "0ae74149-60fc-407a-8782-8865458606b5",
		Name:         "Fear of Burning Alive",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fear of Burning Alive — it deals 4 damage to each opponent", Do(DamageEachOpponent{N: 4})),
			delirium,
		},
	})
}
