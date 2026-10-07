package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Caustic Exhale — Instant {B}:
//
//	"As an additional cost to cast this spell, behold a Dragon or pay {1}. (To behold a Dragon, choose a Dragon you control or reveal a Dragon card from your hand.)
//	 Target creature gets -3/-3 until end of turn."
//
// The additional cost is the either/or branch cost of ADR 0100 §2
// with the behold branch added by its 2026-10-07 amendment
// (BeholdOrPay): the caster announces the branch, names the
// card on reveal_ids, and either shows it to the table or pays {1}
// more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "c4023e1f-26e3-4225-98d0-3d33f6654616",
		Name:           "Caustic Exhale",
		Completeness:   CompletenessFull,
		AdditionalCost: BeholdOrPay("a", "Dragon", "{1}"),
		Targets:        TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -3,
				Toughness: -3,
				Label:     "Caustic Exhale — -3/-3 until end of turn",
			}.Apply(ctx)
		},
	})
}
