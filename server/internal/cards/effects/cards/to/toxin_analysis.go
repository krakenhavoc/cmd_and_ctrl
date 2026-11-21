package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Toxin Analysis — Instant {B} (Murders at Karlov Manor, #4469):
//
//	"Target creature gains deathtouch and lifelink until end of turn.
//	 Investigate. (Create a Clue token. It's an artifact with '{2},
//	 Sacrifice this token: Draw a card.')"
//
// Offer Immortality's two keywords on the still-legal target
// (CR 608.2b), then one Clue (CR 701.16a), as Auspicious Arrival makes
// its.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f3b0b15-2c38-4aaa-b527-afbb540a6715",
		Name:         "Toxin Analysis",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch", "lifelink")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.JmpB7FirstLegalTargetGains("Toxin Analysis — deathtouch and lifelink",
					"deathtouch", "lifelink")(ctx.Game, ctx.Item)
			}),
			effects.Investigate{N: 1},
		),
	})
}
