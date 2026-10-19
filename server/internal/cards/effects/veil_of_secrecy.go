package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veil of Secrecy — Instant — Arcane {1}{U}:
//
//	"Target creature gains shroud until end of turn and can't be blocked
//	 this turn.
//	 Splice onto Arcane—Return a blue creature you control to its owner's
//	 hand. (As you cast an Arcane spell, you may reveal this card from your
//	 hand and pay its splice cost. If you do, add this card's effects to
//	 that spell.)"
//
// The splice cost is a return to hand (#4055, ADR 0202's 2026-10-19
// amendment): the caster names a blue creature they control as the spell is
// cast, it returns with the spell on the stack, and a spell with no such
// creature to return cannot have this card spliced onto it (CR 118.3). The
// creature returned may be the spell's own target, so the spell's target
// can leave before the spell resolves, as it can for any cost.
//
// Shroud and "can't be blocked" are the two halves of the text, granted to
// the one target as the spell (or the spell this is spliced onto) resolves.
func init() {
	Register(Spec{
		OracleID:     "abfde51c-dfe7-4382-8aa2-f0a7ac1b7048",
		Name:         "Veil of Secrecy",
		Completeness: CompletenessFull,
		Splice: SpliceOntoArcaneCost(
			SpliceReturnToHand("a blue creature you control", Creature(), OfColor("U"))),
		Targets: TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (GrantKeywordUntilEOT{
				Target:   t.ID,
				Keywords: []string{"shroud"},
				Label:    "Veil of Secrecy — gains shroud until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return RestrictUntilEOT{
				Target:       t.ID,
				Restrictions: game.CantBeBlocked,
				Label:        "Veil of Secrecy — can't be blocked this turn",
			}.Apply(ctx)
		},
	})
}
