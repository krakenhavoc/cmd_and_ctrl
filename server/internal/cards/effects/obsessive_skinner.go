package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Obsessive Skinner — Creature — Human Rogue {1}{G}, 1/1 (Duskmourn
// Commander):
//
//	"When this creature enters, put a +1/+1 counter on target creature.
//	 Delirium — At the beginning of each opponent's upkeep, if there are four
//	 or more card types among cards in your graveyard, put a +1/+1 counter on
//	 target creature."
//
// The enters trigger is Dominion Supervisor's. The upkeep trigger is an
// intervening if (CR 603.4) on the Skinner's controller's own graveyard,
// checked as each opponent's upkeep begins and again as the trigger resolves;
// its target is chosen as it goes on the stack, so with no creature to target
// it is not put on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7e59437-cdd0-476a-b1ec-231be4139d31",
		Name:         "Obsessive Skinner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Obsessive Skinner — put a +1/+1 counter on target creature", PlusOneCounterOnChosenTarget),
				TargetCreature("target creature")),
			Targeting(InterveningIf(b24CondDelirium{}, On(game.EventBeginUpkeep, AllOf(ByAnOpponent, youHaveDeliriumWhen),
				"Obsessive Skinner — delirium: put a +1/+1 counter on target creature", PlusOneCounterOnChosenTarget)),
				TargetCreature("target creature")),
		},
	})
}
