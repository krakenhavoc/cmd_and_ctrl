package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Debris Field Crusher — Artifact — Spacecraft {4}{R}, 1/5 (Edge of
// Eternities, #4461):
//
//	"When this Spacecraft enters, it deals 3 damage to any target.
//	 Station
//	 8+ | Flying
//	 {1}{R}: This Spacecraft gets +2/+0 until end of turn."
//
// Warmaker Gunship's enters damage on any target, Shivan Dragon's pump
// at {1}{R} for +2/+0. The pump has no tap in its cost and works whether
// or not the Spacecraft is a creature (the boost simply has no effect on
// the printed power until it is one). Station and the 8+ line are the
// shared constructors: at eight or more charge counters it is a 1/5
// artifact creature with flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db0894e1-2644-48d4-8de4-8cc43e940bc1",
		Name:         "Debris Field Crusher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Debris Field Crusher — 3 damage to any target", effects.SourceDealsDamageToEachLegalTarget(3)),
				effects.TargetAny()),
		},
		Activated: []effects.ActivatedAbility{
			effects.Station(),
			{
				Label:   "{1}{R}: This Spacecraft gets +2/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{1}{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.BoostUntilEOT{Target: ctx.Source(), Power: 2, Label: "Debris Field Crusher — +2/+0"}.Apply(ctx)
				},
			},
		},
		Static: append(
			effects.SpacecraftAt(8, 1, 5),
			effects.ThresholdKeywords(8, "flying"),
		),
	})
}
