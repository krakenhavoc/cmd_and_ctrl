package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b14_helpers.go — shared pieces of Phyrexia: All Will Be One batch
// 14, the Equipment-matters cards (#5479, S82).

// s82OneB14Equipped passes for a creature that has an Equipment attached,
// read as the effect resolves ("equipped creatures you control").
func S82OneB14Equipped() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return IsEquippedCreature(g, c)
	}
}

// s82OneB14LosesIndestructible is "that permanent loses indestructible until
// end of turn" as a damage continuation, whatever the damage did (Rebel
// Salvo prints no "dealt damage this way"). An on-disk key.
var S82OneB14LosesIndestructible = RecipientTail("recipient/s82-one-b14-loses-indestructible",
	func(ctx *Context, target uuid.UUID, _ int, _ game.TailOutcome) error {
		return G2LosesIndestructibleUntilEOT(ctx, target)
	})
