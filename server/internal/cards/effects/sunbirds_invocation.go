package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sunbird's Invocation — Enchantment {5}{R}:
//
//	"Whenever you cast a spell from your hand, reveal the top X cards
//	 of your library, where X is that spell's mana value. You may cast
//	 a spell with mana value X or less from among cards revealed this
//	 way without paying its mana cost. Put the rest on the bottom of
//	 your library in a random order."
//
// The trigger is a cast trigger narrowed to a spell cast from HAND: the
// cast event records where the spell came from (CR 601.2a), so a spell
// cast from exile, a graveyard, a library or the command zone does not
// fire it. That is also what ends the chain: the free spell this
// ability lets you cast is cast from your library and cannot trigger it
// again.
//
// X is that spell's mana value, read as the trigger goes on the stack
// (CR 202.3e counts an X in the cost while the spell is on the stack)
// and carried on the item. A spell with X = 0 reveals nothing.
//
// The rest is cascade's machinery, because it is the same instruction:
//
//   - "You may cast a spell" is a choose-up-to-one over the revealed
//     nonland cards whose mana value is X or less (a land is not a
//     spell). The pick is hidden from nobody, since the cards were
//     revealed.
//   - The chosen card gets a free-cast GRANT, not an inline cast (ADR
//     0066's posture on every "you may cast it" a resolution offers):
//     {0}, flash timing so a sorcery can be cast at once (CR 608.2g),
//     mana value capped at X against the face actually cast, and the
//     cast is offered during the resolution (#4250); the window is only the
//     fallback and closes on the caster's next pass, sending an uncast card
//     to the bottom of the library.
//   - The rest goes to the bottom in a random order from the game's
//     keyed stream, so an undone or restored game bottoms them the same
//     way.
//
// The chosen card is cast from the LIBRARY, as printed, rather than
// exiled first the way a cascade hit is. It stays where it was
// revealed, and once the rest are bottomed it is the top card, still
// known to every seat from the reveal, so its owner's pile shows it and
// the view stamps it castable. The grant names that object (ScopeCards
// in ZoneLibrary) rather than the top position, so it opens that card
// and no other, and the cast reaches the stack from the library, where
// a "cast from your library" reader looks for it.
//
// No simplification beyond the grant posture every resolution-time
// "you may cast" takes.
func init() {
	t := On(game.EventCast, sunbirdsCastFromHand,
		"Sunbird's Invocation — reveal the top X cards and cast one free", sunbirdsReveal)
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, "Sunbird's Invocation — reveal the top X cards and cast one free")
		item.Params.Amount = spellManaValueForEffect(g, ev.CardID)
		return item
	}
	Register(Spec{
		OracleID:     "10d482fd-e034-4187-b148-51fe15885360",
		Name:         "Sunbird's Invocation",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

// sunbirdsCastFromHand is "whenever you cast a spell from your hand".
func sunbirdsCastFromHand(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor == source.Controller && ev.OldZone == game.ZoneHand
}

// sunbirdsReveal reveals the top X and offers the free cast.
func sunbirdsReveal(g *game.Game, item *game.StackItem) error {
	x := item.Params.Amount
	controller, source := item.Controller, item.SourceCardID
	revealed := g.RevealTopOfLibraryForEffect(controller, source, x, "Sunbird's Invocation — the top "+strconv.Itoa(x)+" cards")
	if len(revealed) == 0 {
		return nil
	}
	var eligible []uuid.UUID
	for _, id := range revealed {
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.IsLand() {
			continue
		}
		if mv, readable := g.ManaValueForEffect(c); readable && mv <= x {
			eligible = append(eligible, id)
		}
	}
	finish := sunbirdsFinisher(controller, source, x, revealed)
	if len(eligible) == 0 {
		return finish(g, nil)
	}
	// #4250: the cast is made during the trigger's resolution (CR
	// 608.2g), through the resolution_cast prompt over the eligible cards
	// where they sit in the library: ONE free, capped permission, spent by
	// the one cast. Once the prompt is answered, the rest of the revealed
	// cards still in the library go to the bottom in a random order
	// (lookedAtRestToBottom). The prompt is plain data, so a table waiting
	// on it is a restore point. The pick below is the offer outside an open
	// resolution, which the trigger never reaches.
	limit := x
	perm := game.CastPermission{
		Player:            controller,
		Zone:              game.ZoneLibrary,
		Duration:          g.UntilEndOfTurnDuration(),
		Cost:              "{0}",
		Timing:            game.TimingFlash,
		CastOnly:          true,
		MaxSpellManaValue: &limit,
		// Nothing moves on the decline: the continuation bottoms every
		// revealed card left, in one random order.
		LapseOnPass: game.LapseStaysInHand,
		Source:      source,
		SourceName:  "Sunbird's Invocation",
		Label:       "Sunbird's Invocation — cast it without paying its mana cost",
	}
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       "Sunbird's Invocation",
		Cards:        eligible,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put them on the bottom",
		Then:         lookedAtRestToBottom,
		Carry:        revealed,
	}) {
		return nil
	}
	// Keyed (#4253): the revealed cards and X ride the pick as data.
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:   controller,
		Source:    source,
		Question:  "Sunbird's Invocation — cast a spell with mana value " + strconv.Itoa(x) + " or less without paying its mana cost?",
		Cards:     eligible,
		Min:       0,
		Max:       1,
		LiveZone:  game.ZoneLibrary,
		ThenKey:   sunbirdsPicked,
		Carry:     revealed,
		CarryNums: []int{x},
	})
	return nil
}

