package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:      "81d62f0a-b2ba-4e87-9003-e5472bd1101f",
		Name:          "Illicit Shipment",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Casualty(3)},
		Purpose:       game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:  item.Controller,
				Source:  item.SourceCardID,
				Dest:    game.ZoneHand,
				Limit:   1,
				Shuffle: true,
				Reason:  "Illicit Shipment — search for a card",
			}.Apply(ctx)
		},
	})
}
