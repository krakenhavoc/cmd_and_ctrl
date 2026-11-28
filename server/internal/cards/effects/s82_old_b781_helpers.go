package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB781CruelFateLook is Cruel Fate: "Look at the top five cards of
// target opponent's library. Put one of those cards into that player's
// graveyard and the rest on top of their library in any order." Only the
// controller looks (CR 701.20) and must pick exactly one; the rest return in an
// order the controller chooses (CR 401.4). A library of one card has no choice.
func S82OldB781CruelFateLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	owner, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	looked := g.LookAtTopOfPlayersLibraryForEffect(ctx.Controller(), owner, 5)
	switch len(looked) {
	case 0:
		return nil
	case 1:
		return g.PutCardsIntoGraveyardKeyedForEffect(looked, game.KeyedTail{})
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: owner,
		Source:     ctx.Source(),
		Question:   "Cruel Fate — put one of those cards into that player's graveyard; the rest go back on top in any order",
		Cards:      looked,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneLibrary,
		ThenKey:    s82OldB781CruelFatePicked,
		Carry:      looked,
	})
	return nil
}

// s82OldB781CruelFatePicked puts the picked card into its owner's graveyard,
// then the rest of the look on top of that library in the controller's order.
// An on-disk key: never renamed, never reused.
var s82OldB781CruelFatePicked = game.RegisterChooseCardsThen("choose-cards/s82-old-b781-cruel-fate", func(g *game.Game, r game.CardsChosen) error {
	if ids := CardsStillInALibrary(g, r.Picked); len(ids) > 0 {
		if err := g.PutCardsIntoGraveyardKeyedForEffect(ids, game.KeyedTail{}); err != nil {
			return err
		}
	}
	var rest []uuid.UUID
	for _, id := range r.Carry {
		if len(r.Picked) > 0 && id == r.Picked[0] {
			continue
		}
		rest = append(rest, id)
	}
	rest = CardsStillInALibrary(g, rest)
	if len(rest) == 0 {
		return nil
	}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   r.Chooser,
		Source:    r.Source,
		Cards:     rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceTop,
		Reason:    "Cruel Fate — put the rest on top of that library in any order",
	})
})
