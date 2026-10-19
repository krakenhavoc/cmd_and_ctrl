package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inspiring Overseer — Creature — Angel Cleric {2}{W}, 2/1 (EDHREC
// rank 1658):
//
//	"Flying
//	 When this creature enters, you gain 1 life and draw a card."
//
// White's three-mana cantrip flyer. One entry trigger doing both
// halves in printed order — the life first, so a "whenever you gain
// life" watcher and a "whenever you draw" watcher queue in the
// order the card reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d646e42b-5635-4798-b633-29c093b66a55",
		Name:            "Inspiring Overseer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Inspiring Overseer — gain 1 life and draw a card", Do(
				// The draw waits for a prompt the life gain asks (CR 608.2c, #4155).
				GainLife{Amount: 1},
				DrawCards{N: 1},
			)),
		},
	})
}
