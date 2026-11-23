package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deny the Witch — Instant {1}{W}{U}{B} (Warhammer 40,000 Commander, #5948):
//
//	"Counter target spell, activated ability, or triggered ability. Its
//	 controller loses life equal to the number of creatures you control."
//
// Disallow's clause. The countered item's controller is read before it is
// removed from the stack (CR 608.2h), the creatures are counted as the
// spell resolves, and that controller loses the life whether the item was a
// spell or an ability. A target that is gone fizzles the whole spell
// (CR 608.2b), so no life is lost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ddff34d8-715b-4d6a-8ced-7ff5c4c911bc",
		Name:         "Deny the Witch",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpellOrAbility("target spell, activated ability, or triggered ability"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			victim := ctx.Game.StackItemForEffect(item.Targets[0].ID)
			if victim == nil {
				return nil
			}
			loser := victim.Controller
			n := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsCreature() })
			if err := (effects.CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Player: loser, Amount: -n}.Apply(ctx)
		},
	})
}
