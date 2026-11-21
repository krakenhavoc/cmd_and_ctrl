package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushTappedVanillaCreature is pushVanillaCreature with the permanent
// already tapped, so a test can pin the untap clause.
func PushTappedVanillaCreature(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	id := PushVanillaCreature(g, owner, name, power, toughness)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	return id
}
