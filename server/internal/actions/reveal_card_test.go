package actions

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reveal_card_test.go covers the reveal_card verb (#671) at the
// dispatch layer: payload shape and controller gating. The per-viewer
// consequences are in protocol/reveal_hand_view_test.go.

func firstHandCard(g *game.Game, p *game.Player) uuid.UUID {
	var id uuid.UUID
	g.ReadSnapshot(func() { id = p.Hand.Cards[0].InstanceID })
	return id
}

func handKnownTo(g *game.Game, p *game.Player, id, viewer uuid.UUID) bool {
	var known bool
	g.ReadSnapshot(func() {
		for _, c := range p.Hand.Cards {
			if c.InstanceID == id {
				known = c.KnownBy[viewer]
			}
		}
	})
	return known
}

func TestDispatchRevealCard(t *testing.T) {
	g := newGame(t)
	owner, foe := g.Seats[0], g.Seats[1]
	id := firstHandCard(g, owner)

	if handKnownTo(g, owner, id, foe.ID) {
		t.Fatalf("the opponent knows an untouched hand card")
	}
	act := mustAction(t, TypeRevealCard, params(t, map[string]string{"card_id": id.String(), "to": foe.ID.String()}))
	act.Caller = owner.ID
	if err := Dispatch(g, act); err != nil {
		t.Fatalf("reveal_card to a player: %v", err)
	}
	if !handKnownTo(g, owner, id, foe.ID) {
		t.Errorf("the shown seat did not become a knower")
	}

	other := g.Seats[0].Hand.Cards[1].InstanceID
	act = mustAction(t, TypeRevealCard, params(t, map[string]string{"card_id": other.String()}))
	act.Caller = owner.ID
	if err := Dispatch(g, act); err != nil {
		t.Fatalf("reveal_card to the table: %v", err)
	}
	if !handKnownTo(g, owner, other, foe.ID) {
		t.Errorf("a table reveal did not make the opponent a knower")
	}
}

func TestDispatchRevealCardGatingAndShape(t *testing.T) {
	g := newGame(t)
	owner, foe := g.Seats[0], g.Seats[1]
	id := firstHandCard(g, owner)

	// Another seat cannot reveal your card.
	act := mustAction(t, TypeRevealCard, params(t, map[string]string{"card_id": id.String()}))
	act.Caller = foe.ID
	if err := Dispatch(g, act); !errors.Is(err, game.ErrCardCallerMismatch) {
		t.Errorf("foreign reveal: %v, want ErrCardCallerMismatch", err)
	}
	if handKnownTo(g, owner, id, foe.ID) {
		t.Errorf("a refused reveal still made the opponent a knower")
	}
	// The admin may.
	act.Caller = uuid.Nil
	if err := Dispatch(g, act); err != nil {
		t.Errorf("admin reveal: %v", err)
	}
	for name, p := range map[string]map[string]string{
		"bad card id": {"card_id": "nope"},
		"bad to":      {"card_id": id.String(), "to": "nope"},
	} {
		a := mustAction(t, TypeRevealCard, params(t, p))
		a.Caller = owner.ID
		if err := Dispatch(g, a); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
