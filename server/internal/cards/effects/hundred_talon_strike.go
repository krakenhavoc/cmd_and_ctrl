package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hundred-Talon Strike — Instant — Arcane {W}:
//
//	"Target creature gets +1/+0 and gains first strike until end of turn.
//	 Splice onto Arcane—Tap an untapped white creature you control. (As you
//	 cast an Arcane spell, you may reveal this card from your hand and pay
//	 its splice cost. If you do, add this card's effects to that spell.)"
//
// The splice cost taps a white creature the caster controls, named as the
// spell is cast (#4055, ADR 0202's 2026-10-19 amendment). It is not the {T}
// symbol, so a creature that came under the caster's control this turn may
// pay (CR 302.6), and paying does not target, so a hexproof one may too.
func init() {
	Register(Spec{
		OracleID:     "fbfaf009-8276-4687-a3db-5a71dc13f174",
		Name:         "Hundred-Talon Strike",
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0, "first strike")),
		Completeness: CompletenessFull,
		Splice: SpliceOntoArcaneCost(
			SpliceTapUntapped("an untapped white creature you control", Creature(), OfColor("W"))),
		Targets: TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (BoostUntilEOT{
				Target: t.ID,
				Power:  1,
				Label:  "Hundred-Talon Strike — +1/+0",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   t.ID,
				Keywords: []string{"first strike"},
				Label:    "Hundred-Talon Strike — gains first strike",
			}.Apply(ctx)
		},
	})
}
