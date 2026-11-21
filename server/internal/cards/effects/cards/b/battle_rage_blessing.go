package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "533d6b9f-a145-44c4-9d16-74fcd9c0ae2f",
		Name:         "Battle-Rage Blessing",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch", "indestructible")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.GrantKeywordUntilEOT{
				Target:   t.ID,
				Keywords: []string{"deathtouch", "indestructible"},
				Label:    "Battle-Rage Blessing — deathtouch and indestructible",
			}.Apply(ctx)
		},
	})
}
