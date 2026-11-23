package effectstest

import (
	"github.com/google/uuid"

	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// raffine_a_praetors_test.go — Jin-Gitaxias // The Great Synthesis and
// Sheoldred // The True Scriptures (#3212, slice raffine-a).

func PraetorCard(owner uuid.UUID, oracle, front, frontType string, fp, ft int, back, backType string, backFace bool) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: front, TypeLine: frontType, Power: fp, Toughness: ft},
			{Name: back, TypeLine: backType},
		},
	}
	if backFace {
		c.SetFace(1)
	} else {
		c.SetFace(0)
	}
	return c
}

func JinCard(owner uuid.UUID, back bool) game.Card {
	return PraetorCard(owner, cardsj.JinGitaxiasOracleID, "Jin-Gitaxias", "Legendary Creature — Phyrexian Praetor", 5, 5,
		"The Great Synthesis", "Enchantment — Saga", back)
}

func PraetorFillHand(p *game.Player, n int) []uuid.UUID {
	g := []uuid.UUID{}
	for i := 0; i < n; i++ {
		g = append(g, PushCatalogHandCard(p, "Filler", "Sorcery", ""))
	}
	return g
}

func PraetorOnBattlefield(g *game.Game, oracle string) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].OracleID == oracle {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}
