package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b615_helpers.go — S82 OLD batch 615 (Eighth Edition enchantments,
// #5607): Cowardice, Death Pits of Rath, Sacred Ground and Warped Devotion.

// S82OldB615TargetedCreature is the condition of "whenever a creature becomes
// the target of a spell or ability" (Cowardice): the target is a creature on
// the battlefield (a player or another kind of object never qualifies).
func S82OldB615TargetedCreature(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && OnBattlefield(g, ev.CardID) && c.IsCreature()
}

// s82OldB615NamedOnBattlefield is the object a trigger named when it was
// created (Params.Object), when it is still that object on the battlefield
// (CR 400.7).
func s82OldB615NamedOnBattlefield(g *game.Game, item *game.StackItem) (game.Card, bool) {
	c, ok := g.LookupCardForEffect(item.Params.Object.ID)
	if !ok || c.ObjectEpoch != item.Params.Object.Epoch || !OnBattlefield(g, c.InstanceID) {
		return game.Card{}, false
	}
	return c, true
}

// S82OldB615BounceTheNamedCreature returns the creature the trigger names to
// its owner's hand.
func S82OldB615BounceTheNamedCreature(g *game.Game, item *game.StackItem) error {
	c, ok := s82OldB615NamedOnBattlefield(g, item)
	if !ok {
		return nil
	}
	return BounceToHand{Target: c.InstanceID}.Apply(NewContext(g, item))
}

// S82OldB615DestroyTheNamedCreatureNoRegen destroys the creature the trigger
// names; it can't be regenerated (CR 701.19c).
func S82OldB615DestroyTheNamedCreatureNoRegen(g *game.Game, item *game.StackItem) error {
	c, ok := s82OldB615NamedOnBattlefield(g, item)
	if !ok {
		return nil
	}
	return DestroyTarget{Target: c.InstanceID, CantBeRegenerated: true}.Apply(NewContext(g, item))
}

// S82OldB615OpponentPutYourLandInYourGraveyard is Sacred Ground's condition: a
// land was put into the source controller's graveyard from the battlefield by
// a spell or ability an opponent controls (Event.LeaveController).
func S82OldB615OpponentPutYourLandInYourGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.OldZone != game.ZoneBattlefield || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	if ev.LeaveController == uuid.Nil || ev.LeaveController == source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Owner == source.Controller && LeftAsType(ev, c, "land")
}

// S82OldB615ReturnTheNamedCardToTheBattlefield returns the card the trigger
// names (EventCardTriggerItem) from its owner's graveyard to the battlefield,
// if it is still that object there.
func S82OldB615ReturnTheNamedCardToTheBattlefield(g *game.Game, item *game.StackItem) error {
	id, ok := DeadCardInGraveyard(g, item)
	if !ok {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(NewContext(g, item))
}

// S82OldB615PermanentReturnedToAHand is Warped Devotion's condition: a
// permanent moved from the battlefield to a hand.
func S82OldB615PermanentReturnedToAHand(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.OldZone == game.ZoneBattlefield && ev.NewZone == game.ZoneHand
}

// S82OldB615BuildNamingTheHandOwner is Warped Devotion's Build: it names the
// player whose hand the permanent was returned to (its owner, CR 400.3), read
// as the trigger is created. A card that is gone (a token) names nobody and
// the trigger is not put on the stack.
func S82OldB615BuildNamingTheHandOwner(label string) func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return nil
		}
		item := game.NewTriggeredItem(source, label)
		item.Params.Player = c.Owner
		return item
	}
}

// S82OldB615ThatPlayerDiscards makes the player the trigger names discard a
// card of their choice.
func S82OldB615ThatPlayerDiscards(g *game.Game, item *game.StackItem) error {
	if item.Params.Player == uuid.Nil {
		return nil
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: item.Params.Player, Source: item.SourceCardID, N: 1})
	return nil
}
