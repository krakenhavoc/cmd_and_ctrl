package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// grantedTriggerCount counts the GRANTED abilities a permanent lists
// that no activated or mana row accounts for — its granted triggers.
func GrantedTriggerCount(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return
		}
		for _, info := range game.GrantedAbilitiesOf(c) {
			if len(game.CatalogTriggers(info.Key)) > 0 {
				n++
			}
		}
	})
	return n
}
