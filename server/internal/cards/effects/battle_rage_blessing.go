package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battle-Rage Blessing — Instant {1}{B} (Dominaria United, #5487):
//
//	"Target creature gains deathtouch and indestructible until end of
//	 turn."
//
// Both keywords are one grant on the one target as the spell resolves,
// through the layered view; a target that has become illegal counters the
// spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "533d6b9f-a145-44c4-9d16-74fcd9c0ae2f",
		Name:         "Battle-Rage Blessing",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "deathtouch", "indestructible")),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return GrantKeywordUntilEOT{
				Target:   t.ID,
				Keywords: []string{"deathtouch", "indestructible"},
				Label:    "Battle-Rage Blessing — deathtouch and indestructible",
			}.Apply(ctx)
		},
	})
}
