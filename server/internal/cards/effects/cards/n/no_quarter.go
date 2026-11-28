package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// No Quarter — Enchantment {3}{R} (Tempest, S82 OLD batch 765, #5628):
//
//	"Whenever a creature becomes blocked by a creature with lesser power,
//	 destroy the blocking creature.
//	 Whenever a creature blocks a creature with lesser power, destroy the
//	 attacking creature."
//
// Two triggers over the same per-pair block event (EventBlock: the blocker is
// the event's card, the attacker its target), so a creature blocked by two
// creatures triggers once for each (CR 509.3d). They never both fire for one
// pair: the first needs the attacker's power to be greater than the blocker's,
// the second the blocker's to be greater than the attacker's. Power is
// compared as the block is declared, as printed (an intervening "with lesser
// power" is the trigger's condition, not a check on resolution), and the
// creature to destroy is named by the event and its object identity, so one
// that has left or come back as a new object is not destroyed (CR 400.7).
// Whoever controls the creatures, a creature of the Enchantment's controller
// is destroyed as well as anyone's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d238a5c1-f992-4f89-851a-de75b59f6793",
		Name:         "No Quarter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			noQuarterTrigger(true, "No Quarter — destroy the blocking creature (it has lesser power)"),
			noQuarterTrigger(false, "No Quarter — destroy the attacking creature (it has lesser power)"),
		},
	})
}

// noQuarterTrigger is one half. `destroyBlocker` is the first sentence: the
// attacker has the greater power and the blocker is destroyed; otherwise the
// blocker has the greater power and the attacker is destroyed.
func noQuarterTrigger(destroyBlocker bool, label string) game.TriggeredAbility {
	victim := func(ev game.Event) uuid.UUID {
		if destroyBlocker {
			return ev.CardID
		}
		return ev.Target
	}
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBlock},
		AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
			if ev.CardID == uuid.Nil || ev.Target == uuid.Nil {
				return false
			}
			blocker, ok := g.LookupCardForEffect(ev.CardID)
			if !ok {
				return false
			}
			attacker, ok := g.LookupCardForEffect(ev.Target)
			if !ok {
				return false
			}
			if destroyBlocker {
				return blocker.PowerForComparison() < attacker.PowerForComparison()
			}
			return attacker.PowerForComparison() < blocker.PowerForComparison()
		},
		Key: label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			id := victim(ev)
			item.Params.Object = game.ObjectRef{ID: id, Epoch: g.ObjectEpochForEffect(id)}
			return item
		},
		Effect: effects.DestroyTheObjectNamedByTheTrigger,
	}
}
