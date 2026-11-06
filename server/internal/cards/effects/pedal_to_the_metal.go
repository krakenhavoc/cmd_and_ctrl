package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pedal to the Metal — Instant {X}{R} (Aetherdrift, #4464):
//
//	"Target creature gets +X/+0 and gains first strike until end of
//	 turn."
//
// Hatred's +X/+0, X announced with the cast (CR 107.3a), with first
// strike beside it. Both end with the turn. A target gone by resolution
// gets neither (CR 608.2b). With X of 0 the creature still gains first
// strike.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22196948-9f30-4fcd-b43b-c7f81a0f4f36",
		Name:         "Pedal to the Metal",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetCreature("target creature"),
		Purpose: game.Purpose{
			Answers: game.AnswerPump | game.AnswerCombatGrant,
			Targets: game.ForTargets(PowerXUntilEOT(0, "first strike")),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			const label = "Pedal to the Metal — +X/+0 and first strike"
			if err := (BoostUntilEOT{Target: id, Power: ctx.X(), Label: label}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: label}.Apply(ctx)
		},
	})
}
