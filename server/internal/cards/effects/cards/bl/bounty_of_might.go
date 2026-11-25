package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bounty of Might — Instant {4}{G}{G} (Guilds of Ravnica, #5525):
//
//	"Target creature gets +3/+3 until end of turn.
//	 Target creature gets +3/+3 until end of turn.
//	 Target creature gets +3/+3 until end of turn."
//
// Three separate "target creature" clauses (Martial Glory's shape), each its
// own pick: they may be three creatures, or the same one two or three times,
// which then gets +6/+6 or +9/+9. Each pick is re-checked as the spell
// resolves (CR 608.2b) and one that has become illegal gets nothing, the others
// still do.
//
// No simplification.
func init() {
	pump := func(slot int, label string) effects.Instruction {
		return effects.Instruction(func(ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(slot)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 3, Label: "Bounty of Might — " + label + " target +3/+3"}.Apply(ctx)
		})
	}
	effects.Register(effects.Spec{
		OracleID:     "6f067bfd-1c59-463f-99e1-61161b43b405",
		Name:         "Bounty of Might",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3), effects.BenefitUntilEOT(1, 3, 3), effects.BenefitUntilEOT(2, 3, 3)),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature (first +3/+3)"),
			effects.TargetCreature("target creature (second +3/+3)"),
			effects.TargetCreature("target creature (third +3/+3)"),
		),
		OnResolve: effects.ResolveSteps(pump(0, "first"), pump(1, "second"), pump(2, "third")),
	})
}
