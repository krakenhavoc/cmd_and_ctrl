package aiseat_test

import (
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// disturb_game_test.go — ADR 0107 §4 (#1855), the whole-game half: a
// heuristic bot with a disturb card in its graveyard, on the
// Forests-only fixture flashback_game_test.go explains, finds the
// disturb cast, claims it, and the back face resolves onto the
// battlefield — no rejected move and no stall on the way.

const oracleBaithookAnglerBot = "c6bb4b41-8dae-429a-b928-ae9d39c74711"

func TestBotDisturbsOutOfItsGraveyardInARun(t *testing.T) {
	requireGameTests(t)

	room := newGraveyardRoom(t, 18550)
	g := room.Game
	bot := g.Seats[0]

	angler := game.Card{
		OracleID: oracleBaithookAnglerBot, Layout: game.LayoutTransform,
		Faces: []game.Face{
			{Name: "Baithook Angler", TypeLine: "Creature — Human Peasant", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 2, Toughness: 1},
			{Name: "Hook-Haunt Drifter", TypeLine: "Creature — Spirit", Colors: []string{"U"}, Power: 1, Toughness: 2},
		},
	}
	angler.SetFace(0)
	id := seedGraveyard(g, bot, angler)
	for i := 0; i < 2; i++ {
		seedBattlefield(g, bot, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	}

	watch := newCastWatcher(heuristic.New())
	res := playGameIn(t, room, 18550, []aiseat.Policy{watch, passPolicy{}}, 6, 120*time.Second)

	if watch.offeredFor(id) == 0 {
		t.Fatalf("Baithook Angler was never offered out of the graveyard in %d turns", res.turns)
	}
	if paid := watch.paidKeys(); paid["disturb"] == 0 {
		t.Errorf("the bot never disturbed anything; prices claimed: %v", paid)
	}
	var drifter *game.Card
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				c := g.Battlefield.Cards[i]
				drifter = &c
			}
		}
	})
	if drifter == nil || drifter.ActiveFace != 1 || drifter.Name != "Hook-Haunt Drifter" {
		t.Errorf("no Hook-Haunt Drifter on the board after %d turns (got %+v)", res.turns, drifter)
	}
	assertNoEnumeratorBugs(t, res)
}
