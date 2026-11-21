package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Desert Warfare — Enchantment {3}{G} (Modern Horizons 3 Commander, #5948):
//
//	"Whenever you sacrifice a Desert and whenever a Desert card is put into
//	 your graveyard from your hand or library, put that card onto the
//	 battlefield under your control at the beginning of your next end step.
//	 At the beginning of combat on your turn, if you control five or more
//	 Deserts, create that many 1/1 red, green, and white Sand Warrior
//	 creature tokens. They gain haste."
//
// The first ability is one trigger over three events: a sacrifice by you
// (the card is already in the graveyard as the trigger resolves), a discard
// from your hand and a mill from your library (each path into a graveyard
// emits exactly one of them). The card is named as the trigger resolves
// and a delayed trigger returns it at the beginning of YOUR next end step
// if it is still in that graveyard (CR 603.7c); a token Desert is not a
// card and is not returned. The second is checked as combat begins and
// again as the trigger resolves (CR 603.4), and counts the Deserts you
// control then; the Sand Warriors have haste for as long as they exist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de301ef8-8ba0-423d-9117-a8492a58d01b",
		Name:         "Desert Warfare",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventSacrifice, game.EventDiscardCard, game.EventMill},
				desertWarfareWhen,
				"Desert Warfare — put that card onto the battlefield at the beginning of your next end step",
				desertWarfareSchedule),
			desertWarfareCombat(),
		},
	})
}

// desertWarfareCount is the Deserts a player controls.
func desertWarfareCount(g *game.Game, controller uuid.UUID) int {
	return effects.CountControlled(g, controller, func(c game.Card) bool { return c.HasSubtypeWord("Desert") })
}

// desertWarfareWhen is "you sacrifice a Desert" and "a Desert card is put
// into your graveyard from your hand or library".
func desertWarfareWhen(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || effects.IsToken(c) || !c.HasSubtypeWord("Desert") {
		return false
	}
	switch ev.Kind {
	case game.EventSacrifice:
		return ev.Actor == source.Controller
	case game.EventDiscardCard, game.EventMill:
		return ev.NewZone == game.ZoneGraveyard && ev.OldZone != game.ZoneBattlefield && c.Owner == source.Controller
	}
	return false
}

func desertWarfareSchedule(g *game.Game, item *game.StackItem) error {
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		Label:              "Desert Warfare — put the Desert card onto the battlefield",
		ControllerTurnOnly: true,
		Cards:              []uuid.UUID{id},
		Body:               effects.GapB45DesertWarfareReturnBody,
	}.Apply(effects.NewContext(g, item))
}

// desertWarfareCombat is the second ability: the combat trigger gated, as
// it begins, on five Deserts.
func desertWarfareCombat() game.TriggeredAbility {
	t := effects.AtBeginningOfYourCombat("Desert Warfare — create a 1/1 Sand Warrior for each Desert you control",
		func(g *game.Game, item *game.StackItem) error {
			n := desertWarfareCount(g, item.Controller)
			if n < 5 { // CR 603.4
				return nil
			}
			tmpl := effects.TokenCard("1/1 red and green and white Sand Warrior")
			tmpl.Keywords = append(tmpl.Keywords, "haste")
			return effects.CreateToken{Template: tmpl, N: n}.Apply(effects.NewContext(g, item))
		})
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return when(ev, source, ch, g) && desertWarfareCount(g, source.Controller) >= 5
	}
	return t
}
