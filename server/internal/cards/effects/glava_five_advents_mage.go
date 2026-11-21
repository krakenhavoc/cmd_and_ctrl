package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glava, Five-Advents Mage — {5}{G}{G} Legendary Creature — Human Wizard
// 5/5 (Mystery Booster Commander, S82 GAP batch 3, #5948):
//
//	"Whenever you cast a permanent spell with {X} in its mana cost or
//	 activate an ability with {X} in its activation cost that isn't a
//	 mana ability, you may have the value of X become 5. Do this only once
//	 each turn."
//
// One optional trigger over two events (OnAny): a cast of a permanent
// spell whose own mana cost has an {X} (an {X} in an additional or
// alternative cost does not count), or the activation of a non-mana
// ability whose activation cost has an {X} mana symbol (Event.XCost, set
// only for a stack ability). The "you may" is the CR 603.5 prompt asked
// as the trigger resolves; "do this only once each turn" limits the DOING
// (b28DoneThisTurn), as Hela's does, so a declined offer leaves the
// turn's use for a later spell and the second of two triggers stacked
// before either resolved does nothing. The new value replaces X on the
// stack item (SetStackItemXForEffect), nothing is paid or refunded, and
// it follows the item's mana value and everything that reads X.
//
// No simplification.
func init() {
	const label = "Glava, Five-Advents Mage — have the value of X become 5"
	const done = "Glava, Five-Advents Mage — X became 5 this turn"
	Register(Spec{
		OracleID:     "94d9a596-96b3-4fd6-b3f9-62920929c668",
		Name:         "Glava, Five-Advents Mage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(OnAny([]game.EventKind{game.EventCast, game.EventActivateAbility},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Actor != source.Controller || b28DoneThisTurn(g, source.InstanceID, done) {
						return false
					}
					return glavaXThing(ev, g)
				}, label,
				func(g *game.Game, item *game.StackItem) error {
					if b28DoneThisTurn(g, item.SourceCardID, done) || item.Trigger == nil {
						return nil
					}
					ev := item.Trigger.Event
					id := ev.CardID
					if ev.Kind == game.EventActivateAbility {
						id = ev.StackItemID
					}
					if id == uuid.Nil {
						return nil
					}
					b28NoteDoneThisTurn(g, item.SourceCardID, done)
					g.SetStackItemXForEffect(id, 5)
					return nil
				}), "Glava, Five-Advents Mage — have the value of X become 5?"),
		},
	})
}

// glavaXThing is "a permanent spell with {X} in its mana cost" or "an
// ability with {X} in its activation cost that isn't a mana ability".
func glavaXThing(ev game.Event, g *game.Game) bool {
	switch ev.Kind {
	case game.EventCast:
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && !c.IsInstant() && !c.IsSorcery() && socB9HasXInManaCost(g, ev.CardID)
	case game.EventActivateAbility:
		return ev.XCost && ev.StackItemID != uuid.Nil
	}
	return false
}
