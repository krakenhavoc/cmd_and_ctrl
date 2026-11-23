package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pirates_test.go — the Mary Read and Anne Bonny decklist, batch 1:
// artifact-ETB payoffs, discard payoffs, the Pirate lord, and the
// commander's loot-into-Treasure engine.

const (
	RecklessFireweaverOracle   = "180e1a7e-890d-477c-80a5-da8a5f2857b3"
	IngeniousArtilleristOracle = "752c7723-90f8-4e3a-8266-f251ee0dadd8"
	MaraudingMakoOracle        = "e349be42-5f14-44a9-9608-281985c10e2d"
	GlintHornOracle            = "64ad5657-78e9-4f34-8877-18c4f51fff9a"
	CorsairCaptainOracle       = "a7ec13c6-7ade-433a-b5a2-047854eef486"
	ImpulsivePilfererOracle    = "7d9fc9e7-d80b-49c3-871c-ed25b3059ae8"
	AngrathsMaraudersOracle    = "2d4976d4-649c-4d42-ac5a-ada4b46a480c"
	FaithlessLootingOracle     = "3d6fa57a-aa53-4b5c-b8af-a7612c823117"
	MaryReadOracle             = "5182de2d-aceb-450e-bd20-8bc7db124334"
)

// discardTo puts a card with the given type line into a player's
// hand and returns its ID, so a test can discard something specific.
func HandCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func LifeOfOpponents(g *game.Game) []int {
	out := make([]int, 0, len(g.Seats)-1)
	for _, p := range g.Seats[1:] {
		out = append(out, p.Life)
	}
	return out
}
