package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b144_helpers.go — Planar Atlas (Jumpstart 2022, #5948).

// gapB144AtlasLook is Planar Atlas's enters trigger resolving, after "you
// may": look at the top four cards of your library; if you do, reveal up to
// one land card from among them, put it on top and the rest on the bottom in
// a random order. The land is the card the player names (none is a legal
// answer); with no land among the four every card goes under at once.
func GapB144AtlasLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	looked := g.LookAtTopOfLibraryForEffect(player, 4)
	if len(looked) == 0 {
		return nil
	}
	var lands []uuid.UUID
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
			lands = append(lands, id)
		}
	}
	if len(lands) == 0 {
		return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Planar Atlas — you may reveal a land card from among them and put it on top of your library; the rest go on the bottom in a random order",
		Cards:    lands,
		Min:      0,
		Max:      1,
		ThenKey:  gapB144AtlasPlace,
		Carry:    looked,
	})
	return nil
}

// gapB144AtlasPlace reveals the named land and sends every other looked-at
// card to the bottom in a random order. The land is left where it is, the top
// of the library once the rest are under it. An on-disk key: never renamed,
// never reused.
var gapB144AtlasPlace = game.RegisterChooseCardsThen("choose-cards/s82-gap-b144-planar-atlas-place",
	func(g *game.Game, r game.CardsChosen) error {
		kept := map[uuid.UUID]bool{}
		if r.Answered {
			for _, id := range r.Picked {
				kept[id] = true
			}
		}
		if len(kept) > 0 {
			g.RevealForEffect(game.RevealSpec{Player: r.Chooser, Source: r.Source, Cards: r.Picked, Reason: "Planar Atlas — reveal a land card"})
		}
		var rest []uuid.UUID
		for _, id := range r.Carry {
			if !kept[id] {
				rest = append(rest, id)
			}
		}
		if len(rest) == 0 {
			return nil
		}
		return g.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, rest)
	})
