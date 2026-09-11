package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swan Song — "Counter target enchantment, instant, or sorcery
// spell. Its controller creates a 2/2 blue Bird creature token
// with flying."
//
// Composition: CounterTarget → CreateToken (under the countered
// spell's controller).
//
// No simplification remains, and both notes that used to sit here
// outlived their fixes (corrected in the #338 sweep):
//
//   - The Bird's flying is real. WhiteBirdToken carries it on
//     Card.Keywords (S21 sub-PR 1).
//   - The "enchantment, instant, or sorcery" gate IS enforced. The
//     Targets clause below declares the predicate, which S20's
//     structured targeting checks at announce (CR 601.2c) and
//     re-checks on resolution (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "8ddfc283-c9b4-41a5-af88-cf0068e986cc",
		Name:         "Swan Song",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target enchantment, instant, or sorcery spell", Or(Enchantment(), Instant(), Sorcery())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			// Pull the countered spell's controller BEFORE CounterTarget
			// fires — the StackMeta entry is deleted by CounterTarget,
			// and we need the controller to route the token.
			var victim *game.StackItem
			if g := ctx.Game; g != nil {
				// StackMeta is a *game.Game field; reach via the public
				// FindCardZoneForEffect / side channel. Here we call into
				// the protocol-compatible helper directly.
				victim = g.StackItemForEffect(stackID)
			}
			if err := (CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			tokenOwner := ctx.Controller()
			if victim != nil {
				tokenOwner = victim.Controller
			}
			return CreateToken{
				Controller: tokenOwner,
				Template:   WhiteBirdToken(),
				N:          1,
			}.Apply(ctx)
		},
	})
}
