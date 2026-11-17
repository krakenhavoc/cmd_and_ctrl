package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// minus_counter_payment_test.go — #5746: a pay-unless whose payment puts
// a -1/-1 counter on a creature the payer controls (Thelon's Chant) is a
// move per creature they control, tapped or not, each naming it, plus the
// decline. Another player's creature and a noncreature are never offered.
func TestMinusOneCounterPaymentOffersAMovePerCreatureYouControl(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	known := map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		known[s.ID] = true
	}
	push := func(owner uuid.UUID, name, typeLine string, tapped bool) uuid.UUID {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2,
			Owner: owner, Controller: owner, Tapped: tapped, KnownBy: known})
		return id
	}
	source := push(other.ID, "Thelon's Chant", "Enchantment", false)
	a := push(seat.ID, "Bear A", "Creature — Bear", false)
	b := push(seat.ID, "Tired Bear", "Creature — Bear", true)
	push(seat.ID, "Rock", "Artifact", false)
	push(other.ID, "Their Bear", "Creature — Bear", false)
	g.WithWriteLock(func() {
		if err := g.QueueUpkeepPayUnlessForEffect(game.UpkeepPayUnlessPrompt{
			Chooser: seat.ID, Source: source, Cost: "Put a -1/-1 counter on a creature you control", Question: "Thelon's Chant?",
			Action: &game.PayAction{Kind: game.PayActionMinusOneCounter, Count: 1, Of: game.PermanentQuery{Types: []string{"creature"}}},
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
		t.Fatalf("moves offer %v and %d declines, want both of the seat's creatures and one decline", offered, declines)
	}
}
