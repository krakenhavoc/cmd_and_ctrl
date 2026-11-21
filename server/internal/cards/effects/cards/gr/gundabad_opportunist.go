package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gundabad Opportunist — Creature — Goblin Rogue {3}{R}, 4/2 (The Hobbit,
// HOB):
//
//	"When this creature enters, exile the top card of your library. Until
//	 the end of your next turn, you may play that card."
//
// Gila Courser's body (b20ExileTopUntilEndOfNextTurn): "play", so a land is
// not stranded, and the duration is ADR 0063's seat-turn "end of your next
// turn".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61773534-8da8-48fa-b51a-73c0a9a751c4",
		Name:         "Gundabad Opportunist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Gundabad Opportunist — exile the top card of your library; you may play it until the end of your next turn",
				effects.B20ExileTopUntilEndOfNextTurn),
		},
	})
}
