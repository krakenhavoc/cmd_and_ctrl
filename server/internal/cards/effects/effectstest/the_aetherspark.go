package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const TheAethersparkOracle = "483c747a-b602-4747-95bc-72c2c17ed509"

// pushAetherspark seats The Aetherspark with `loyalty` loyalty
// counters. The type line is the whole point of the card and it is
// spelled out here rather than reusing pushCatalogWalker: the
// "Planeswalker" is what CR 606 reads for the loyalty abilities, and
// the "Equipment" subtype is what CR 704.5n reads to decide the
// attachment is legal.
func PushAetherspark(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "The Aetherspark",
		TypeLine:   "Legendary Artifact Planeswalker — Equipment",
		OracleID:   TheAethersparkOracle,
		Owner:      owner,
		Controller: owner,
		Counters:   map[string]int{game.CounterLoyalty: loyalty},
	})
	return id
}
