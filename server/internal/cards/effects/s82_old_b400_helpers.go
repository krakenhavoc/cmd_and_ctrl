package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB400NonbasicLandTappedForMana is Manabarbs' "a player taps a land for
// mana" (a mana ability whose source is a land that is now tapped) narrowed to
// a land without the basic supertype (CR 205.4c), read as it now stands.
func S82OldB400NonbasicLandTappedForMana(ev game.Event, g *game.Game) bool {
	if !B20LandTappedForMana(ev, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && !c.HasSupertype("basic")
}

// S82OldB400RevealRandomFromDamagedHand is "that player reveals a card at
// random from their hand" for a combat-damage trigger: the player the damage
// event named, the card picked from the game's keyed random stream (so an undo
// rewinds it) and revealed to the table. An empty hand reveals nothing.
func S82OldB400RevealRandomFromDamagedHand(name string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		p := g.PlayerByIDForEffect(item.Trigger.Event.Target)
		if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
			return nil
		}
		ids := make([]uuid.UUID, 0, p.Hand.Size())
		for _, c := range p.Hand.Cards {
			ids = append(ids, c.InstanceID)
		}
		pick := g.ChooseAtRandomForEffect(RandomDraw(NewContext(g, item)), ids, 1)
		if len(pick) == 0 {
			return nil
		}
		g.RevealForEffect(game.RevealSpec{Player: p.ID, Source: item.SourceCardID, Cards: pick, Reason: name})
		return nil
	}
}

// S82OldB400ArcanistImprint asks for an instant card in the controller's hand
// and exiles it linked to this object (CR 607.2a).
func S82OldB400ArcanistImprint(g *game.Game, item *game.StackItem) error {
	controller, source := item.Controller, item.SourceCardID
	if _, ok := g.SourceObjectForEffect(item); !ok {
		return nil
	}
	p := g.PlayerByIDForEffect(controller)
	if p == nil || p.Hand == nil {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.IsInstant() {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   source,
		Question: "Elite Arcanist — you may exile an instant card from your hand",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		ThenKey:  s82OldB400ArcanistImprintPick,
		LiveZone: game.ZoneHand,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// s82OldB400ArcanistImprintPick is the picked card exiled linked to the object
// that asked. An on-disk key: never renamed, never reused.
var s82OldB400ArcanistImprintPick = AskedPick("choose-cards/elite-arcanist-imprint", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	ref, ok := ctx.Game.SourceObjectForEffect(ctx.Item)
	if !ok {
		return nil
	}
	return ctx.Game.ExileLinkedForEffect(ref, r.Picked[0])
})

// CopyExiledWithAndCast is "copy the exiled card; you may cast the copy without
// paying its mana cost" for an imprint permanent (Isochron Scepter, Elite
// Arcanist): it copies the cards exiled with this object and offers the free
// cast of the copy; nothing happens when none is left in exile.
func CopyExiledWithAndCast(g *game.Game, item *game.StackItem) error {
	ref, ok := g.SourceObjectForEffect(item)
	if !ok {
		return nil
	}
	var cards []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		cards = append(cards, c.InstanceID)
	}
	if len(cards) == 0 {
		return nil
	}
	return CopyAndCast{Cards: cards, Free: true}.Apply(NewContext(g, item))
}

// S82OldB400MindseekerMill is Jace's Mindseeker's trigger as it resolves: the
// target opponent mills five, and the cast is offered over the cards that mill
// put into the graveyard.
func S82OldB400MindseekerMill(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opp := TargetedPlayerOf(ctx)
	if opp == uuid.Nil {
		return nil
	}
	return g.MillToZoneKeyedMovedForEffect(opp, 5, game.ZoneGraveyard, ctx.ItemTail(s82OldB400MindseekerCast))
}

// s82OldB400MindseekerCast is "you may cast an instant or sorcery spell from
// among them without paying its mana cost", after the mill has finished. An
// on-disk key: never renamed, never reused.
var s82OldB400MindseekerCast = CardTail("then/jaces-mindseeker-cast", func(ctx *Context, o game.TailOutcome) error {
	const name = "Jace's Mindseeker"
	g := ctx.Game
	var ids []uuid.UUID
	for _, id := range o.IDs {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneGraveyard {
			ids = append(ids, id)
		}
	}
	perm := FreeInstantOrSorcery().On(game.CastPermission{
		Player:      ctx.Controller(),
		Zone:        game.ZoneGraveyard,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		AltCostKey:  game.AltCostKeyGrantedCost,
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInHand,
		Source:      ctx.Source(),
		SourceName:  name,
		Label:       name + " — cast it without paying its mana cost",
	})
	candidates := FreeCastCandidates(g, &perm, ids)
	if len(candidates) == 0 {
		return nil
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      ctx.Controller(),
		Source:       ctx.Source(),
		Reason:       name,
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: DeclineLeaveInGraveyard,
	})
	return nil
})
