package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sage of Days — Creature — Human Wizard {2}{U}, 3/2 (The Lost Caverns of
// Ixalan, #4470):
//
//	"When this creature enters, look at the top three cards of your
//	 library. You may put one of those cards back on top of your library.
//	 Put the rest into your graveyard."
//
// A look (only the controller sees the three) answered by one
// choose_cards prompt with a floor of zero: the picked card never moves,
// so it stays on top of the library, and every other looked card goes into
// the graveyard from the top (a keyed continuation, so a table waiting on
// the pick is a restore point). Declining puts all three into the
// graveyard. Fewer than three cards in the library look at what is there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "656fc672-efa5-484a-b5e8-eac262331439",
		Name:         "Sage of Days",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sage of Days — look at the top three cards of your library", sageOfDaysLook),
		},
	})
}

func sageOfDaysLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	looked := g.LookAtTopOfLibraryForEffect(player, 3)
	if len(looked) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Sage of Days — you may put one of those cards back on top of your library; put the rest into your graveyard",
		Cards:    looked,
		Min:      0,
		Max:      1,
		ThenKey:  sageOfDaysPick,
		LiveZone: game.ZoneLibrary,
		Carry:    looked,
	})
	return nil
}

// sageOfDaysPick leaves the picked card where it is (on top of the
// library) and puts every other looked card into its owner's graveyard.
var sageOfDaysPick = game.RegisterChooseCardsThen("choose-cards/sage-of-days-rest-to-graveyard", func(g *game.Game, r game.CardsChosen) error {
	var rest []uuid.UUID
	for _, id := range r.Carry {
		if !slices.Contains(r.Picked, id) {
			rest = append(rest, id)
		}
	}
	return restIntoGraveyard(g, rest)
})
