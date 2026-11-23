package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cgTwoGreenCreatures puts two untapped green 1/1s under `owner`.
func CgTwoGreenCreatures(g *game.Game, owner uuid.UUID) []uuid.UUID {
	return []uuid.UUID{
		PushTapCostPermanent(g, owner, "Elf A", "Creature — Elf", "G"),
		PushTapCostPermanent(g, owner, "Elf B", "Creature — Elf", "G"),
	}
}

// cgBear is a {1}{G} vanilla creature card owned by `owner`.
func CgBear(owner uuid.UUID) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: owner, Controller: owner}
}

func CgTapped(g *game.Game, ids []uuid.UUID) bool {
	for _, id := range ids {
		if c, ok := g.LookupCardForEffect(id); !ok || !c.Tapped {
			return false
		}
	}
	return true
}
