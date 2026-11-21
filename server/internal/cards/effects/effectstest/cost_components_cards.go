package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cost_components_cards_test.go — #1213's card proofs: the seven cards
// the three seam rows were waiting on, each exercised through its own
// printed cost.
//
//	return-to-hand cost   Meloku the Clouded Mirror, Master Transmuter,
//	                      Quirion Ranger, Wirewood Symbiote
//	variable sacrifice    Radiant Lotus (one or more), Grim Hireling (X)
//	mana-ability discard  Skirge Familiar
//
// The engine contracts are pinned in game/return_cost_test.go,
// game/variable_sacrifice_cost_test.go and
// game/mana_discard_cost_test.go; what is here is that each card
// declares the component it prints and that the cost actually pays.

func CcLand(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner,
	})
	return id
}

func CcHandCard(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = manaCost
	p.Hand.PushTop(c)
	return c.InstanceID
}
