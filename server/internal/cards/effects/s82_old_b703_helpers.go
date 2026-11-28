package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b703_helpers.go — shared shapes of S82 OLD batch 703 (Urza's Destiny
// Auras, #5621): the counter Auras that grow at their controller's upkeep, and
// the Auras that read those counters when the creature dies.

// S82OldB703MayPutCounterOnThisAtUpkeep is "At the beginning of your upkeep,
// you may put a <kind> counter on this Aura." (Momentum, Festering Wound,
// Incendiary, Archery Training, Private Research).
func S82OldB703MayPutCounterOnThisAtUpkeep(name, kind string) game.TriggeredAbility {
	return Optional(AtYourUpkeep(name+" — put a "+kind+" counter on this Aura", PutACounterOnThis(kind)),
		name+" — put a "+kind+" counter on it?")
}

// S82OldB703AttachedCreatureDealtDamageToPlayer is the condition "enchanted
// creature deals damage to a player": any damage, not only combat damage.
func S82OldB703AttachedCreatureDealtDamageToPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventDealDamage && ev.Amount > 0 &&
		source.IsAttachedTo(ev.Source) && g.PlayerByIDForEffect(ev.Target) != nil
}

// S82OldB703FesteringWoundDamage is "this Aura deals X damage to that player,
// where X is the number of infection counters on this Aura", at the upkeep of
// the enchanted creature's controller (the event's player).
func S82OldB703FesteringWoundDamage(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	x := CountersOnThisAsItResolves("infection")(ctx)
	if x <= 0 {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: item.Trigger.Event.Actor, Amount: x}.Apply(ctx)
}

// S82OldB703DisappearBounce is "Return enchanted creature and this Aura to
// their owners' hands.": the host is read first, then both are returned.
func S82OldB703DisappearBounce(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := B43AttachedCreature(g, item.SourceCardID)
	if ok {
		if err := (BounceToHand{Target: host}).Apply(ctx); err != nil {
			return err
		}
	}
	return BounceToHand{Target: item.SourceCardID}.Apply(ctx)
}

// S82OldB703PatternOfRebirth is "that creature's controller may search their
// library for a creature card, put that card onto the battlefield, then
// shuffle." The controller is the dead creature's, as it last was (CR 603.10a).
func S82OldB703PatternOfRebirth(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	info, ok := g.LastKnownPermanentForEffect(item.Trigger.Event.CardID)
	if !ok || info.Controller == uuid.Nil {
		return nil
	}
	return SearchLibrary{
		Player:    info.Controller,
		Predicate: IsCreatureCard,
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Optional:  true,
		Reason:    "Pattern of Rebirth — you may search for a creature card to put onto the battlefield",
	}.Apply(NewContext(g, item))
}

// S82OldB703IncendiaryDamage is "this Aura deals X damage to any target, where
// X is the number of fuse counters on this Aura", X read as the Aura last
// existed.
func S82OldB703IncendiaryDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := CountersOnThisAsItResolves("fuse")(ctx)
	ref, ok := ctx.SourceRef()
	if x <= 0 || !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		return DealDamage{SourceObject: &ref, Target: t.ID, Amount: x}.Apply(ctx)
	}
	return nil
}

// S82OldB703PrivateResearchDraw is "draw a card for each page counter on this
// Aura", counted as the Aura last existed.
func S82OldB703PrivateResearchDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := CountersOnThisAsItResolves("page")(ctx)
	if n <= 0 {
		return nil
	}
	return DrawCards{N: n}.Apply(ctx)
}

// S82OldB703ArcheryTrainingShot is the granted ability's "{T}: This creature
// deals X damage to target attacking or blocking creature, where X is the
// number of arrow counters on Archery Training." The creature is the source;
// the Aura that granted the ability (StackItem.GrantedBy) holds the counters.
func S82OldB703ArcheryTrainingShot(g *game.Game, item *game.StackItem) error {
	if item.GrantedBy == uuid.Nil {
		return nil
	}
	aura, ok := g.LookupCardForEffect(item.GrantedBy)
	if !ok {
		return nil
	}
	x := aura.Counters["arrow"]
	if x <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}.Apply(ctx)
	}
	return nil
}
