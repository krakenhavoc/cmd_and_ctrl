package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silhana Wayfinder — Creature — Elf Scout {1}{G}, 2/1 (Jumpstart, #5510):
//
//	"When this creature enters, look at the top four cards of your library.
//	 You may reveal a creature or land card from among them and put it on top
//	 of your library. Put the rest on the bottom of your library in a random
//	 order."
//
// Planar Atlas's look with a creature-or-land filter: the named card is
// revealed and left where it is, the top of the library once the rest are
// under it. Naming nothing is a legal answer, and with no creature or land
// among the four every card goes under at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc858fc1-d97a-4cc5-af31-2c33d9639e1b",
		Name:         "Silhana Wayfinder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Silhana Wayfinder — look at the top four cards of your library", silhanaWayfinderLook),
		},
	})
}

func silhanaWayfinderLook(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Controller()
	looked := g.LookAtTopOfLibraryForEffect(player, 4)
	if len(looked) == 0 {
		return nil
	}
	var fits []uuid.UUID
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok && (c.IsCreature() || c.IsLand()) {
			fits = append(fits, id)
		}
	}
	if len(fits) == 0 {
		return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Silhana Wayfinder — you may reveal a creature or land card from among them and put it on top of your library; the rest go on the bottom in a random order",
		Cards:    fits,
		Min:      0,
		Max:      1,
		ThenKey:  silhanaWayfinderPlace,
		Carry:    looked,
	})
	return nil
}

// silhanaWayfinderPlace reveals the named card and sends every other looked-at
// card to the bottom in a random order. An on-disk key: never renamed, never
// reused.
var silhanaWayfinderPlace = game.RegisterChooseCardsThen("choose-cards/s82-old-b103-silhana-wayfinder-place",
	func(g *game.Game, r game.CardsChosen) error {
		kept := map[uuid.UUID]bool{}
		if r.Answered {
			for _, id := range r.Picked {
				kept[id] = true
			}
		}
		if len(kept) > 0 {
			g.RevealForEffect(game.RevealSpec{Player: r.Chooser, Source: r.Source, Cards: r.Picked, Reason: "Silhana Wayfinder — reveal a creature or land card"})
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
