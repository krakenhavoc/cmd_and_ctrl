package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lady Loki, Agent of Chaos — {5}{R} Legendary Creature — God Sorcerer
// Villain, 5/5 (Marvel Super Heroes Commander, #4455):
//
//		"Whenever you cast your first instant, sorcery, or Villain spell each
//		 turn, exile it, then exile cards from the top of your library until you
//		 exile a nonland card. Lady Loki deals damage to each opponent equal to
//		 the difference between that spell's mana value and that nonland card's
//		 mana value. You may cast that card without paying its mana cost."
//
//	  - "First ... each turn" is CastTally.InstantSorceryVillain, bumped as
//	    the spell becomes cast (before EventCast) and counting a spell once
//	    however many of the three it is; the trigger also requires the spell
//	    itself to be one of them, so a creature spell cast after the first
//	    qualifying one does not read the tally of the earlier spell.
//	  - "Exile it" is the spell on the stack, moved to exile and so never
//	    resolved (ExileSpellKeyedForEffect: not a counter, so an
//	    uncounterable spell is exiled all the same). The rest of the
//	    sentence goes on whether or not it worked (a spell already countered
//	    or resolved in response, a commander the owner sent to the command
//	    zone): the spell's mana value is read BEFORE the exile, with the X it
//	    was cast with (CR 202.3e), and carried on the tail.
//	  - The walk exiles every card down to the first nonland one and leaves
//	    the passed-over pile in exile (the cards are exiled, not milled). The
//	    damage is then dealt BETWEEN the walk and the free-cast offer: the
//	    tail told the cards exiled reads the nonland card's mana value (an X
//	    is 0 in exile, CR 107.3b), deals the absolute difference to each
//	    opponent from Lady Loki, and only then grants and offers the cast
//	    (exileUntilNonlandGrant, the offer Nico Minoru and Wanda's Vision
//	    make). A library that runs out before a nonland card has no card to
//	    compare, so no damage, and nothing to cast. Equal mana values deal 0.
//	  - The cast is optional ("may"), at no mana cost, as the trigger
//	    resolves (CR 608.2g); a card not cast stays in exile castable for the
//	    turn, like the other free-cast exile loops.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6efae65b-5ae5-4f62-97e7-cd4d1249fa9e",
		Name:         "Lady Loki, Agent of Chaos",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast,
				AllOf(YouCast(Or(Instant(), Sorcery(), Subtype("Villain"))), youCastYourFirstInstantSorceryOrVillainEachTurn),
				ladyLokiLabel,
				ladyLokiResolve),
		},
	})
}

// ladyLokiLabel is the trigger's stack label.
const ladyLokiLabel = "Lady Loki, Agent of Chaos — exile that spell, then exile cards until a nonland card; damage to each opponent equal to the difference in mana values; you may cast it free"

// ladyLokiName is the source name on the free-cast permission.
const ladyLokiName = "Lady Loki, Agent of Chaos"

// youCastYourFirstInstantSorceryOrVillainEachTurn is the "first ... each
// turn" half of the condition; YouCast supplies the type half. The tally
// is bumped before EventCast, so one means this spell is the first.
func youCastYourFirstInstantSorceryOrVillainEachTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor == source.Controller &&
		g.CastTallyFor(ev.Actor).InstantSorceryVillain == 1
}

// ladyLokiResolve reads the spell's mana value, then exiles the spell.
func ladyLokiResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if item.Trigger == nil {
		return nil
	}
	t := ctx.ItemTail(ladyLokiSpellExiled)
	t.N = triggeringSpellManaValue(g, item)
	return g.ExileSpellKeyedForEffect(item.Trigger.Event.CardID, t)
}

// ladyLokiSpellExiled is "then exile cards from the top of your library
// until you exile a nonland card", run whether or not the spell reached
// exile. The spell's mana value rides the tail as N.
var ladyLokiSpellExiled = CardTail("lady-loki/spell-exiled", func(ctx *Context, o game.TailOutcome) error {
	perm := exileUntilNonlandFreeCastPermission(ctx, ladyLokiName)
	t := ctx.ItemTail(ladyLokiWalked)
	t.N = o.Tail.N
	t.Permission = &perm
	return ctx.Game.MillToZoneKeyedForEffect(ctx.Controller(), 0, game.ZoneExile, UntilNonland, t)
})

// ladyLokiWalked is the hook between the walk and the free-cast offer:
// the damage, then the grant and the offer.
var ladyLokiWalked = CardTail("lady-loki/damage-then-cast", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.Permission == nil {
		return nil
	}
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsLand() {
			continue
		}
		mv, _ := ctx.Game.ManaValueForEffect(c)
		diff := o.Tail.N - mv
		if diff < 0 {
			diff = -diff
		}
		if diff > 0 {
			if err := (DamageEachOpponent{N: diff}).Apply(ctx); err != nil {
				return err
			}
		}
		break
	}
	return exileUntilNonlandGrant(ctx, *o.Tail.Permission, o.IDs)
})
