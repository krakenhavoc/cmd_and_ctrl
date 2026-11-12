package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wojek Siren — Instant {W}:
//
//	"Radiance — Target creature and each other creature that shares a
//	 color with it get +1/+1 until end of turn."
//
// The creatures are read as the spell resolves (radianceSet, #5736),
// whoever controls them, and get +1/+1 as one effect locked now
// (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "164b21ce-834b-4ec2-9094-33e51a46f9ef",
		Name:         "Wojek Siren",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(BenefitUntilEOT(0, 1, 1))},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ids := radianceSet(ctx, Creature())
			if len(ids) == 0 {
				return nil
			}
			return BoostUntilEOT{Match: b10OneOf(ids), Power: 1, Toughness: 1, Label: "Wojek Siren — +1/+1"}.Apply(ctx)
		},
	})
}
