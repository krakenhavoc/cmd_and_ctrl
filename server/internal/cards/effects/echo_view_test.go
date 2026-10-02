package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// echo_view_test.go — ADR 0108 §5 decisions 4 and 5: what the table sees
// of echo (the "Echo due" marker) and of a non-mana payment (the picker's
// options, the chooser's alone for a discard), and the bot's moves.

func cardViewOf(t *testing.T, g *game.Game, viewer, id uuid.UUID) protocol.CardView {
	t.Helper()
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID == id.String() {
			return c
		}
	}
	t.Fatalf("%s not on the viewed battlefield", id)
	return protocol.CardView{}
}

// The marker is on from entry until the upkeep that charges it, and off
// once the echo has been dealt with.
func TestEchoDueMarker(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	patrol := castEchoCreature(t, g, "Goblin Patrol", "Creature — Goblin", oracleGoblinPatrol)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if !cardViewOf(t, g, them.ID, patrol).EchoDue {
		t.Fatal("a freshly cast echo creature is not marked Echo due")
	}
	if cardViewOf(t, g, me.ID, bear).EchoDue {
		t.Fatal("a creature without echo is marked Echo due")
	}
	leaveUpkeepThenAdvanceTo(t, g, 0)
	passPriorityAroundTable(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.ResolvePayUnless(echoPrompt(g, me.ID).ID, me.ID, true); err != nil {
		t.Fatalf("pay echo: %v", err)
	}
	if cardViewOf(t, g, me.ID, patrol).EchoDue {
		t.Fatal("Goblin Patrol is still marked Echo due after its echo was paid")
	}
}

// A discard payment's options are the chooser's hand: the chooser sees
// them, nobody else does. Every seat sees the payment and its count.
func TestPayCardsViewHidesADiscardFromOtherSeats(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	castEchoCreature(t, g, "Deepcavern Imp", "Creature — Imp Rebel", oracleDeepcavernImp)
	leaveUpkeepThenAdvanceTo(t, g, 0)
	passPriorityAroundTable(t, g)
	prompt := echoPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no echo prompt")
	}
	find := func(viewer uuid.UUID) *protocol.PayCardsView {
		v := protocol.FilterViewFor(protocol.ViewOfGame(g), viewer.String())
		for _, c := range v.PendingChoices {
			if c.ID == prompt.ID.String() {
				return c.PayCards
			}
		}
		t.Fatalf("prompt not in %s's view", viewer)
		return nil
	}
	mine := find(me.ID)
	if mine == nil || mine.Action != "discard" || mine.Count != 1 || len(mine.Options) != me.Hand.Size() {
		t.Fatalf("chooser's pay_cards = %+v, want discard 1 of %d", mine, me.Hand.Size())
	}
	theirs := find(them.ID)
	if theirs == nil || theirs.Count != 1 || len(theirs.Options) != 0 {
		t.Fatalf("another seat's pay_cards = %+v, want the count and no options", theirs)
	}
}

// The enumerator offers one move per way to pay plus the decline, and a
// payment it offers is one the engine takes (#544).
func TestEnumeratorOffersSacrificePaymentsTheEngineAccepts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	surger := castEchoCreature(t, g, "Skizzik Surger", "Creature — Elemental", oracleSkizzikSurger)
	pushLand(g, me.ID, "Island A")
	pushLand(g, me.ID, "Island B")
	pushLand(g, me.ID, "Island C")
	leaveUpkeepThenAdvanceTo(t, g, 0)
	passPriorityAroundTable(t, g)
	if echoPrompt(g, me.ID) == nil {
		t.Fatal("no echo prompt")
	}
	var pays, declines []legal.Move
	for _, m := range legal.EnumerateFor(g, me.ID) {
		var p struct {
			Apply   *bool    `json:"apply"`
			CardIDs []string `json:"card_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil || p.Apply == nil {
			continue
		}
		if *p.Apply {
			if len(p.CardIDs) != 2 {
				t.Fatalf("a pay move names %d permanents, want 2", len(p.CardIDs))
			}
			pays = append(pays, m)
		} else {
			declines = append(declines, m)
		}
	}
	if len(pays) != 3 || len(declines) != 1 {
		t.Fatalf("moves: %d pays, %d declines; want 3 (two of three lands) and 1", len(pays), len(declines))
	}
	m := pays[0]
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: m.Player, Params: m.Params}); err != nil {
		t.Fatalf("the engine refused an enumerated payment: %v", err)
	}
	if _, ok := battlefieldCard(g, surger); !ok {
		t.Fatal("an enumerated payment did not pay the echo")
	}
	if n := countLandsOf(g, me.ID); n != 1 {
		t.Fatalf("lands left = %d, want 1", n)
	}
}

// With nothing to pay with, the decline is the only move.
func TestEnumeratorOffersOnlyTheDeclineWhenThePaymentIsImpossible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castEchoCreature(t, g, "Skizzik Surger", "Creature — Elemental", oracleSkizzikSurger)
	pushLand(g, me.ID, "Island A")
	leaveUpkeepThenAdvanceTo(t, g, 0)
	passPriorityAroundTable(t, g)
	n := 0
	for _, m := range legal.EnumerateFor(g, me.ID) {
		var p struct {
			Apply *bool `json:"apply"`
		}
		if json.Unmarshal(m.Params, &p) == nil && p.Apply != nil {
			if *p.Apply {
				t.Fatal("a pay move was offered with one land for \"Sacrifice two lands\"")
			}
			n++
		}
	}
	if n != 1 {
		t.Fatalf("declines offered = %d, want 1", n)
	}
}

func countLandsOf(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && c.IsLand() {
			n++
		}
	}
	return n
}
