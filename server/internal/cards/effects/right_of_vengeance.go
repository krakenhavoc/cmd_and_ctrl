package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Right of Vengeance — Instant {3}{W} (Star Trek Commander, #4974):
//
//	"Until end of turn, whenever a creature deals combat damage to you,
//	 destroy that creature and create a 2/1 red Klingon Warrior creature
//	 token with haste."
//
// A repeating delayed triggered ability with a stated duration (CR 603.7b),
// controlled by the caster. It is an each-condition, so every creature that
// deals combat damage to the caster triggers it once, and the fired
// instance carries that creature with its CR 400.7 epoch: a creature that
// has left the battlefield, or left and come back, by resolution is not
// destroyed (CR 603.7c), but the Klingon Warrior is still created. Any
// creature counts, whoever controls it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e019f06-7e87-419e-b51d-e2cb86a6c7f1",
		Name:         "Right of Vengeance",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DelayedOnEvent{
				Label:     "Right of Vengeance — destroy that creature and create a 2/1 Klingon Warrior with haste",
				On:        []game.EventKind{game.EventDealDamage},
				Condition: trekB33CreatureHitsYouCondition,
				Body:      trekB33DestroyItAndMakeAKlingonBody,
				Repeats:   true,
			}.Apply(ctx)
		},
	})
}
