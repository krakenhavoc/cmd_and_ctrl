package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// draw_step.go — the shared shape of "you draw a card during each
// opponent's draw step" (Teferi, Who Slows the Sunset's emblem,
// #1315). See game/draw_step.go for why this is a permission on a
// turn-based action rather than a triggered ability, and Spec.DrawStep
// / EmblemSpec.DrawStep for what the difference buys.

// eachOtherPlayersDrawStep is the AppliesTo shared by the whole
// family — untap_step.go's eachOtherPlayersUntapStep, one turn-based
// action over. "Each OTHER player" is how the clause is printed, and
// it is not skippable: the active player already draws their own card
// by CR 504.1, so a permission that also fired there would double it.
func eachOtherPlayersDrawStep(_ *game.Game, source *game.Card, activePlayer uuid.UUID) bool {
	return activePlayer != uuid.Nil && activePlayer != source.Controller
}

// drawDuringEachOtherPlayersDrawStep builds the permission for "you
// draw a card during each other player's draw step" — the source's
// controller draws N cards (default 1) during any other player's draw
// step.
func drawDuringEachOtherPlayersDrawStep(label string, n int) game.DrawStepPermission {
	return game.DrawStepPermission{
		Label:     label,
		AppliesTo: eachOtherPlayersDrawStep,
		Drawer: func(_ *game.Game, source *game.Card) uuid.UUID {
			return source.Controller
		},
		N: n,
	}
}