// sunbirdsPicked is the pick outside an open resolution answered: the carry
// is the revealed cards and the number is X. An on-disk key: never renamed,
// never reused.
var sunbirdsPicked = game.RegisterChooseCardsThen("choose-cards/sunbirds-invocation-cast", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Nums) == 0 {
		return nil
	}
	return sunbirdsFinisher(r.Chooser, r.Source, r.Nums[0], r.Carry)(g, r.Picked)
})

// sunbirdsFinisher is what happens once the pick is known: the rest of
// the revealed cards go to the bottom in a random order, and the chosen
// card, if any, keeps its place (now the top of the library) and gets
// its free cast there. Everything it captures is an ID or a scalar, and it
// runs at once (the keyed pick, sunbirdsPicked, rebuilds it from data).
func sunbirdsFinisher(controller, source uuid.UUID, x int, revealed []uuid.UUID) func(g *game.Game, picked []uuid.UUID) error {
	return func(g *game.Game, picked []uuid.UUID) error {
		var chosen uuid.UUID
		if len(picked) > 0 {
			chosen = picked[0]
		}
		var rest []uuid.UUID
		for _, id := range cardsStillInALibrary(g, revealed) {
			if id != chosen {
				rest = append(rest, id)
			}
		}
		if err := g.PutOnBottomInRandomOrderForEffect(controller, game.ZoneLibrary, rest); err != nil {
			return err
		}
		if chosen != uuid.Nil {
			sunbirdsGrantFreeCast(g, controller, source, chosen, x)
		}
		return nil
	}
}

// sunbirdsGrantFreeCast stamps the free cast on the chosen card where it
// sits in its owner's library: cascade's grant shape with Sunbird's own
// mana value cap, pointed at the library. A card that has left the
// library since the reveal gets nothing. The holder's next pass bottoms an
// uncast card only as the fallback, for an offer made outside the
// resolution (lapseCastPermissionLocked reads the zone the grant opens).
func sunbirdsGrantFreeCast(g *game.Game, controller, source, cardID uuid.UUID, x int) {
	z := g.FindCardZoneForEffect(cardID)
	c, ok := g.LookupCardForEffect(cardID)
	if !ok || z == nil || z.Kind != game.ZoneLibrary {
		return
	}
	limit := x
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:            controller,
		Zone:              game.ZoneLibrary,
		Duration:          g.UntilEndOfTurnDuration(),
		Cost:              "{0}",
		Timing:            game.TimingFlash,
		CastOnly:          true,
		MaxSpellManaValue: &limit,
		LapseOnPass:       game.LapseToLibraryBottom,
		Source:            source,
		SourceName:        "Sunbird's Invocation",
		Label:             "Sunbird's Invocation — cast it without paying its mana cost",
	}, []game.Card{c})
}
