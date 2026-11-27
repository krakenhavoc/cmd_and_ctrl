package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b605_helpers.go — Mirrodin, batch 605 (#5606): Grim Reminder's
// search, Sword of Kaldra's trigger and Scythe of the Wretched's return.

// OldB605GrimReminderSearch is Grim Reminder's spell: "Search your library
// for a nonland card and reveal it." The card stays where it is (a search
// that only reveals, CR 701.23, HoldFound); the tail charges the opponents
// and shuffles.
func OldB605GrimReminderSearch(_ *game.StackItem, ctx *Context) error {
	return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:    ctx.Controller(),
		Source:    ctx.Source(),
		Pred:      func(c game.Card) bool { return !c.HasCardType("land") },
		Dest:      game.ZoneLibrary,
		HoldFound: true,
		Limit:     1,
		Reveal:    true,
		Reason:    "Grim Reminder — search your library for a nonland card",
		ThenTail:  ctx.ItemTail(oldB605GrimReminderTail),
	})
}

// oldB605GrimReminderTail: each opponent who cast a spell this turn with the
// revealed card's name loses 6 life, then the library is shuffled. A search
// that found nothing only shuffles. An on-disk key.
var oldB605GrimReminderTail = CardTail("then/old-b605-grim-reminder", func(ctx *Context, o game.TailOutcome) error {
	g := ctx.Game
	if err := g.ShuffleLibraryForEffect(ctx.Controller()); err != nil {
		return err
	}
	if len(o.IDs) == 0 {
		return nil
	}
	found, ok := g.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	var losers []uuid.UUID
	for _, opp := range ctx.Opponents() {
		if OldB605CastASpellNamed(g, opp, found.Name) {
			losers = append(losers, opp)
		}
	}
	if len(losers) == 0 {
		return nil
	}
	return g.LoseLifeEachKeyedForEffect(ctx.Source(), losers, 6, game.KeyedTail{})
})

// OldB605CastASpellNamed reports whether `player` cast a spell this turn
// whose card has `name`: the turn's cast events, read as the cards they
// named (a countered spell is still the card it was).
func OldB605CastASpellNamed(g *game.Game, player uuid.UUID, name string) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventCast || ev.Actor != player {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && c.Name == name {
			return true
		}
	}
	return false
}

// OldB605EquippedCreatureDealtDamageToACreature is the condition of Sword of
// Kaldra's trigger: the creature this Equipment is attached to dealt damage
// (combat or not) to a creature that is still on the battlefield.
func OldB605EquippedCreatureDealtDamageToACreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source == uuid.Nil || !source.IsAttachedTo(ev.Source) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	if !ok || !c.IsCreature() {
		return false
	}
	z := g.FindCardZoneForEffect(ev.Target)
	return z != nil && z.Kind == game.ZoneBattlefield
}

// OldB605WheneverEquippedCreatureDealsDamageToACreature is Sword of Kaldra's
// trigger: "Whenever equipped creature deals damage to a creature, exile that
// creature." The victim is named as the object it was when the damage was
// dealt, so one that died and came back is not exiled (CR 400.7).
func OldB605WheneverEquippedCreatureDealsDamageToACreature(label string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventDealDamage},
		AppliesTo: OldB605EquippedCreatureDealtDamageToACreature,
		Key:       label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Object = DamagedCreatureRef(g, ev.Target)
			return item
		},
		Effect: ExileDamagedCreature,
	}
}

// OldB605DealtDamageByEquippedCreatureDied is Scythe of the Wretched's
// condition: a creature that the equipped creature dealt damage to this turn
// died. The equipped creature is the one attached when the creature died; it
// is read as the object it is now or last was (CR 400.7).
func OldB605DealtDamageByEquippedCreatureDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if _, died := DiedCreature(ev, g); !died {
		return false
	}
	host := source.AttachedTo
	if host.Kind != game.TargetCard || host.ID == uuid.Nil || host.ID == ev.CardID {
		return false
	}
	return DealtDamageByThisTurn(g, host.ID, ev.CardID)
}

// OldB605ReturnTheDeadAndAttach is Scythe of the Wretched's body: return that
// card to the battlefield under your control, then attach this Equipment to
// it. A card that has left the graveyard since (another effect, a replaced
// death) is not returned and nothing is attached; an Equipment that has left
// the battlefield attaches nothing (CR 701.3b, game.AttachForEffect).
func OldB605ReturnTheDeadAndAttach(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
		return err
	}
	if !OnBattlefield(g, id) {
		return nil
	}
	if info, ok := ctx.SourcePermanent(); !ok || info.Left {
		return nil
	}
	return g.AttachForEffect(item.SourceCardID, game.TargetRef{Kind: game.TargetCard, ID: id})
}
