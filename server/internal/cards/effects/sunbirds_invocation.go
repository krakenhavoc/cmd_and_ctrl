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
// cast from exile, a graveyard or the command zone does not fire it.
// That is also what ends the chain: the free spell this ability lets you
// cast is cast from outside your hand and cannot trigger it again.
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
//     window closes on the caster's next pass, sending an uncast card to
//     the bottom of the library.
//   - The rest goes to the bottom in a random order from the game's
//     keyed stream, so an undone or restored game bottoms them the same
//     way.
//
// The one place it differs from the printed card: the chosen card is
// exiled from the library to be cast, where the printed card casts it
// from the library. The grant machinery casts from exile, so a "cast
// from your library" watcher would not see this cast.
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
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The card you choose to cast is exiled first and cast from exile, so it isn't cast from your library.",
		},
		Triggered: []game.TriggeredAbility{t},
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
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   source,
		Question: "Sunbird's Invocation — cast a spell with mana value " + strconv.Itoa(x) + " or less without paying its mana cost?",
		Cards:    eligible,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneLibrary,
		Then:     finish,
	})
	return nil
}

// sunbirdsFinisher is what happens once the pick is known: the chosen
// card, if any, is exiled and given its free cast only if it really
// reached exile (a CR 614 replacement or a commander's CR 903.9 offer
// can send it elsewhere), and the rest of the revealed cards go to the
// bottom in a random order. Everything it captures is an ID or a
// scalar, so an undo across the prompt resolves against the restored
// game.
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
		bottomRest := func(g *game.Game) error {
			return g.PutOnBottomInRandomOrderForEffect(controller, game.ZoneLibrary, rest)
		}
		if chosen == uuid.Nil {
			return bottomRest(g)
		}
		return g.ExileCardThenForEffect(chosen, func(g *game.Game, exiled bool) error {
			if exiled {
				sunbirdsGrantFreeCast(g, controller, source, chosen, x)
			}
			return bottomRest(g)
		})
	}
}

// sunbirdsGrantFreeCast stamps the free cast on the exiled card, the
// shape cascade's grant has, with Sunbird's own mana value cap.
func sunbirdsGrantFreeCast(g *game.Game, controller, source, cardID uuid.UUID, x int) {
	z := g.FindCardZoneForEffect(cardID)
	c, ok := g.LookupCardForEffect(cardID)
	if !ok || z == nil || z.Kind != game.ZoneExile {
		return
	}
	limit := x
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:            controller,
		Zone:              game.ZoneExile,
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
