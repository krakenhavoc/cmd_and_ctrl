package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boughside Wanderers — Creature — Elf Scout {4}{G}{G} (The Hobbit), 4/4:
//
//	"When this creature enters, look at the top four cards of your
//	 library. You may reveal a permanent card from among them and put it
//	 into your hand. Put the rest on the bottom of your library in a
//	 random order.
//	 Landfall — Whenever a land you control enters, this creature gets
//	 +2/+2 until end of turn."
//
// #4454, S76. The enters ability is Horn of the Mark's sentence
// (LookAtTopThenMayTakeToHand) with a permanent-card filter: a land is a
// permanent card, an instant or sorcery is not. The landfall pump is
// Kazandu Mammoth's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc788b66-d03d-4066-be35-481be96ed2f7",
		Name:         "Boughside Wanderers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Boughside Wanderers — look at the top four cards and may take a permanent card",
				LookAtTopThenMayTakeToHand(4, Permanent(), 1,
					"Boughside Wanderers — reveal a permanent card and put it into your hand")),
			Landfall("Boughside Wanderers — +2/+2 until end of turn (landfall)",
				thisGetsUntilEndOfTurn(2, 2, "Boughside Wanderers — +2/+2 until end of turn")),
		},
	})
}
