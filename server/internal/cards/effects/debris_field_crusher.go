package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "db0894e1-2644-48d4-8de4-8cc43e940bc1",
		Name:         "Debris Field Crusher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Debris Field Crusher — 3 damage to any target", debrisFieldCrusherDamage),
				TargetAny()),
		},
		Activated: []ActivatedAbility{
			Station(),
			{
				Label:   "{1}{R}: This Spacecraft gets +2/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{1}{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return BoostUntilEOT{Target: ctx.Source(), Power: 2, Label: "Debris Field Crusher — +2/+0"}.Apply(ctx)
				},
			},
		},
		Static: append(
			SpacecraftAt(8, 1, 5),
			ThresholdKeywords(8, "flying"),
		),
	})
}

// debrisFieldCrusherDamage deals 3 damage from the Spacecraft to the
// announced target if it is still legal.
func debrisFieldCrusherDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 3}.Apply(ctx)
	}
	return nil
}
