package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Borrowed Hostility — Instant {R}:
//
//	"Escalate {3} (Pay this cost for each mode chosen beyond the
//	 first.)
//	 Choose one or both —
//	 • Target creature gets +3/+0 until end of turn.
//	 • Target creature gains first strike until end of turn."
//
// Escalate (CR 702.120a, #2126): {3} for the second mode.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b45fc33b-0658-4fe1-89da-54a07a12ebd4",
		Name:         "Borrowed Hostility",
		Completeness: CompletenessFull,
		Modes: Escalating(ChooseN("Choose one or both", 1, 2,
			ModeDoing("Target creature gets +3/+0 until end of turn.",
				TargetCreature("target creature"),
				BoostTheModesTarget(3, 0, "Borrowed Hostility")),
			ModeDoing("Target creature gains first strike until end of turn.",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike"}, Label: "Borrowed Hostility"}.Apply(ctx)
				}),
		), EscalateMana("{3}")),
	})
}
