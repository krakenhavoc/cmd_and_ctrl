package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hide in Plain Sight — Sorcery {3}{G} (Murders at Karlov Manor, #4469):
//
//	"Look at the top five cards of your library, cloak two of them, and put
//	 the rest on the bottom of your library in a random order. (To cloak a
//	 card, put it onto the battlefield face down as a 2/2 creature with ward
//	 {2}. Turn it face up any time for its mana cost if it's a creature
//	 card.)"
//
// The look is private to you (CR 701.20), the two are chosen in one pick
// (exactly two, or every card when the library held fewer) and any card may
// be cloaked, a nonpermanent one included (CR 701.58a). They enter together,
// face down under your control (game.FaceDownCloaked, the cloak seam,
// cloak.go), and the rest go to the bottom in a random order once the entry
// is done, so a cloaked card's own question cannot leave the others on top.
// Both steps are keyed, so a table waiting on the pick is a restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94b15f1d-827b-45e9-a684-5d0c52ea0a09",
		Name:         "Hide in Plain Sight",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			g := ctx.Game
			cards := g.LookAtTopOfLibraryForEffect(ctx.Controller(), 5)
			if len(cards) == 0 {
				return nil
			}
			n := min(2, len(cards))
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  ctx.Controller(),
				Source:   ctx.Source(),
				Question: "Hide in Plain Sight — choose two of these cards to cloak; the rest go on the bottom of your library in a random order",
				Cards:    cards,
				Min:      n,
				Max:      n,
				Zone:     game.ZoneLibrary,
				ThenKey:  hideInPlainSightCloak,
			})
			return nil
		},
	})
}

// hideInPlainSightCloak is the pick answered: the chosen cards are cloaked
// together, and the rest are put on the bottom when the entry is complete.
var hideInPlainSightCloak = game.RegisterChooseCardsThen("choose-cards/hide-in-plain-sight-cloak", func(g *game.Game, r game.CardsChosen) error {
	var picked, rest []uuid.UUID
	for _, id := range r.Candidates {
		z := g.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneLibrary {
			continue
		}
		if slices.Contains(r.Picked, id) {
			picked = append(picked, id)
		} else {
			rest = append(rest, id)
		}
	}
	then := game.KeyedTail{Ref: hideInPlainSightRest, Player: r.Chooser, Source: r.Source, Carry: rest}
	if len(picked) == 0 {
		return then.Run(g, nil)
	}
	return g.PutCardsFromLibraryOntoBattlefieldKeyedForEffect(picked,
		game.LibraryEntryOptions{Controller: r.Chooser, FaceDown: game.FaceDownCloaked}, then)
})

// hideInPlainSightRest puts the cards that were not cloaked on the bottom of
// their owner's library in a random order. A card that is no longer in a
// library (it was cloaked and left, or was moved) is skipped.
var hideInPlainSightRest = game.RegisterTail("tail/hide-in-plain-sight-rest", func(g *game.Game, o game.TailOutcome) error {
	var ids []uuid.UUID
	for _, id := range o.Tail.Carry {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneLibrary {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.PutOnBottomInRandomOrderForEffect(o.Tail.Player, game.ZoneLibrary, ids)
})
