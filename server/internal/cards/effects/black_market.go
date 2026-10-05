package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Black Market — Enchantment {3}{B}{B}:
//
//	"Whenever a creature dies, put a charge counter on this
//	 enchantment.
//	 At the beginning of your first main phase, add {B} for each
//	 charge counter on this enchantment."
//
// The first line counts every creature that dies, anyone's, the
// Black Market's own controller included (diedCreature, graveyard
// only: a bounced or exiled creature does not count). The second is
// Hulking Raptor's "first main phase" trigger: EventBeginPrecombatMain
// is the first main phase on an ordinary turn, and the mana arrives
// when the trigger resolves, so it empties with the pool at the end
// of the step like any other (CR 106.4). The count is read when the
// trigger RESOLVES from the live permanent, so a counter added in
// response is included; a Market that has left the battlefield by
// then adds nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "21338b71-37f5-4121-9b84-565025ebcd17",
		Name:         "Black Market",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := diedCreature(ev, g)
				return ok
			}, "Black Market — put a charge counter", putChargeCounterOnThis),
			AtYourPrecombatMain("Black Market — add {B} for each charge counter", func(g *game.Game, item *game.StackItem) error {
				if sourceIsNewObject(g, item) {
					return nil
				}
				src, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || src.Counters["charge"] == 0 {
					return nil
				}
				return g.AddManaForEffect(item.Controller, item.SourceCardID, strings.Repeat("{B}", src.Counters["charge"]))
			}),
		},
	})
}
