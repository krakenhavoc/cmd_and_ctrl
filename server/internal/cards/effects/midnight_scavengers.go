package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Midnight Scavengers — Creature — Human Rogue {4}{B}, 3/3 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"When this creature enters, you may return target creature card with
//	 mana value 3 or less from your graveyard to your hand.
//	 (Melds with Graf Rats.)"
//
// Nimraiser Paladin's enters trigger made optional: the target is chosen
// as the trigger goes on the stack and the "you may" is asked as it
// resolves (CR 603.5). The reminder line is the other half of Graf Rats'
// meld ability (CR 701.42, ADR 0145): Graf Rats' beginning-of-combat
// trigger melds the two into Chittering Host.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c29ae09-cc93-4a2a-b432-3c78ae01c1ed",
		Name:         "Midnight Scavengers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(WhenThisEnters("Midnight Scavengers — you may return target creature card with mana value 3 or less from your graveyard to your hand", returnTargetedCardToHand),
				TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard", Creature(), ManaValueLE(3), YouOwn())),
				"Midnight Scavengers — return the creature card to your hand?"),
		},
	})
}
