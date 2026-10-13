package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_graveyard_search_bot_test.go — #3217. A bot at a "search your
// library and/or graveyard" prompt is enumerated both zones' cards and
// is not offered fail to find while a public match is on the menu, so a
// policy that prefers it falls through to a take. Not behind
// AISEAT_GAME_TESTS: one prompt, milliseconds.

func stockGraveyard(me *game.Player, c game.Card) uuid.UUID {
	c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
	me.Graveyard.PushTop(c)
	return c.InstanceID
}

func TestHeuristicSeatAnswersALibraryAndGraveyardSearch(t *testing.T) {
	g := newSettledTable(t, 3217)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := stockSearchLibrary(me, game.Card{Name: "Command Tower", TypeLine: "Land"})
	gy := stockGraveyard(me, game.Card{Name: "Command Tower", TypeLine: "Land"})
	searchForBot(t, g, game.SearchLibrarySpec{
		Player: me.ID, Pred: func(c game.Card) bool { return c.Name == "Command Tower" },
		Dest: game.ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
	})
	if !owesChoice(g, me.ID) {
		t.Fatal("two matches in two zones did not prompt")
	}

	driveChoices(t, g, heuristic.New(), me.ID, 4)

	if owesChoice(g, me.ID) {
		t.Fatal("the seat still owes the search")
	}
	if !me.Hand.Contains(lib[0]) && !me.Hand.Contains(gy) {
		t.Error("the heuristic seat found nothing although a public match was on offer")
	}
}

func TestBotCannotFailToFindWhileAGraveyardMatchIsOnOffer(t *testing.T) {
	g := newSettledTable(t, 3218)
	me := g.Seats[g.Turn.ActiveSeat]
	stockSearchLibrary(me, game.Card{Name: "Command Tower", TypeLine: "Land"})
	stockGraveyard(me, game.Card{Name: "Command Tower", TypeLine: "Land"})
	searchForBot(t, g, game.SearchLibrarySpec{
		Player: me.ID, Pred: func(c game.Card) bool { return c.Name == "Command Tower" },
		Dest: game.ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
	})

	pol := &failToFindPolicy{}
	driveChoices(t, g, pol, me.ID, 4)

	if pol.failed != 0 {
		t.Error("fail to find was on the menu although the graveyard holds a match")
	}
	if owesChoice(g, me.ID) {
		t.Error("the seat still owes the search")
	}
}
