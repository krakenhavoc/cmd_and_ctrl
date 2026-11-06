package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eclipsed Elf — {B/G}{B/G}{B/G} Creature — Elf Scout 3/2 (Lorwyn
// Eclipsed, #4458):
//
//	"When this creature enters, look at the top four cards of your library.
//	 You may reveal an Elf, Swamp, or Forest card from among them and put
//	 it into your hand. Put the rest on the bottom of your library in a
//	 random order."
//
// Eclipsed Boggart's enters trigger with Elf, Swamp and Forest: each an
// effective subtype, so a changeling card is an Elf card and any land with
// the land type counts. Only the controller sees the four; the card taken is
// revealed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c89ef4d-b375-4644-ac8b-034effdbe98c",
		Name:         "Eclipsed Elf",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Eclipsed Elf — look at the top four cards; you may reveal an Elf, Swamp, or Forest card and put it into your hand",
				LookAtTopThenMayTakeToHand(4, AnySubtype("Elf", "Swamp", "Forest"), 1,
					"Eclipsed Elf — reveal an Elf, Swamp, or Forest card and put it into your hand")),
		},
	})
}
