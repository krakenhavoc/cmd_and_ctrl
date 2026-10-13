package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silver Scrutiny — Sorcery {X}{U}{U}:
//
//	"You may cast this spell as though it had flash if X is 3 or less.
//	 Draw X cards."
//
// The draw is complete. The flash clause is NOT: CR 601.3c lets the
// caster begin the cast at instant speed and then requires the
// announced X to be 3 or less, so the timing window depends on a number
// chosen after the window is checked. The only shapes for "as though it
// had flash" the engine has are a flat grant (CastKindAsThoughFlash) and
// an alternative cost that opens the window (AlternativeCost.AsThoughFlash,
// FlashForMore); neither can bound the X claimed under it, and shipping
// either unbounded would let a player draw ten cards at instant speed,
// which is stronger than printed. So the card is a plain sorcery, which
// is weaker than printed and says so in its caveat. The seam is a
// per-offer X ceiling on AlternativeCost, read by the announce gate, the
// legal-move enumerator and the view's x_max.
func init() {
	Register(Spec{
		OracleID:     "2b0b4fc1-9ad6-4259-a0a2-9a14484c246e",
		Name:         "Silver Scrutiny",
		XMatters:     true,
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The flash clause isn't implemented — Silver Scrutiny can only be cast at sorcery speed, even when X is 3 or less."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: ctx.X()}.Apply(ctx)
		},
	})
}
