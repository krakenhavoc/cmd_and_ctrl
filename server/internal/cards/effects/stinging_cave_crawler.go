package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stinging Cave Crawler — Creature — Insect Horror {2}{B}, 1/3 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Deathtouch
//	 Descend 4 — Whenever this creature attacks, if there are four or more
//	 permanent cards in your graveyard, you draw a card and you lose 1
//	 life."
//
// Deathtouch is the printed keyword (CR 702.2b). Descend 4 is an ability
// word (CR 207.2c); the "if" is an intervening if (CR 603.4), checked as
// the creature attacks and again as the trigger resolves
// (WheneverThisAttacksIfDescend). The life is lost after the draw, and
// waits for any prompt the draw asks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a9a35c77-637f-4d56-afa2-6c8a4ded4838",
		Name:            "Stinging Cave Crawler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacksIfDescend(4, "Stinging Cave Crawler — you draw a card and you lose 1 life",
				Do(DrawCards{N: 1}, youLoseLife(1))),
		},
	})
}
