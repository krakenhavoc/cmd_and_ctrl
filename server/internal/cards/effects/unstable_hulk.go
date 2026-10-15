package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unstable Hulk — Creature — Goblin Mutant {1}{R}{R}, 2/2:
//
//	"Morph {3}{R}{R}
//	 When this creature is turned face up, it gets +6/+6 and gains trample
//	 until end of turn. You skip your next turn."
//
// The morph half is the keyword's (Morph). The trigger is one ordinary
// ability on the stack: the boost and the trample are scoped effects
// pinned to this object (CR 400.7), the skip is a separate instruction
// that is written whether or not the Hulk is still there (CR 614.10, ADR
// 0188 Decision 1, #3565).
//
// No simplification.
func init() {
	const label = "Unstable Hulk — it gets +6/+6 and gains trample until end of turn, and you skip your next turn"
	Register(Spec{
		OracleID:         "8d01c85c-5f83-4514-b81c-4165051d5722",
		Name:             "Unstable Hulk",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{3}{R}{R}")},
		// The skip happens only when it is turned face up, so it rides
		// Purpose.FaceUp (ADR 0126's amendment of 2026-10-15, #3762): the
		// heuristic charges the turn when it prices the turn_face_up
		// special action. The +6/+6 and trample are not priced; no purpose
		// field prices a one-turn pump on a permanent being flipped.
		Purpose: game.Purpose{FaceUp: &game.Purpose{SkipsTurns: 1}},
		Triggered: []game.TriggeredAbility{
			WhenThisIsTurnedFaceUp(label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (BoostUntilEOT{Target: ctx.Source(), Power: 6, Toughness: 6, Label: "Unstable Hulk — +6/+6"}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"trample"}, Label: "Unstable Hulk — trample"}).Apply(ctx); err != nil {
					return err
				}
				return SkipYourNextTurns{}.Apply(ctx)
			}),
		},
	})
}
