package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Break Out — Sorcery {R}{G} (Murders at Karlov Manor, #4469):
//
//	"Look at the top six cards of your library. You may reveal a creature
//	 card from among them. If that card has mana value 2 or less, you may
//	 put it onto the battlefield and it gains haste until end of turn. If you
//	 didn't put the revealed card onto the battlefield this way, put it into
//	 your hand. Put the rest on the bottom of your library in a random
//	 order."
//
// The look is private; revealing the creature card tells the table (CR
// 701.20). The reveal is optional (a pick of none or one among the creature
// cards), the put is a yes/no only for a card with mana value 2 or less (read
// as it sits in the library), and a card that did not reach the battlefield
// (declined, too expensive, or replaced) goes to your hand if it is still in
// the library. The rest go to the bottom in a random order once that is done,
// so nothing is left on top while a question is open. Every leg is keyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "99eaac39-28d4-4af7-984b-f75fd37902ed",
		Name:         "Break Out",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			g := ctx.Game
			cards := g.LookAtTopOfLibraryForEffect(ctx.Controller(), 6)
			if len(cards) == 0 {
				return nil
			}
			var creatures []uuid.UUID
			for _, id := range cards {
				if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() {
					creatures = append(creatures, id)
				}
			}
			if len(creatures) == 0 {
				return g.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, cards)
			}
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  ctx.Controller(),
				Source:   ctx.Source(),
				Question: "Break Out — you may reveal a creature card from among them",
				Cards:    creatures,
				Min:      0,
				Max:      1,
				ThenKey:  breakOutRevealed,
				Carry:    cards,
			})
			return nil
		},
	})
}

// breakOutRevealed is the optional reveal answered. r.Carry holds the six.
var breakOutRevealed = game.RegisterChooseCardsThen("choose-cards/break-out-reveal", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return g.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, breakOutRest(g, r.Carry, uuid.Nil))
	}
	id := r.Picked[0]
	c, ok := g.LookupCardForEffect(id)
	if z := g.FindCardZoneForEffect(id); !ok || z == nil || z.Kind != game.ZoneLibrary {
		return g.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, breakOutRest(g, r.Carry, uuid.Nil))
	}
	g.RevealForEffect(game.RevealSpec{Player: r.Chooser, Source: r.Source, Cards: []uuid.UUID{id}, Reason: "Break Out — reveal a creature card"})
	carry := append([]uuid.UUID{id}, r.Carry...)
	if c.ManaValue() > 2 {
		return breakOutToHand(g, r.Chooser, r.Source, carry)
	}
	_, err := g.QueueConfirmKeyedForEffect(game.ConfirmPrompt{
		Chooser:      r.Chooser,
		Source:       r.Source,
		Question:     "Break Out — put the revealed card onto the battlefield? It gains haste until end of turn. Otherwise it goes to your hand.",
		AcceptLabel:  "Put it onto the battlefield",
		DeclineLabel: "Put it into my hand",
	}, breakOutPut, carry)
	return err
})

// breakOutPut is the yes/no: accept puts the card onto the battlefield, decline
// puts it into the hand. carry is the revealed card, then the six.
var breakOutPut = game.RegisterConfirmThen("confirm/break-out-put", func(g *game.Game, r game.ConfirmAnswer) error {
	if !r.Accepted || len(r.Carry) == 0 {
		return breakOutToHand(g, r.Chooser, r.Source, r.Carry)
	}
	return g.PutCardsFromLibraryOntoBattlefieldKeyedForEffect([]uuid.UUID{r.Carry[0]},
		game.LibraryEntryOptions{Controller: r.Chooser}, game.KeyedTail{Ref: breakOutEntered, Player: r.Chooser, Source: r.Source, Carry: r.Carry})
})

// breakOutEntered runs once the entry is done: the card, if it entered, gains
// haste until end of turn; if it did not enter and is still in the library it
// goes to hand; then the rest go to the bottom.
var breakOutEntered = game.RegisterTail("tail/break-out-entered", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	ctx := keyedContext(g, o.Tail.Player, o.Tail.Source)
	if len(o.IDs) > 0 {
		if err := (GrantKeywordUntilEOT{
			Target:   o.IDs[0],
			Keywords: []string{"haste"},
			Label:    "Break Out — haste until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
		return g.PutOnBottomInRandomOrderForEffect(o.Tail.Player, game.ZoneLibrary, breakOutRest(g, o.Tail.Carry[1:], o.Tail.Carry[0]))
	}
	return breakOutToHand(g, o.Tail.Player, o.Tail.Source, o.Tail.Carry)
})

// breakOutToHand puts the revealed card (carry[0]) into its owner's hand if it
// is still in the library, then the rest on the bottom in a random order.
func breakOutToHand(g *game.Game, player, _ uuid.UUID, carry []uuid.UUID) error {
	if len(carry) == 0 {
		return nil
	}
	id := carry[0]
	if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneLibrary {
		if err := g.BounceToHandForEffect(id); err != nil {
			return err
		}
	}
	return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, breakOutRest(g, carry[1:], id))
}

// breakOutRest is the cards of `six` that are still in a library and are not
// `except`: the rest that go to the bottom.
func breakOutRest(g *game.Game, six []uuid.UUID, except uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range six {
		if id == except {
			continue
		}
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneLibrary {
			out = append(out, id)
		}
	}
	return out
}
