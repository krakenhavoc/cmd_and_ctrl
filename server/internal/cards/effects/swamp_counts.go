package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// swamp_counts.go — "the number of Swamps you control", the count
// behind Defile and Consuming Corruption. It reads the EFFECTIVE type
// line (MatchLandSubtype), so Urborg, Tomb of Yawgmoth turns every
// land into a Swamp for the count. Cabal Coffers counts the same way
// through ProducedPerPermanent. Append-only.

// swampsControlledBy is the number of Swamps `player` controls as the
// effect resolves. `swamp` is the land type as the counting object's
// text now reads it (ctx.TextSubtype, CR 612.2).
func swampsControlledBy(g *game.Game, player uuid.UUID, swamp string) int {
	return countControlled(g, player, MatchLandSubtype(swamp))
}
