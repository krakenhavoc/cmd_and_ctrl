package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demon's Disciple — Creature — Human Cleric {2}{B}, 3/1 (EDHREC
// rank 2240):
//
//	"When this creature enters, each player sacrifices a creature or
//	 planeswalker of their choice."
//
// Fleshbag Marauder that also eats planeswalkers. The ETB goes on
// the stack as a triggered ability, and its body is
// EachPlayerSacrifices with the controller included — "each player"
// means you too, and the Disciple is on the battlefield when its own
// trigger resolves, so feeding it to itself is the printed line.
// Every player picks their own creature or planeswalker in their own
// prompt; a player with neither is skipped (CR 701.17b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5a33091-a348-4b13-8dbd-79ab0ad99afe",
		Name:         "Demon's Disciple",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Demon's Disciple — each player sacrifices a creature or planeswalker", Do(EachPlayerSacrifices{
				Match: b10CreatureOrPlaneswalker(),
				Label: "a creature or planeswalker",
			})),
		},
	})
}
