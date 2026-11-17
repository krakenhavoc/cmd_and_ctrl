package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kolaghan Warmonger — Creature — Ogre Warrior {2}{R}, 3/2 (March of the
// Machine: The Aftermath):
//
//	"Haste
//	 Whenever this creature attacks, look at the top six cards of your
//	 library. You may reveal a Dragon card from among them and put it
//	 into your hand. Put the rest on the bottom of your library in a
//	 random order."
//
// Ashe, Princess of Dalmasca's sentence (LookAtTopThenMayTakeToHand) with
// six cards and a Dragon filter: only the controller sees the six, the
// Dragon taken is revealed, and the rest go to the bottom in an order
// drawn from the game's seeded RNG. "Dragon" is a word of the Warmonger's
// text (CR 612.2), read through AnySubtype like Dragonologist's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4c7e6520-d179-4590-afdc-92eb2c0740a6",
		Name:            "Kolaghan Warmonger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Kolaghan Warmonger — look at the top six cards of your library",
				LookAtTopThenMayTakeToHand(6, AnySubtype("Dragon"), 1,
					"Kolaghan Warmonger — reveal a Dragon card and put it into your hand")),
		},
	})
}
