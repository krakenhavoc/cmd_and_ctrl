package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glorious Purpose — Enchantment — Plan {1}{U} (Marvel Super Heroes
// Commander, #4455):
//
//	"Whenever a creature you control connives, put a +1/+1 counter on that
//	 creature and a plan counter on this enchantment.
//	 When the sixth plan counter is put on this enchantment, sacrifice it.
//	 If you do, exile the top four cards of your library. You may cast any
//	 number of spells from among them without paying their mana costs. Put
//	 the rest into your hand."
//
// Two printed abilities, two triggers. The first watches EventConnived
// (CR 701.50f) like Iron Monger: the counter goes on the creature that
// connived, read off the event (a creature that left gets nothing), then the
// plan counter on the enchantment. The second is a placement that CROSSES
// six from below (Claim the Kingdom's shape). "If you do" is the enchantment
// really having left the battlefield: a sacrifice a replacement stopped
// exiles nothing.
//
// The four cards are exiled as one run (a short library exiles what it has),
// each nonland card gets its own free, cast-only permission (Etali's
// "any number"), and the casts are made during the resolution (CR 608.2g),
// each its own answer to the resolution_cast prompt until the player
// declines or none is left. "The rest" is whatever of the four is still in
// exile once that prompt is answered (the prompt's continuation): the lands,
// and every spell not cast, go into the hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "182baad0-ec6e-4e93-9dfc-7c80b37005de",
		Name:         "Glorious Purpose",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverACreatureYouControlConnives("Glorious Purpose — put a +1/+1 counter on that creature and a plan counter on this enchantment",
				gloriousPurposeCounters),
			On(game.EventCounterPlaced, mshB2PlanCounterReached(6),
				"Glorious Purpose — sacrifice it; if you do, exile the top four cards of your library and cast any number free",
				gloriousPurposeSixth),
		},
	})
}

// gloriousPurposeCounters is the connive trigger's body: the +1/+1 counter
// on the creature that connived, then the plan counter on this.
func gloriousPurposeCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if creature := item.Trigger.Event.CardID; onBattlefield(g, creature) {
		if err := (AddCounter{Target: creature, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if !onBattlefield(g, item.SourceCardID) || ctx.isNewSourceObjectAsThis(item.SourceCardID) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: "plan", N: 1}.Apply(ctx)
}

// gloriousPurposeSixth sacrifices the enchantment and, if it left, exiles
// the top four cards.
func gloriousPurposeSixth(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !onBattlefield(g, item.SourceCardID) || ctx.isNewSourceObjectAsThis(item.SourceCardID) {
		return nil
	}
	if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	if onBattlefield(g, item.SourceCardID) {
		return nil // "If you do": it was not sacrificed
	}
	return MillToZone{
		Player:   item.Controller,
		N:        4,
		To:       game.ZoneExile,
		ThenTail: ctx.ItemTail(gloriousPurposeExiled),
	}.Apply(ctx)
}

// gloriousPurposeExiled grants the free casts over the spells among the
// exiled cards and offers them, the rest going to hand afterwards.
var gloriousPurposeExiled = CardTail("glorious-purpose/after-exile", func(ctx *Context, o game.TailOutcome) error {
	g, controller, source := ctx.Game, ctx.Controller(), ctx.Source()
	var exiled, castable []uuid.UUID
	for _, id := range o.IDs {
		z := g.FindCardZoneForEffect(id)
		c, ok := g.LookupCardForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		exiled = append(exiled, id)
		if c.IsLand() {
			continue // a land is played, never cast
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:      controller,
			Zone:        game.ZoneExile,
			Duration:    g.UntilEndOfTurnDuration(),
			Cost:        "{0}",
			Timing:      game.TimingFlash,
			CastOnly:    true,
			LapseOnPass: game.LapseStaysInExile,
			Source:      source,
			SourceName:  "Glorious Purpose",
			Label:       "Glorious Purpose — cast it without paying its mana cost",
		}, []game.Card{c})
		castable = append(castable, id)
	}
	if len(exiled) == 0 {
		return nil
	}
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       "Glorious Purpose",
		Cards:        castable,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put it into your hand",
		Then:         gloriousPurposeRestToHand,
		Carry:        exiled,
	}) {
		return nil
	}
	return gloriousPurposePutRest(g, controller, source, exiled)
})

// gloriousPurposeRestToHand is "Put the rest into your hand", once the
// resolution_cast prompt is answered: Carry is the four exiled cards. An
// on-disk key, never renamed.
var gloriousPurposeRestToHand = game.RegisterChooseCardsThen("resolution-cast/glorious-purpose-rest-to-hand",
	func(g *game.Game, r game.CardsChosen) error {
		return gloriousPurposePutRest(g, r.Chooser, r.Source, r.Carry)
	})

// gloriousPurposePutRest puts every one of the cards still in exile into
// its controller's hand (the ones cast are on the stack, not in exile).
func gloriousPurposePutRest(g *game.Game, controller, source uuid.UUID, ids []uuid.UUID) error {
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if _, err := g.ReturnFromAnyZoneToHandForEffect(controller, source, id); err != nil {
			return err
		}
	}
	return nil
}
