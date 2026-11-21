package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushGideon seats a Gideon planeswalker (subtype Gideon, as printed)
// with `loyalty` counters. No battlefield-entry event is emitted, so it
// is not summoning sick.
func PushGideon(g *game.Game, owner uuid.UUID, name, oracle string, loyalty int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Legendary Planeswalker — Gideon",
		OracleID:   oracle,
		Owner:      owner,
		Controller: owner,
		Counters:   map[string]int{game.CounterLoyalty: loyalty},
	})
	return id
}
