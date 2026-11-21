package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const SorinOfHouseMarkovOracle = "84e84532-498b-40ea-9510-46fb2403939b"

// sorinCard is the double-faced card as the deck importer builds it,
// front face up.
func SorinCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   SorinOfHouseMarkovOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Sorin of House Markov", TypeLine: "Legendary Creature — Human Noble", ManaCost: "{1}{B}",
				Colors: []string{"B"}, Power: 1, Toughness: 4},
			{Name: "Sorin, Ravenous Neonate", TypeLine: "Legendary Planeswalker — Sorin",
				Colors: []string{"B"}, StartingLoyalty: 3},
		},
	}
	c.SetFace(0)
	return c
}
