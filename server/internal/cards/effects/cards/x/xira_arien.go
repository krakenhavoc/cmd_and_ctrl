package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Xira Arien — Legendary Creature — Insect Wizard {B}{R}{G}, 1/2
// (Legends, #5639):
//
//	"Flying
//	 {B}{R}{G}, {T}: Target player draws a card."
//
// Kenrith, the Returned King's "Target player draws a card." row behind
// a tap: any player, the controller included, and nothing if the player
// is no longer a legal target as it resolves (CR 608.2b). The tap means
// summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4a6e367c-7bc9-44a3-8ede-f2d0651abad3",
		Name:            "Xira Arien",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{R}{G}, {T}: Target player draws a card.",
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1}),
			Cost:    effects.Plus(effects.ManaCost("{B}{R}{G}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if p, ok := effects.FirstLegalPlayerTarget(ctx); ok {
					return effects.DrawCards{Player: p, N: 1}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
