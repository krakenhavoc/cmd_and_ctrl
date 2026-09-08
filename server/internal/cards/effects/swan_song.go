package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swan Song — "Counter target enchantment, instant, or sorcery
// spell. Its controller creates a 2/2 blue Bird creature token
// with flying."
//
// Composition: CounterTarget → CreateToken (under the countered
// spell's controller). Flying is cosmetic in S14 (keyword pipeline
// is S18); the token enters as a vanilla 2/2.
//
// Sandbox predicate note: like Counterspell / Negate, Swan Song
// does not enforce the "enchantment, instant, or sorcery" gate —
// announce-time UI lets the caster pick any stack spell. Target
// predicate enforcement lands in S20.
func init() {
	Register(Spec{
		OracleID: "8ddfc283-c9b4-41a5-af88-cf0068e986cc",
		Name:     "Swan Song",
		Targets:  TargetSpell("target enchantment, instant, or sorcery spell", Or(Enchantment(), Instant(), Sorcery())),
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
