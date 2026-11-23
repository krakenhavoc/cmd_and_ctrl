package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b228_helpers.go — shared shapes of the GAP batch 228 cards
// (#5948), in a file of their own so concurrent batches never collide.

// YouOrAPlayerYoureAttackingHasTheInitiative is "you or a player you're
// attacking has the initiative" (Undercellar Sweep): you hold it, or so does
// the player (the controller of an attacked planeswalker, the protector of an
// attacked battle) any creature you control is attacking.
func YouOrAPlayerYoureAttackingHasTheInitiative(g *game.Game, you uuid.UUID) bool {
	if YouHaveTheInitiative(g, you) {
		return true
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you || c.AttackingTarget == uuid.Nil {
			continue
		}
		if defender := g.DefendingPlayerForAttackForEffect(c.AttackingTarget); defender != uuid.Nil && defender != you && YouHaveTheInitiative(g, defender) {
			return true
		}
	}
	return false
}
