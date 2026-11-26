package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b311_helpers.go — Battle for Zendikar lands (#5547).

// S82OldB311DoesntUntapTheTarget is "target creature an opponent controls
// doesn't untap during its controller's next untap step" as a trigger body:
// the first target still legal as it resolves (CR 608.2b).
func S82OldB311DoesntUntapTheTarget(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return DoesntUntapNextUntapStep{Targets: []uuid.UUID{id}, Label: label}.Apply(ctx)
	}
}

// S82OldB311LookAtFiveTopBasicLand is Fertile Thicket's whole clause: "look
// at the top five cards of your library. If you do, reveal up to one basic
// land card from among them, then put that card on top of your library and
// the rest on the bottom in any order." The card kept on top stays where it
// is while the rest go to the bottom, which leaves it on top. An on-disk key.
func S82OldB311LookAtFiveTopBasicLand(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		player := item.Controller
		looked := g.LookAtTopOfLibraryForEffect(player, 5)
		if len(looked) == 0 {
			return nil
		}
		var lands []uuid.UUID
		for _, id := range looked {
			if c, ok := g.LookupCardForEffect(id); ok && IsBasicLand(c) {
				lands = append(lands, id)
			}
		}
		if len(lands) == 0 {
			return s82OldB311RestOnBottom(g, item.SourceCardID, player, looked, nil)
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  player,
			Source:   item.SourceCardID,
			Question: label,
			Cards:    lands,
			Min:      0,
			Max:      1,
			ThenKey:  s82OldB311BasicLandPicked,
			Carry:    looked,
			LiveZone: game.ZoneLibrary,
		})
		return nil
	}
}

var s82OldB311BasicLandPicked = game.RegisterChooseCardsThen("choose-cards/s82-old-b311-fertile-thicket",
	func(g *game.Game, r game.CardsChosen) error {
		var picked []uuid.UUID
		if r.Answered {
			picked = r.Picked
		}
		if len(picked) > 0 {
			g.RevealForEffect(game.RevealSpec{
				Player: r.Chooser,
				Source: r.Source,
				Reason: "Fertile Thicket — reveal up to one basic land card",
				Cards:  CardsStillInALibrary(g, picked),
			})
		}
		return s82OldB311RestOnBottom(g, r.Source, r.Chooser, r.Carry, picked)
	})

// s82OldB311RestOnBottom puts the looked-at cards that are still in the
// library, but for `keep`, on the bottom in the order the player chooses.
func s82OldB311RestOnBottom(g *game.Game, source, player uuid.UUID, looked, keep []uuid.UUID) error {
	kept := make(map[uuid.UUID]bool, len(keep))
	for _, id := range keep {
		kept[id] = true
	}
	var rest []uuid.UUID
	for _, id := range CardsStillInALibrary(g, looked) {
		if !kept[id] {
			rest = append(rest, id)
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   player,
		Source:    source,
		Cards:     rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Reason:    LibraryOrderLabel(g, source, "put the rest on the bottom of your library in any order"),
	})
}
