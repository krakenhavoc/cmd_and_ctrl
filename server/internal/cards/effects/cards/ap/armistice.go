package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armistice — Enchantment {2}{W} (Commander 2014):
//
//	"{3}{W}{W}: You draw a card and target opponent gains 3 life."
//
// The target opponent is chosen on activation. If they are gone by
// resolution (CR 608.2b) the ability does nothing, including the draw: it
// has one target and it is no longer legal.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da103316-2a85-4de9-8531-ac2cd2859d6f",
		Name:         "Armistice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}{W}: You draw a card and target opponent gains 3 life.",
			Purpose: game.Purpose{Draws: 1},
			Cost:    effects.ManaCost("{3}{W}{W}"),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				opp, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				if err := (effects.DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.GainLife{Player: opp, Amount: 3}.Apply(ctx)
			},
		}},
	})
}
