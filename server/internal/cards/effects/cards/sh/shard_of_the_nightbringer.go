package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shard of the Nightbringer — Creature — C'tan {5}{B}{B}{B}, 8/8 (Warhammer
// 40,000 Commander, #5948):
//
//	"Flying
//	 Drain Life — When this creature enters, if you cast it, target opponent
//	 loses half their life, rounded up. You gain life equal to the life lost
//	 this way."
//
// Geological Appraiser's "if you cast it" (read off the entering permanent's
// provenance, CR 603.4), a targeted opponent, and the half taken as the ability
// resolves. The gain is the life actually lost, so a prevented loss gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eedd5a17-d0fe-4a13-8f2c-f343bfa43f92",
		Name:            "Shard of the Nightbringer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.AllOf(effects.Self, effects.GeologicalAppraiserWasCast), "Shard of the Nightbringer — target opponent loses half their life, rounded up; you gain that much life",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GapB189LoseHalfLifeAndGainIt(effects.NewContext(g, item))
				}), effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
