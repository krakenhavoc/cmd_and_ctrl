package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b851_helpers.go — shared pieces of S82 OLD batch 851 (Ice Age, #5636).

// CanAttackAsThoughHasteUnlessItEnteredThisTurn is "This creature can attack as
// though it had haste unless it entered this turn" (Chaos Lord): a layer-6
// static on the source itself that writes the attack-haste waiver of
// Frenzied Saddlebrute (game.Characteristic.AttackHasteFor) with no holder, so
// every player and planeswalker may be attacked while it is summoning sick
// (CR 302.6). The creature does not gain haste: its {T} abilities stay sick.
// The waiver is withheld in the turn the permanent entered the battlefield
// (Game.EnteredThisTurn, which a change of control does not touch); the layer
// cache is rebuilt as each turn begins.
func CanAttackAsThoughHasteUnlessItEnteredThisTurn() game.StaticAbility {
	return game.StaticAbility{
		Label: "This creature can attack as though it had haste unless it entered this turn.",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return SelfOnly(target, g, source) && !g.EnteredThisTurn(target.InstanceID)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			if !slices.Contains(c.AttackHasteFor, uuid.Nil) {
				c.AttackHasteFor = append(c.AttackHasteFor, uuid.Nil)
			}
		},
	}
}
