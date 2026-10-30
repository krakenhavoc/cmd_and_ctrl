package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bucky Barnes, Eager Ally — Legendary Creature — Human Soldier Hero {1}{W},
// 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Vigilance
//	 When Bucky Barnes dies, look at the top four cards of your library. You
//	 may reveal an Equipment, Hero, or Soldier card from among them and put it
//	 into your hand. Put the rest on the bottom of your library in a random
//	 order."
//
// Narset, Parter of Veils' look-and-take body (LookAtTopThenMayTakeToHand)
// with a three-subtype filter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "55bc5d06-7dd1-4403-af32-0e10d8bbe327",
		Name:            "Bucky Barnes, Eager Ally",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Bucky Barnes — look at the top four cards; you may reveal an Equipment, Hero, or Soldier card and put it into your hand",
				LookAtTopThenMayTakeToHand(4, AnySubtype("Equipment", "Hero", "Soldier"), 1,
					"Bucky Barnes — reveal an Equipment, Hero, or Soldier card and put it into your hand")),
		},
	})
}
