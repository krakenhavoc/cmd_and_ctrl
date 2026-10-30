package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Captain Marvel, Apex Avenger — Legendary Creature — Human Kree Hero
// {5}{R}{W}, 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"Flying, double strike, indestructible
//	 Whenever you put one or more counters on another creature, if it's
//	 not a Kree, you may put the same number and kind of counters on
//	 Captain Marvel."
//
// The engine emits one EventCounterPlaced per permanent and kind per
// placement, so "one or more counters" is the event; the number is its
// delta (b33CountersPlacedDelta) and the kind its label. "You put" is the
// placement's own actor (ADR 0056). "Another creature" is any controller's
// creature but Captain Marvel. The intervening if (CR 603.4) is read at
// the trigger and again at resolution, a creature that has since left
// being judged by its last-known types (CR 608.2h). The counters Captain
// Marvel gets are put on her, not on another creature, so they never
// trigger the ability again.
//
// No simplification.
func init() {
	t := Optional(On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return msc17CountersOnAnotherNonKree(ev, source, g)
	}, "Captain Marvel — put the same number and kind of counters on Captain Marvel",
		func(g *game.Game, item *game.StackItem) error {
			if item.Trigger == nil {
				return nil
			}
			ev := item.Trigger.Event
			src, ok := g.LookupCardForEffect(item.SourceCardID)
			if !ok || !msc17CountersOnAnotherNonKree(ev, &src, g) {
				return nil // CR 603.4: re-checked on resolution
			}
			n := b33CountersPlacedDelta(ev, ev.Label, g)
			if n <= 0 {
				return nil
			}
			return AddCounter{Target: item.SourceCardID, Kind: ev.Label, N: n}.Apply(NewContext(g, item))
		}), "Captain Marvel — put the same number and kind of counters on Captain Marvel?")
	Register(Spec{
		OracleID:        "6d86d783-7116-4426-8c0f-718c2d904887",
		Name:            "Captain Marvel, Apex Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "double strike", "indestructible"},
		Triggered:       []game.TriggeredAbility{t},
	})
}

// msc17CountersOnAnotherNonKree is the trigger's condition and its
// intervening if: the controller put one or more counters on a creature
// other than the source, and that creature is not a Kree.
func msc17CountersOnAnotherNonKree(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCounterPlaced || ev.Target == source.InstanceID {
		return false
	}
	if b33CountersPlacedDelta(ev, ev.Label, g) <= 0 || b11CounterPlacerOf(ev, g) != source.Controller {
		return false
	}
	return msc17IsCreatureNotOfType(g, ev.Target, source.TextSubtype("Kree"))
}

// msc17IsCreatureNotOfType reports that the permanent is a creature that
// is not of the creature type: as it is now while it is on the
// battlefield, and as it last existed there once it has gone.
func msc17IsCreatureNotOfType(g *game.Game, id uuid.UUID, subtype string) bool {
	if c, ok := g.LookupCardForEffect(id); ok && onBattlefield(g, id) {
		return c.IsCreature() && !c.HasSubtype(subtype)
	}
	ref, ok := g.PermanentRefForEffect(id)
	if !ok {
		return false
	}
	info, ok := g.PermanentForEffect(ref)
	if !ok {
		return false
	}
	isCreature := false
	for _, ty := range info.Characteristic.Types {
		if ty == "Creature" {
			isCreature = true
		}
	}
	for _, s := range info.Characteristic.Subtypes {
		if s == subtype {
			return false
		}
	}
	return isCreature
}
