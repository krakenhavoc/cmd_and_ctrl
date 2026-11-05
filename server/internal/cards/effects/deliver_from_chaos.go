package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deliver from Chaos — Instant {3}{W}{B} (Star Trek Commander, #4974):
//
//	"Exile target artifact or creature. Create X 1/1 colorless Borg
//	 artifact creature tokens, where X is that permanent's mana value."
//
// X is the permanent's mana value as it last existed on the battlefield
// (CR 608.2h), read just before it is exiled, so a token that is not a
// copy of anything gives 0 and a commander its owner sends to the command
// zone instead still counts. The tokens are the caster's. With the target
// gone or illegal the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "edea104c-b563-4fae-9e74-82164a8e090b",
		Name:         "Deliver from Chaos",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				tail := ctx.ItemTail(trekB39DeliverExiled)
				tail.Nums = []int{c.ManaValue()}
				return ExileTarget{Target: t.ID, ThenTail: tail}.Apply(ctx)
			}
			return nil
		},
	})
}

// trekB39DeliverExiled creates the Borgs once the exile has settled. X was
// read before the exile, as the permanent last existed (CR 608.2h), and the
// tokens do not depend on where the card went: the text has no "if you do".
// An on-disk key: never renamed, never reused.
var trekB39DeliverExiled = CardTail("deliver-from-chaos/after-exile", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Nums) == 0 || o.Tail.Nums[0] <= 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 colorless Borg artifact"), N: o.Tail.Nums[0]}.Apply(ctx)
})
