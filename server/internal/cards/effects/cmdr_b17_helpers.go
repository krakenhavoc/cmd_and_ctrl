package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b17_helpers.go — shared pieces of Commander deck cards, batch 17
// (#5368).

// cmdrB17DefendersControlled is the number of creatures with defender
// `controller` controls right now, read through the layers (Blight Pile's
// X, Wingmantle Chaplain's count).
func cmdrB17DefendersControlled(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && game.HasKeyword(&c, "defender") {
			n++
		}
	}
	return n
}

// cmdrB17InstantSorceryCardsInGraveyard is the number of instant and
// sorcery cards in `player`'s graveyard (Haughty Djinn). A split card
// counts once. It walks the seats directly because it runs inside the
// layer pass, where the lock is already held.
func cmdrB17InstantSorceryCardsInGraveyard(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.ID != player || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.IsInstant() || c.IsSorcery() {
				n++
			}
		}
	}
	return n
}
