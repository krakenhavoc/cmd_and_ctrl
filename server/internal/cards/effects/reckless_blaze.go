package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reckless Blaze — Sorcery — Lesson {3}{R}{R} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Reckless Blaze deals 5 damage to each creature. Whenever a creature
//	 you control dealt damage this way dies this turn, add {R}."
//
// Anger of the Gods' damage (DealDamageToEachThen) with a per-recipient
// rider: each creature actually dealt damage (a prevented blow is not
// "dealt damage this way") gets a delayed "when that creature dies this
// turn" about it as the object it is now (CR 603.7a, 603.7c). The rider
// is a "whenever", but each such creature can die only once as that
// object, so one delayed trigger per creature is the same ability. It
// fires only when the creature dies under the controller of Reckless
// Blaze (its last-known controller, CR 603.10a), and adds {R} to that
// player; the "this turn" duration ends it at cleanup (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75137acb-dc8e-439b-8d84-c5cf682ff6bc",
		Name:         "Reckless Blaze",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 5}},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, c := range MatchingBattlefield(ctx, Creature()) {
				ids = append(ids, c.InstanceID)
			}
			return DealDamageToEachThen(ctx, ids, 5, itemRecipient(item, recklessBlazeDealt))
		},
	})
}

var (
	// recklessBlazeDealt is the per-recipient rider: a creature dealt
	// more than 0 damage gets the delayed death watch. An on-disk key.
	recklessBlazeDealt = recipientTail("recipient/reckless-blaze-watch-death", func(ctx *Context, target uuid.UUID, dealt int, _ game.TailOutcome) error {
		if dealt <= 0 {
			return nil
		}
		return whenThatCreatureDiesThisTurnIf(ctx, target, "Reckless Blaze — a creature you control dealt damage this way died: add {R}",
			recklessBlazeDiedUnderYourControl, recklessBlazeAddRed)
	})
	// recklessBlazeDiedUnderYourControl is thatCreatureDied narrowed to a
	// creature that died under the delayed trigger's controller's
	// control. An on-disk key.
	recklessBlazeDiedUnderYourControl = game.DelayedEachCondition("dies/reckless-blaze-that-creature-under-your-control",
		func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, p game.EffectParams) []uuid.UUID {
			ids := thatCreatureDied(ev, dt, g, p)
			if len(ids) == 0 {
				return nil
			}
			dead, ok := diedCreature(ev, g)
			if !ok || leftUnderControlOf(ev, dead) != dt.Controller {
				return nil
			}
			return ids
		})
	// recklessBlazeAddRed is the fired trigger's body: add {R}. An
	// on-disk key.
	recklessBlazeAddRed = game.DelayedBody("reckless-blaze/add-red", simpleBody(func(g *game.Game, item *game.StackItem) error {
		return AddMana{Produced: "{R}"}.Apply(NewContext(g, item))
	}))
)
