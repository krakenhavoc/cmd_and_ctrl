package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nimraiser Paladin — Creature — Phyrexian Knight {4}{B}, 4/4:
//
//	"Toxic 2
//	 When this creature enters, return target creature card with mana
//	 value 3 or less from your graveyard to your hand."
//
// The target is chosen as the trigger goes on the stack and re-checked as it
// resolves (CR 608.2b). With no such card, the trigger is removed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c9daff24-1af8-4133-bff1-e0e4d1042db4",
		Name:            "Nimraiser Paladin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"toxic 2"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Nimraiser Paladin — return target creature card with mana value 3 or less from your graveyard to your hand", returnTargetedCardToHand),
				TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard", Creature(), ManaValueLE(3), YouOwn())),
		},
	})
}
