package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necron Overlord — Artifact Creature — Necron Noble {2}{B}{B}, 2/5
// (Warhammer 40,000 Commander, S82 GAP2 batch 04, #5948):
//
//	"Relentless March — {X}, {T}, Tap X untapped artifacts you control:
//	 Target opponent loses X life."
//
// One announced X pays the {X} and counts the artifacts tapped (CR
// 107.3a, #6085): Apothecary White's TapXUntapped beside ManaCost("{X}").
// The {T} is the Overlord's own, so it is summoning-sick the turn it
// arrives (CR 302.6) and, already tapped, is never one of the X
// artifacts. The opponent loses the announced X.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c0de95f-4269-439c-a1bc-faa11348b185",
		Name:         "Necron Overlord",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label: "Relentless March — {X}, {T}, Tap X untapped artifacts you control: Target opponent loses X life.",
			Cost: effects.Plus(effects.ManaCost("{X}"), effects.TapCost(),
				effects.TapXUntapped("X untapped artifacts you control", effects.Artifact())),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if ctx.X() <= 0 {
					return nil
				}
				return effects.TargetPlayerLosesLife{N: ctx.X()}.Apply(ctx)
			},
		}},
	})
}
