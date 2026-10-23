package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// tap_payment_test.go — #1997: a pay-unless whose payment taps an
// untapped creature (Koskun Falls) is a move per untapped creature the
// payer controls, each naming it, plus the decline. A tapped creature
// and another player's are never offered.
func TestTapPaymentOffersAMovePerUntappedCreature(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	known := map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		known[s.ID] = true
	}
	push := func(owner uuid.UUID, name string, tapped bool) uuid.UUID {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Owner: owner, Controller: owner, Tapped: tapped, KnownBy: known})
		return id
	}
	source := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Koskun Falls", TypeLine: "World Enchantment", Owner: seat.ID, Controller: seat.ID, KnownBy: known})
	a := push(seat.ID, "Bear A", false)
	b := push(seat.ID, "Bear B", false)
	push(seat.ID, "Tired Bear", true)
	push(other.ID, "Their Bear", false)
	g.WithWriteLock(func() {
		if err := g.QueueUpkeepPayUnlessForEffect(game.UpkeepPayUnlessPrompt{
			Chooser: seat.ID, Source: source, Cost: "Tap an untapped creature you control", Question: "Koskun Falls?",
			Action:    &game.PayAction{Kind: game.PayActionTap, Count: 1, Of: game.PermanentQuery{Types: []string{"creature"}}},
			OnDecline: func(*game.Game) error { return nil },
		}); err != nil {
			t.Fatalf("queue: %v", err)
		}
	})

	offered := map[string]bool{}
	declines := 0
	for _, m := range legal.EnumerateFor(g, seat.ID) {
		if m.Type != legal.TypeResolveChoice {
			continue
		}
		var p struct {
			Apply   *bool    `json:"apply"`
			CardIDs []string `json:"card_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		switch {
		case p.Apply != nil && !*p.Apply:
			declines++
		case len(p.CardIDs) == 1:
			offered[p.CardIDs[0]] = true
		default:
			t.Errorf("unexpected move %+v", m)
		}
	}
	if declines != 1 || len(offered) != 2 || !offered[a.String()] || !offered[b.String()] {
		t.Fatalf("moves offer %v and %d declines, want the two untapped Bears and one decline", offered, declines)
	}
}
