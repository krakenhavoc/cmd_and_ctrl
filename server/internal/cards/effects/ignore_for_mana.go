package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ignore_for_mana.go — #6247: CR 116.2d's ignore paid with mana. Leonin
// Arbiter: "Players can't search libraries. Any player may pay {2} for
// that player to ignore this effect until end of turn."
//
// Append-only: add a builder, never change what one means.

// AnyPlayerMayPayToIgnore is the "any player may pay <cost> for that
// player to ignore this effect until end of turn" escape of a static
// that bears on every player (Spec.IgnorableEffect). The paying player
// ignores it, and nobody else does.
func AnyPlayerMayPayToIgnore(name, cost string) *game.IgnorableEffect {
	return &game.IgnorableEffect{
		Label:    "Pay " + cost + ": ignore " + name + "'s effect until end of turn",
		Applies:  func(*game.Game, *game.Card, uuid.UUID) bool { return true },
		ManaCost: cost,
	}
}
