package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b383_helpers.go — shared bodies for Commander 2013 batch 383
// (#5560): the Curses and Flickerform.

// oldB383AttacksEnchantedPlayer is the condition of every Curse that watches
// attackers: the attack is at the player this Curse enchants.
func oldB383AttacksEnchantedPlayer(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && source.AttachedTo.Kind == game.TargetPlayer &&
		source.AttachedTo.ID == ev.Target && ev.Actor != uuid.Nil
}

// OldB383WhenAPlayerAttacksEnchantedPlayer is "Whenever a player attacks
// enchanted player with one or more creatures, that attacking player <does
// X>": one trigger per declaration per attacking player (CR 506.3), so two
// creatures from the same player earn one. The attacking player, whoever they
// are (the Curse's controller included), controls the ability, so any "may"
// and any target are theirs (CR 603.3a-b), Curse of Chaos, Inertia and
// Shallow Graves.
func OldB383WhenAPlayerAttacksEnchantedPlayer(label string, effect Effect) game.TriggeredAbility {
	t := game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventAttack},
		Key:       label,
		AppliesTo: oldB383AttacksEnchantedPlayer,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Controller, item.Owner = ev.Actor, ev.Actor
			return item
		},
		Effect: effect,
	}
	t.OncePerBatch = true
	t.BatchKey = func(ev game.Event, _ *game.Card, _ *game.Game) string {
		return ev.Actor.String()
	}
	return t
}

// OldB383WheneverACreatureAttacksEnchantedPlayer is "Whenever a creature
// attacks enchanted player, <effect>": one trigger per attacking creature.
// The effect is the Curse's controller's, unless attackerControls is set, in
// which case the attacking creature's controller controls the ability
// ("its controller gains 1 life", Curse of the Forsaken).
func OldB383WheneverACreatureAttacksEnchantedPlayer(label string, attackerControls bool, effect Effect) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventAttack},
		Key:       label,
		AppliesTo: oldB383AttacksEnchantedPlayer,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			if attackerControls {
				item.Controller, item.Owner = ev.Actor, ev.Actor
			}
			return item
		},
		Effect: effect,
	}
}

// flickerformBody is Flickerform's activated ability as it resolves: the
// enchanted creature and every Aura attached to it (Flickerform itself
// included) are exiled as one event, creature first, and from that exile's
// continuation the delayed return is scheduled for the next end step. The
// host is read off Flickerform as it is now, or as it last existed on the
// battlefield (CR 608.2h) if it has gone.
func OldB383FlickerformExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var host uuid.UUID
	if info, ok := ctx.SourcePermanent(); ok && info.AttachedTo.Kind == game.TargetCard {
		host = info.AttachedTo.ID
	}
	if host == uuid.Nil || !OnBattlefield(g, host) {
		return nil
	}
	ids := S82OldB366WithAttachedAuras(ctx, []uuid.UUID{host})
	tail := ctx.ItemTail(oldB383FlickerformSchedule)
	tail.Carry = []uuid.UUID{host}
	return g.ExileCardsKeyedForEffect(ids, tail)
}

// oldB383FlickerformSchedule runs once the exile has settled. When the
// creature itself reached exile, the delayed return is scheduled over the
// cards that did; when it didn't, nothing can return it, so nothing is.
var oldB383FlickerformSchedule = CardTail("flickerform/schedule-return", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 || len(o.IDs) == 0 || o.IDs[0] != o.Tail.Carry[0] {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:    game.StepEnd,
		Label: "Flickerform — return the exiled creature and Auras",
		Cards: o.IDs,
		Body:  oldB383FlickerformReturnBody,
	}.Apply(ctx)
})

var oldB383FlickerformReturnBody = game.SimpleDelayedBody("flickerform/return", func(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	ctx := NewContext(g, item)
	if z := g.FindCardZoneForEffect(ids[0]); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	c, ok := g.LookupCardForEffect(ids[0])
	if !ok || c.IsToken() || !c.IsPermanent() {
		return nil
	}
	tail := ctx.ItemTail(oldB383FlickerformReattach)
	tail.Carry = append([]uuid.UUID(nil), ids[1:]...)
	return g.PutOntoBattlefieldTogetherKeyedForEffect([]game.BatchEntry{{CardID: ids[0], From: game.ZoneExile}}, game.ZoneEntryOptions{}, tail)
})

// oldB383FlickerformReattach is "If you do, return the other cards exiled this
// way to the battlefield under their owners' control attached to that
// creature": it runs with the creature's new ID, and only if it entered.
var oldB383FlickerformReattach = CardTail("flickerform/return-auras", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	g := ctx.Game
	var entries []game.BatchEntry
	for _, id := range o.Tail.Carry {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneExile})
		}
	}
	if len(entries) == 0 {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(entries,
		game.ZoneEntryOptions{AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: o.IDs[0]}}, game.KeyedTail{})
})
