package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Discover the Impossible — Instant {2}{U} (Kamigawa: Neon Dynasty, S82
// #5493):
//
//	"Look at the top five cards of your library. Exile one of them face
//	 down and put the rest on the bottom of your library in a random order.
//	 You may cast the exiled card without paying its mana cost if it's an
//	 instant spell with mana value 2 or less. If you don't, put that card
//	 into your hand."
//
// Gonti's look-and-exile (a choose_cards pick of one, the rest on the
// bottom at random, the pick exiled face down), then Beseech the Mirror's
// "cast it free or put it into your hand", with the kind of spell named:
// Game.CastExiledFreeOrToHandFilteredForEffect stamps a free, flash-timed
// cast-only grant capped at mana value 2 over instants only, and asks the
// cast during this resolution (CR 608.2g). A card that is not an instant
// with mana value 2 or less (a land, a sorcery, a bigger instant) goes
// straight to the hand; so does one that is declined.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c2a3f13-6bcb-42ad-98f6-642151ad54f1",
		Name:         "Discover the Impossible",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			g, controller, source := ctx.Game, ctx.Controller(), ctx.Source()
			looked := g.LookAtTopOfLibraryForEffect(controller, 5)
			if len(looked) == 0 {
				return nil
			}
			if len(looked) == 1 {
				return neoB15DiscoverFinish(g, controller, source, looked, looked)
			}
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  controller,
				Source:   source,
				Question: "Discover the Impossible — exile one of them face down",
				Cards:    looked,
				Min:      1,
				Max:      1,
				LiveZone: game.ZoneLibrary,
				ThenKey:  neoB15DiscoverPicked,
				Carry:    looked,
			})
			return nil
		},
	})
}

// neoB15DiscoverPicked is the pick; the carry is the five looked at. An
// on-disk key: never renamed, never reused.
var neoB15DiscoverPicked = game.RegisterChooseCardsThen("choose-cards/neo-discover-the-impossible-exile", func(g *game.Game, r game.CardsChosen) error {
	return neoB15DiscoverFinish(g, r.Chooser, r.Source, r.Carry, r.Picked)
})

// neoB15DiscoverCast is the exiled card's free cast or hand. An on-disk
// key: never renamed, never reused.
var neoB15DiscoverCast = game.RegisterTail("neo-b15/discover-the-impossible-cast", func(g *game.Game, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := g.CastExiledFreeOrToHandFilteredForEffect(o.Tail.Player, o.Tail.Source, id, 2,
			game.PermissionFilter{CardType: "instant"}); err != nil {
			return err
		}
	}
	return nil
})

func neoB15DiscoverFinish(g *game.Game, controller, source uuid.UUID, looked, picked []uuid.UUID) error {
	took := make(map[uuid.UUID]bool, len(picked))
	for _, id := range picked {
		took[id] = true
	}
	var rest []uuid.UUID
	for _, id := range cardsStillInALibrary(g, looked) {
		if !took[id] {
			rest = append(rest, id)
		}
	}
	firstErr := g.PutOnBottomInRandomOrderForEffect(controller, game.ZoneLibrary, rest)
	if len(picked) > 0 {
		then := game.KeyedTail{Ref: neoB15DiscoverCast, Player: controller, Source: source}
		if err := g.ExileCardsFaceDownChosenKeyedForEffect(controller, picked, then); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
