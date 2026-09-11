package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Red Elemental Blast — Instant {R} (EDHREC rank 438):
//
//	"Choose one —
//	 • Counter target blue spell.
//	 • Destroy target blue permanent."
//
// One red mana against the colour that counters everything. Two
// targeted modes on a "choose one" (legal — the per-mode limit only
// bites when Max > 1; Rakdos Charm is the precedent), and "blue" is
// in the target clause on both: a non-blue spell or permanent is
// not a legal target, so the spell cannot even be pointed at one.
// Compare Pyroblast, whose clauses say "if it's blue" and can target
// anything.
//
// Colour comes from Scryfall's computed colours (mana-cost fallback
// for fixtures), so a colourless artifact creature is not blue and
// a blue-hybrid card is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "bb329a5c-b9f9-4973-a53f-090024146325",
		Name:     "Red Elemental Blast",
		Modes: ChooseOne(
			Mode("Counter target blue spell.", TargetSpell("target blue spell", OfColor("U"))),
			Mode("Destroy target blue permanent.", TargetPermanent("target blue permanent", OfColor("U"))),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			switch {
			case ctx.HasMode(0):
				return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
			case ctx.HasMode(1):
				if item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
			}
			return nil
		},
	})
}
