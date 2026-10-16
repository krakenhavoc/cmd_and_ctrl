package aiseat_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// controlled_search_test.go — #3998: a bot that controls an opponent's
// search (Opposition Agent) answers it in its own interest, through
// controlledSearchPolicy rather than passively and rather than through
// its own policy. Played lockstep.

const (
	csAgentOracle  = "1f438b8f-fe23-4f3b-ab2e-f6c33676c462"
	csDemonicTutor = "82004860-e589-4e38-8d61-8c0210e4ea39"
	csExplosiveVeg = "a0abd957-01a0-4aa9-8fc8-0e6840d21606"
)

// csTable is a three-seat table with an Opposition Agent under Bob (the
// bot) and Alice, the active player, searching with `spell`, whose
// library is exactly `library`. Returns once the search prompt is open.
func csTable(t *testing.T, spell, oracle string, library ...game.Card) (room *ws.Room, alice, bob uuid.UUID) {
	t.Helper()
	room = newRoom(t, 3, 3998)
	g := room.Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	a := g.Seats[g.Turn.ActiveSeat]
	b := g.Seats[(g.Turn.ActiveSeat+1)%3]
	alice, bob = a.ID, b.ID
	advanceToStep(t, g, game.StepPrecombatMain)
	seedBattlefield(g, b, game.Card{Name: "Opposition Agent", OracleID: csAgentOracle, TypeLine: "Creature — Human Rogue", Power: 3, Toughness: 2})
	g.WithWriteLock(func() {
		a.Library.Cards = nil
		for _, c := range library {
			c.InstanceID, c.Owner, c.Controller = uuid.New(), a.ID, a.ID
			a.Library.PushTop(c)
		}
	})
	id := uuid.New()
	a.Hand.PushTop(game.Card{InstanceID: id, Name: spell, OracleID: oracle, TypeLine: "Sorcery", Owner: a.ID, Controller: a.ID})
	if err := g.CastSpell(a.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", spell, err)
	}
	for i := 0; i < 8 && g.ControllerOfPlayer(alice) == alice; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.ControllerOfPlayer(alice) != bob {
		t.Fatal("setup: Alice's search is not controlled by Bob")
	}
	return room, alice, bob
}

// exiledNames lists the names of the cards in exile.
func exiledNames(g *game.Game) map[string]int {
	out := map[string]int{}
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			out[c.Name]++
		}
	})
	return out
}

// "Up to two": the bot finds both, not none.
func TestBotControllingASearchFindsTheMost(t *testing.T) {
	forest := game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}
	room, alice, bob := csTable(t, "Explosive Vegetation", csExplosiveVeg, forest, forest, forest)
	pol := &watching{}
	r := aiseat.NewSteppedRunner(room, bob, pol, aiseat.Config{}, testLogger())
	r.Step(context.Background())
	if pol.sawMoveFor(alice) {
		t.Error("the bot's own policy was handed a move for the searcher")
	}
	if got := exiledNames(room.Game)["Forest"]; got != 2 {
		t.Fatalf("exiled %d Forests, want the two the search allows", got)
	}
	if room.Game.ControllerOfPlayer(alice) != alice {
		t.Error("Alice is still controlled after the search")
	}
	if st := r.Stats(); st.Applied == 0 || st.Rejected != 0 {
		t.Errorf("runner stats %+v, want the take applied and nothing rejected", st)
	}
}

// One card: the bot takes the strongest by mana value.
func TestBotControllingASearchTakesTheBiggest(t *testing.T) {
	room, alice, bob := csTable(t, "Demonic Tutor", csDemonicTutor,
		game.Card{Name: "Small", TypeLine: "Creature — Bear", ManaCost: "{1}{R}"},
		game.Card{Name: "Huge", TypeLine: "Creature — Giant", ManaCost: "{7}{R}{R}"},
		game.Card{Name: "Middle", TypeLine: "Creature — Ogre", ManaCost: "{3}{R}"},
	)
	moves := legal.EnumerateFor(room.Game, bob)
	takes := 0
	for _, m := range moves {
		if m.Player == alice && m.Type == legal.TypeResolveChoice {
			takes++
		}
	}
	if takes < 3 {
		t.Fatalf("setup: the bot is offered %d answers for Alice's search, want every take", takes)
	}
	r := aiseat.NewSteppedRunner(room, bob, &watching{}, aiseat.Config{}, testLogger())
	r.Step(context.Background())
	if got := exiledNames(room.Game); got["Huge"] != 1 || len(got) != 1 {
		t.Fatalf("exiled %v, want the biggest card only", got)
	}
}
