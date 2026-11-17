package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cacophony Scamp — Creature — Phyrexian Goblin Warrior {R}, 1/1 (Phyrexia:
// All Will Be One, S82):
//
//	"Whenever this creature deals combat damage to a player, you may
//	 sacrifice it. If you do, proliferate.
//	 When this creature dies, it deals damage equal to its power to any
//	 target."
//
// The combat trigger is Rotisserie Elemental's may-sacrifice: "if you do" is
// the sacrifice's tail, which proliferates only when the Scamp actually went
// (a Scamp that has already left does nothing). The dies trigger is Balduvian
// Berserker's: the target is chosen as the trigger goes on the stack and the
// damage is the power the Scamp had as it died (CR 603.10), dealt by that
// last-known object. A Scamp sacrificed to its own first ability dies, so it
// triggers its second.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4854a41e-db49-462f-ae57-0ba4005bae75",
		Name:         "Cacophony Scamp",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Cacophony Scamp — you may sacrifice it; if you do, proliferate",
				Do(MayChoice{
					Question: "Cacophony Scamp — sacrifice it to proliferate?",
					YesLabel: "Sacrifice",
					NoLabel:  "Decline",
					Then:     cacophonyScampAnswer,
				})),
			eerieDiesWithPower("Cacophony Scamp — it deals damage equal to its power to any target", TargetAny(), balduvianBerserkerDamage),
		},
	})
}

// cacophonyScampAnswer is the "yes": sacrifice the Scamp, and proliferate if
// it went. Keyed.
var cacophonyScampAnswer = MayYes("confirm/cacophony-scamp-sacrifice", func(ctx *Context, _ game.ConfirmAnswer) error {
	if !onBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return SacrificePermanent{Target: ctx.Source(), ThenTail: ctx.ItemTail(s82OneB06ProliferateIfDone)}.Apply(ctx)
})
