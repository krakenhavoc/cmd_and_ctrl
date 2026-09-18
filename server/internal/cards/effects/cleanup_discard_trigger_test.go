package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cleanup_discard_trigger_test.go — #661 / CR 514.3a from the
// catalog's side.
//
// Five catalog cards watch EventDiscardCard, and the CR 514.1
// hand-size discard is one of the discards they watch. Before the
// fix the engine walked out of the cleanup step the instant the
// discard landed, so those triggers waited on PendingTriggers until
// the NEXT player's upkeep: the life was gained, the counter placed
// and the Treasure made in somebody else's turn, and anything the
// trigger asked about "this turn" read the wrong turn.

const (
	oracleSangromancer   = "920445ab-0ac2-4de7-bc1c-f5e58eb4424c"
	oracleMaraudingMako  = "e349be42-5f14-44a9-9608-281985c10e2d"
	oracleSurlyBadgersar = "0209dc74-ac49-4deb-907a-e9fa49d27a0f"
	oracleMaryAndAnne    = "5182de2d-aceb-450e-bd20-8bc7db124334"
	oracleSkyray         = "3a46d85b-ce1a-4842-a342-92a5bddb1053"
)

// seedWatcher puts a catalog permanent on the battlefield under p.
func seedWatcher(g *game.Game, p *game.Player, name, oracle, typeLine string, power, tough int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		OracleID:   oracle,
		TypeLine:   typeLine,
		Power:      power,
		Toughness:  tough,
		Owner:      p.ID,
		Controller: p.ID,
	})
}

// stockHandForCleanupDiscard replaces p's hand with seven filler
// cards plus `pitch` — one over the default maximum hand size, so
// the cleanup step owes exactly one discard. Returns pitch's ID.
func stockHandForCleanupDiscard(g *game.Game, p *game.Player, pitch game.Card) uuid.UUID {
	p.Hand.Cards = nil
	for i := 0; i < 7; i++ {
		c := game.NewCard("Filler", p.ID)
		c.Controller = p.ID
		p.Hand.PushTop(c)
	}
	pitch.InstanceID = uuid.New()
	pitch.Owner = p.ID
	pitch.Controller = p.ID
	p.Hand.PushTop(pitch)
	return pitch.InstanceID
}

// cleanupDiscard stocks seat `seat`'s hand one card over the cap at
// their end step, steps into the cleanup step, and answers the
// CR 514.1 discard with `pitch`. Stocking at the end step rather than
// earlier keeps the count exact: a hand filled before that seat's
// draw step is one card bigger by the time cleanup counts it.
func cleanupDiscard(t *testing.T, g *game.Game, seat int, pitch game.Card) uuid.UUID {
	t.Helper()
	p := g.Seats[seat]
	advanceToEndStepOf(t, g, seat)
	pitchID := stockHandForCleanupDiscard(g, p, pitch)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep past End: %v", err)
	}
	if g.Turn.Step != game.StepCleanup || g.DiscardPending[p.ID] != 1 {
		t.Fatalf("expected a one-card cleanup discard, at %s pending=%v", g.Turn.Step, g.DiscardPending)
	}
	if err := g.DiscardSelection(p.ID, []uuid.UUID{pitchID}); err != nil {
		t.Fatalf("DiscardSelection: %v", err)
	}
	return pitchID
}

