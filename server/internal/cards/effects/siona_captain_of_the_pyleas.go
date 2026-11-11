package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Siona, Captain of the Pyleas — Legendary Creature — Human Soldier
// {1}{G}{W}, 2/2 (Wilds of Eldraine Commander, #5368):
//
//	"When Siona enters, look at the top seven cards of your library. You
//	 may reveal an Aura card from among them and put it into your hand.
//	 Put the rest on the bottom of your library in a random order.
//	 Whenever an Aura you control becomes attached to a creature you
//	 control, create a 1/1 white Human Soldier creature token."
//
// The enters ability is Horn of the Mark's sentence
// (LookAtTopThenMayTakeToHand). The attach trigger reads the attach event
// Eriette, the Beguiler reads: the Aura (CardID) is mine as it attaches, the
// host (Target) is a creature I control. Re-attaching to the host it is
// already on is not "becomes attached" (CR 701.3b). An Aura entering
// attached (cast, or put onto the battlefield) is attached by the same
// event, so it counts. The token is a plain one-sentence create, written
// with Do for its footprint.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "89e17ccf-cde7-428d-bd64-47e2f79b40d7",
		Name:         "Siona, Captain of the Pyleas",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Siona, Captain of the Pyleas — look at the top seven cards and may take an Aura card",
				LookAtTopThenMayTakeToHand(7, Subtype("Aura"), 1,
					"Siona, Captain of the Pyleas — reveal an Aura card and put it into your hand")),
			On(game.EventAttach, cmdr2B19YourAuraBecameAttachedToYourCreature,
				"Siona, Captain of the Pyleas — create a 1/1 white Human Soldier creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Human Soldier"), N: 1})),
		},
	})
}
