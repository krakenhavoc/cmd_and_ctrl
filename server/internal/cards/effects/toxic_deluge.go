package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Toxic Deluge — Sorcery {2}{B}:
//
//	"As an additional cost to cast this spell, pay X life.
//	 All creatures get -X/-X until end of turn."
//
// The three-mana answer to anything. It is the wipe that gets played
// over Damnation in a deck that can afford the life, for the reason
// spelled out below: it is not destruction.
//
// # It is NOT a DestroyAllMatching, and that is the whole card
//
// Toxic Deluge shrinks creatures and lets the zero-toughness
// state-based action (CR 704.5f) do the killing. Four consequences,
// every one of them why someone paid for the card:
//
//   - INDESTRUCTIBLE DOES NOT SAVE ANYTHING. An indestructible 3/3
//     at -4/-4 is a -1/-1 and it is in the graveyard. (Inert today:
//     indestructible is not modelled — see mass.go — but the card is
//     written so that it stays correct when it is.)
//   - Regeneration and totem armor likewise do nothing.
//   - X is chosen by the caster, so the sweep is aimed: X=2 kills the
//     tokens and leaves the fatties, X=7 kills everything.
//   - A creature that survives stays shrunk for the turn, so a 5/5 at
//     -4/-4 blocks like the 1/1 it currently is.
//
// So the effect is a layer-7c continuous effect through
// BoostUntilEOT with negative values, not a zone change. CR 611.2c
// snapshots the affected set at resolution, which is correct: a
// creature cast after the Deluge resolves is unaffected, and that is
// how the card is played.
//
// # X, paid in life
//
// "Pay X life" is an ADDITIONAL COST (CR 601.2f), paid at cast with
// the spell already on the stack — so the life loss happens before
// the spell resolves and anything watching it triggers above the
// spell. It is not part of the resolution and it is paid even if the
// spell is countered. PayXLifeCost() carries that; CR 119.4's "only
// if your life total is at least X" is enforced at announce.
//
// The X of the life payment and the X of the -X/-X are the same
// announced number by definition. See game/additional_cost.go for
// why it rides the existing XValue slot rather than a second one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "afaef788-34d1-460b-b884-9d7ae6ddeb18",
		Name:           "Toxic Deluge",
		AdditionalCost: PayXLifeCost(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			return BoostUntilEOT{
				Match:     Creature(),
				Power:     -x,
				Toughness: -x,
				Label:     "Toxic Deluge — -X/-X",
			}.Apply(ctx)
		},
	})
}
