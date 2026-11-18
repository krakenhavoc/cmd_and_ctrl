package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	Register(Spec{
		OracleID:      "81d62f0a-b2ba-4e87-9003-e5472bd1101f",
		Name:          "Illicit Shipment",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Casualty(3)},
		Purpose:       game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
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
