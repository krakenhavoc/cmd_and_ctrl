package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grave Betrayal — Enchantment {5}{B}{B} (Return to Ravnica, #5566):
//
//	"Whenever a creature you don't control dies, return it to the battlefield
//	 under your control with an additional +1/+1 counter on it at the beginning of
//	 the next end step. That creature is a black Zombie in addition to its other
//	 colors and types."
//
// The trigger fires for a creature that was not under the Betrayal's controller's
// control as it last existed (CR 603.10a). When it resolves it schedules a delayed
// trigger (CR 603.7) pinned to that card in the graveyard as it is then: if the
// card has left the graveyard by the end step, or left and came back as another
// object, nothing returns (CR 400.7). A token cannot return. The creature returns
// under the delayed trigger's controller, the Betrayal's controller as it
// resolved (CR 603.7d), with the counter as an "enters with" clause, and is a
// black Zombie in addition for as long as it is that object.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "892c8fd1-9553-4ced-84ec-9fc35c314c76",
		Name:         "Grave Betrayal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !dead.IsToken() && effects.LeftUnderControlOf(ev, dead) != source.Controller
			}, "Grave Betrayal — return it at the beginning of the next end step", graveBetrayalSchedule),
		},
	})
}

var graveBetrayalReturnBody = game.DelayedBody("grave-betrayal/return", graveBetrayalReturn)

func graveBetrayalSchedule(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		Label:  "Grave Betrayal — return it to the battlefield under your control",
		Cards:  []uuid.UUID{id},
		Body:   graveBetrayalReturnBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(effects.NewContext(g, item))
}

func graveBetrayalReturn(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	entered, err := g.ReturnFromGraveyardWithCountersForEffect(id, item.Controller, false, map[string]int{game.CounterPlusOne: 1})
	if err != nil {
		return effects.NothingIfGone(err)
	}
	if entered == uuid.Nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	return effects.ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Zombie")...), game.AddColorsMod(ctx.TextColors("B")...)},
		Duration: g.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Grave Betrayal — it is a black Zombie in addition to its other colors and types",
	}.Apply(ctx)
}
