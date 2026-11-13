package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Long River Lurker — Creature — Frog Scout {2}{U}, 2/3 (Bloomburrow,
// #4467):
//
//	"Ward {1}
//	 Other Frogs you control have ward {1}.
//	 When this creature enters, target creature you control can't be blocked
//	 this turn. Whenever that creature deals combat damage this turn, you may
//	 exile it. If you do, return it to the battlefield under its owner's
//	 control."
//
// Ward {1} is the printed trigger (Ward), and the grant to the other Frogs is
// WardGranted over the Frog subtype (a word of its text, CR 612.2), so the
// ward stays with the object granting it and a ward that has already
// triggered is unaffected if the Lurker leaves (the ruling). The entry
// trigger targets a creature you control (the Lurker itself is legal): it
// can't be blocked this turn (RestrictUntilEOT), and a repeating delayed
// trigger (CR 603.7, DelayedOnEvent) bound to that creature as the object it
// is now watches for combat damage it deals to anything this turn, once per
// combat damage step however many objects and players it damaged in it
// (OncePerBatch, #5760; CR 510.2, CR 603.2c). Each
// instance asks its "you may" as it resolves and, on a yes, flickers the
// creature (Flicker, stdB27FlickerObject): exiled, then returned under its
// owner's control as a new object, so counters and Auras are gone and a
// token that is exiled does not come back (the rulings). The creature that
// has left or changed objects since is not touched.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c65c4d01-29da-47e6-993c-a265cf1f198c",
		Name:         "Long River Lurker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{1}"), "Long River Lurker — ward {1}"),
			WardGranted(WardMana("{1}"), "Long River Lurker — other Frogs you control have ward {1}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
						target.IsCreature() && target.HasSubtype(source.TextSubtype("Frog"))
				}),
			Targeting(WhenThisEnters("Long River Lurker — target creature you control can't be blocked this turn; whenever it deals combat damage this turn, you may exile it, then return it",
				longRiverLurkerEnters), TargetCreature("target creature you control", YouControl())),
		},
	})
}

// longRiverLurkerEnters makes the target unblockable and schedules the
// damage watcher.
func longRiverLurkerEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if err := (RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked,
		Label: "Long River Lurker — can't be blocked this turn"}).Apply(ctx); err != nil {
		return err
	}
	return DelayedOnEvent{
		Label:            "Long River Lurker — you may exile it, then return it to the battlefield under its owner's control",
		On:               []game.EventKind{game.EventDealDamage},
		Condition:        thatCreatureDealtCombatDamageCondition,
		CondParams:       game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
		OptionalQuestion: "Long River Lurker — exile that creature, then return it to the battlefield under its owner's control?",
		Body:             longRiverLurkerFlickerBody,
		Params:           game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
		Repeats:          true,
		OncePerBatch:     true,
	}.Apply(ctx)
}

var (
	// "Whenever <that creature> deals combat damage": an EventDealDamage,
	// combat, with amount, from CondParams.Object (the same object, CR
	// 400.7). To anything: a player, a planeswalker, a battle or a creature.
	thatCreatureDealtCombatDamageCondition = game.DelayedCondition("combat/that-creature-dealt-combat-damage",
		func(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
			if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || p.Object.ID == uuid.Nil || ev.Source != p.Object.ID {
				return false
			}
			c, ok := g.LookupCardForEffect(p.Object.ID)
			return ok && c.ObjectEpoch == p.Object.Epoch
		})

	// Exile that creature, then return it under its owner's control.
	longRiverLurkerFlickerBody = game.DelayedBody("long-river-lurker/flicker-that-creature",
		func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
			return stdB27FlickerObject(NewContext(g, item), p.Object)
		})
)