// TestSangromancerGainsLifeInTheCleanupAnOpponentDiscardedIn is the
// issue's repro (#661). Seat 0's Sangromancer watches an opponent
// discarding; seat 1 discards to hand size in their own cleanup
// step. The "may" prompt, the trigger and the life all belong to
// seat 1's turn.
func TestSangromancerGainsLifeInTheCleanupAnOpponentDiscardedIn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedWatcher(g, me, "Sangromancer", oracleSangromancer, "Creature — Vampire Shaman", 3, 3)
	lifeBefore := me.Life

	cleanupDiscard(t, g, 1, game.NewCard("Pitched Card", opp.ID))

	// CR 514.3a: the turn has NOT moved on. Before the fix the game
	// was already in seat 2's upkeep here, with the prompt open there.
	if g.Turn.ActiveSeat != 1 || g.Turn.Step != game.StepCleanup {
		t.Fatalf("the trigger missed its turn: seat %d step %s", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if g.Turn.PriorityHolder != 1 {
		t.Errorf("PriorityHolder = %d, want the active seat 1", g.Turn.PriorityHolder)
	}

	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if me.Life != lifeBefore+3 {
		t.Errorf("Sangromancer life: got %d, want %d", me.Life, lifeBefore+3)
	}
	if g.Turn.ActiveSeat != 1 {
		t.Errorf("the life was gained in seat %d's turn, want seat 1's", g.Turn.ActiveSeat)
	}
}

// TestMaraudingMakoGrowsFromTheHandSizeDiscard is the other side of
// the same window: the ACTIVE player's own watcher, a mandatory
// trigger, and no prompt in the way. It goes on the stack in the
// cleanup step and resolves there.
func TestMaraudingMakoGrowsFromTheHandSizeDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mako := seedWatcher(g, me, "Marauding Mako", oracleMaraudingMako, "Creature — Shark Pirate", 1, 1)

	cleanupDiscard(t, g, 0, game.NewCard("Pitched Card", me.ID))

	if g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepCleanup {
		t.Fatalf("the trigger missed its turn: seat %d step %s", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if triggerOnStack(g, mako) == nil {
		t.Fatalf("the Mako's trigger is not on the stack in the cleanup step")
	}
	passPriorityAroundTable(t, g)

	if n := counterOnBattlefieldCard(g, mako, "+1/+1"); n != 1 {
		t.Errorf("Marauding Mako counters: got %d, want 1", n)
	}
	if g.Turn.ActiveSeat != 0 {
		t.Errorf("the counter was placed in seat %d's turn, want seat 0's", g.Turn.ActiveSeat)
	}
}

// TestCleanupDiscardWatchersFireInTheDiscardingTurn is one assertion
// apiece for the rest of the EventDiscardCard family. Each watcher
// gets its own game so exactly one trigger is on the stack and no
// CR 603.3b ordering prompt is involved — what is being measured is
// the window, not the ordering.
func TestCleanupDiscardWatchersFireInTheDiscardingTurn(t *testing.T) {
	cases := []struct {
		name     string
		oracle   string
		typeLine string
		pitch    game.Card
		check    func(t *testing.T, g *game.Game, watcher uuid.UUID, p *game.Player)
	}{
		{
			name:     "Surly Badgersaur",
			oracle:   oracleSurlyBadgersar,
			typeLine: "Creature — Badger Dinosaur",
			pitch:    game.Card{Name: "Pitched Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
			check: func(t *testing.T, g *game.Game, watcher uuid.UUID, _ *game.Player) {
				if n := counterOnBattlefieldCard(g, watcher, "+1/+1"); n != 1 {
					t.Errorf("Surly Badgersaur counters after a creature card was pitched: got %d, want 1", n)
				}
			},
		},
		{
			name:     "Scrounging Skyray",
			oracle:   oracleSkyray,
			typeLine: "Creature — Fish Pirate",
			pitch:    game.Card{Name: "Pitched Card", TypeLine: "Sorcery"},
			check: func(t *testing.T, g *game.Game, watcher uuid.UUID, _ *game.Player) {
				if n := counterOnBattlefieldCard(g, watcher, "+1/+1"); n != 1 {
					t.Errorf("Scrounging Skyray counters: got %d, want 1", n)
				}
			},
		},
		{
			name:     "Mary Read and Anne Bonny",
			oracle:   oracleMaryAndAnne,
			typeLine: "Legendary Creature — Human Assassin Pirate",
			pitch:    game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
			check: func(t *testing.T, g *game.Game, _ uuid.UUID, p *game.Player) {
				found := false
				for _, c := range g.Battlefield.Cards {
					if c.Controller == p.ID && c.Name == "Treasure" {
						found = true
					}
				}
				if !found {
					t.Errorf("no Treasure token after an Island was pitched to hand size")
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			watcher := seedWatcher(g, me, tc.name, tc.oracle, tc.typeLine, 3, 3)

			cleanupDiscard(t, g, 0, tc.pitch)

			if g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepCleanup {
				t.Fatalf("%s missed its turn: seat %d step %s", tc.name, g.Turn.ActiveSeat, g.Turn.Step)
			}
			passPriorityAroundTable(t, g)
			if g.Turn.ActiveSeat != 0 {
				t.Fatalf("%s resolved in seat %d's turn, want seat 0's", tc.name, g.Turn.ActiveSeat)
			}
			tc.check(t, g, watcher, me)
		})
	}
}

// counterOnBattlefieldCard reads one counter kind off a battlefield
// card, or 0 when the card has left.
func counterOnBattlefieldCard(g *game.Game, id uuid.UUID, kind string) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Counters[kind]
		}
	}
	return 0
}
