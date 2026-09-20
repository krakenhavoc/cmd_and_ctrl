package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hatred — Instant {3}{B}{B}:
//
//	"As an additional cost to cast this spell, pay X life.
//	 Target creature gets +X/+0 until end of turn."
//
// The five-mana "you win now" instant, and the reason a deck that
// gains life plays it: the only limit on X is the life total, so a
// player at 38 can hand an unblocked 1/1 a lethal +37/+0 at instant
// speed. It is the aggressive half of the same life-as-a-resource
// idea Toxic Deluge is the defensive half of.
//
// # The same clause Toxic Deluge prints, aimed at one creature
//
// "Pay X life" is an ADDITIONAL COST (CR 601.2f), paid at cast with
// the spell already on the stack. Three consequences the card is
// actually played for, and all three come free with PayXLifeCost():
//
//   - The life is gone BEFORE the spell resolves, so a Blood Artist
//     style payoff — or anything watching a life total drop — sees it
//     above the spell and resolves first.
//   - It is paid whether or not the spell resolves. Counter the
//     Hatred and the life does not come back; remove the creature in
//     response and it does not either.
//   - CR 119.4 caps X at the caster's life total, enforced at
//     announce, so the spell cannot be used to kill yourself for a
//     bigger pump.
//
// The X of the payment and the X of the +X/+0 are the same announced
// number by definition — see game/additional_cost.go for why it
// rides the existing XValue slot rather than a second one.
//
// The pump itself is layer 7c (BoostUntilEOT), so it stacks with an
// anthem rather than overwriting it and expires at cleanup rather
// than at the end step. Toughness is deliberately untouched: Hatred
// makes a creature lethal, not durable, and a 1/1 handed +9/+0 still
// dies to a single ping.
//
// X = 0 is a legal announcement (CR 602.2b) and does nothing, which
// is correct: nothing is paid and the creature gets +0/+0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "40601061-1b25-4db8-8b99-33324ce945cf",
		Name:           "Hatred",
		Completeness:   CompletenessFull,
		XMatters:       true,
		AdditionalCost: PayXLifeCost(),
		Targets:        TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return BoostUntilEOT{
					Target: t.ID,
					Power:  x,
					Label:  "Hatred — +X/+0",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
