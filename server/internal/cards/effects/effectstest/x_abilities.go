package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// x_abilities_test.go — {X} in an activated ability's cost
// (CR 602.2b). The engine half is in game/activated.go; these are
// the three cards that prove it.

const (
	TreasureVaultOracle   = "3c43efd6-b1a8-452c-ae20-9a936c3340ab"
	SoothsayingOracle     = "517b702a-2c5c-40d7-825e-6c674019b298"
	HelmOfObedienceOracle = "16cadebf-c484-41f8-9e38-5c2c528f5b54"
)

// fillPool drops n colorless mana into a seat's pool.
func FillPool(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
}

func CountTreasures(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == "Treasure" {
			n++
		}
	}
	return n
}

// latestChoiceOfKind returns the newest pending choice of a kind, or
// nil.
func LatestChoiceOfKind(g *game.Game, kind game.PendingChoiceKind) *game.PendingChoice {
	var out *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind {
			out = c
		}
	}
	return out
}

// findTestCard returns the battlefield card with the given ID.
func FindTestCard(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}
