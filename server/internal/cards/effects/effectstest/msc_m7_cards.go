package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mscM7Hand puts a card with the given types in a hand.
func MscM7Hand(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// --- Spider-Rex and Pharaoh Rama-Tut: ward {2} --------------------------------------------

func MscM7WardDeclined(t *testing.T, name, typeLine, oracle string) {
	t.Helper()
	g, _, opp, warded := WardTable(t, name, typeLine, oracle)
	blade := CastAtWardedCreature(t, g, opp, warded)
	for i := 0; i < 8 && !HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !HasPayUnlessFor(g, opp.ID) {
		t.Fatalf("%s: no ward payment prompt for the spell's controller", name)
	}
	AnswerPayUnless(t, g, opp.ID, false)
	PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(warded) {
		t.Errorf("%s died to a spell whose controller declined the ward {2}", name)
	}
	if !opp.Graveyard.Contains(blade) {
		t.Errorf("%s: the countered spell is not in its owner's graveyard", name)
	}
}
