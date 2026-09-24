package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// must_attack_view_test.go — #1571, ADR 0045 Decision 51: the wire half
// of attack requirements. `CardView.must_attack` marks the creature the
// active player owes an attack with, and clears once the declaration
// obeys the requirement, so the client badges it and holds autopass
// without deriving a rule.
func TestMustAttackIsStampedUntilTheRequirementIsObeyed(t *testing.T) {
	g := threeSeatsInDeclareAttackers(t)
	active := g.Seats[g.Turn.ActiveSeat]
	goader := g.Seats[(g.Turn.ActiveSeat+1)%3]
	other := g.Seats[(g.Turn.ActiveSeat+2)%3]
	bear := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: bear, Name: "Goaded Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: active.ID, Controller: active.ID,
	})
	plain := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: plain, Name: "Plain Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: active.ID, Controller: active.ID,
	})
	if err := g.SetGoaded(bear, goader.ID); err != nil {
		t.Fatal(err)
	}
	stamp := func() map[string]bool {
		out := map[string]bool{}
		for _, c := range ViewOfGame(g).Battlefield.Cards {
			out[c.InstanceID] = c.MustAttack
		}
		return out
	}
	s := stamp()
	if !s[bear.String()] {
		t.Error("the goaded creature is not marked must_attack")
	}
	if s[plain.String()] {
		t.Error("a creature with no requirement is marked must_attack")
	}
	if err := g.DeclareAttacker(bear, other.ID); err != nil {
		t.Fatal(err)
	}
	if stamp()[bear.String()] {
		t.Error("must_attack survived the declaration that obeys it")
	}
}
