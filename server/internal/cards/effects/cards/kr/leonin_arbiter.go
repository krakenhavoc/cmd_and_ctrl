package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leonin Arbiter — Creature — Cat Cleric {1}{W}, 2/2 (Commander 2017, #5533):
//
//	"Players can't search libraries. Any player may pay {2} for that player
//	 to ignore this effect until end of turn."
//
// Stranglehold's search lock over every player, its controller included
// (Spec.PlayersCantSearchLibraries, CR 701.23a), with #6247's special action:
// any player may pay {2} for themselves, and only themselves, to ignore it
// until end of turn (CR 116.2d). The action is a special action, so it is
// taken any time that player has priority and does not use the stack; it is
// offered only to a player who can pay {2}. Judged through the card's own
// abilities, so an Arbiter that has lost its abilities blocks nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "a33f6070-590c-4be8-9180-b1ac2608cbc4",
		Name:                       "Leonin Arbiter",
		Completeness:               effects.CompletenessFull,
		PlayersCantSearchLibraries: true,
		IgnorableEffect: &game.IgnorableEffect{
			Label:    "Pay {2}: ignore Leonin Arbiter's effect until end of turn",
			ManaCost: "{2}",
			Applies:  func(*game.Game, *game.Card, uuid.UUID) bool { return true },
		},
	})
}
