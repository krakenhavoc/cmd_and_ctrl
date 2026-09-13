package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Long River's Pull — Instant {U}{U} (EDHREC rank 1460):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, they draw a card before its other effects.)
//	 Counter target creature spell. If the gift was promised, instead
//	 counter target spell."
//
// Essence Scatter that can be Counterspell for the price of a card
// to an opponent.
//
// Sandbox simplification, declared — Into the Flood Maw's posture:
// the GIFT is not offered. Gift is a cast-time promise (CR 702.174)
// that needs a prompt in the cast flow and a per-cast flag the
// resolution reads, and neither exists. So the spell is always its
// base mode — counter target creature spell — which is the printed
// card with one option removed, weaker and never stronger, and the
// target picker makes the missing option visible: it offers creature
// spells only. When a cast-time promise lands, the second clause is a
// TargetSpell("target spell") swap plus a draw for the promisee.
func init() {
	Register(Spec{
		OracleID:     "f1993767-1d07-49c8-b8dc-04ec9840a999",
		Name:         "Long River's Pull",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The gift can't be promised, so the spell always counters a creature spell — never any other spell."},
		Targets:      TargetSpell("target creature spell", Creature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if ctx.Game.StackItemForEffect(item.Targets[0].ID) == nil {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
