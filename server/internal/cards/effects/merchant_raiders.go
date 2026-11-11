package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merchant Raiders — Creature — Human Pirate {3}{U}, 2/4:
//
//	"Whenever this creature or another Pirate you control enters, tap up
//	 to one target creature. That creature doesn't untap during its
//	 controller's untap step for as long as you control this creature."
//
// The hold is tied to Merchant Raiders, not to the Pirate that triggered
// it: it ends the moment the Raiders leaves or changes controller
// (TapAndHoldWhileYouControlThis). A Raiders that is already gone when
// the trigger resolves holds nothing (CR 400.7). "Up to one" lets the
// trigger go on the stack with no target.
func init() {
	Register(Spec{
		OracleID:     "8b388331-eaa7-41aa-99de-4451e51cb221",
		Name:         "Merchant Raiders",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{Targeting(
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventETB {
					return false
				}
				if ev.CardID == source.InstanceID {
					return true
				}
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && isPirate(c)
			}, "Merchant Raiders — tap up to one target creature; it doesn't untap while you control Merchant Raiders",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return TapAndHoldWhileYouControlThis(ctx, holdTargetIDs(ctx))
				}),
			TargetCreature("up to one target creature").WithCount(0, 1),
		)},
	})
}
