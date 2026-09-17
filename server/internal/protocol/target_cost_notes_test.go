package protocol

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestTargetCostNotesStampedForTheOwnerOnly pins where the X picker's
// cost note comes from (#746, ADR 0048 addendum §15 as amended): the
// snapshot, on the caster's own hand card. Fireball carries its printed
// per-target clause; Lightning Bolt, with no target-priced modifier,
// carries none; and an opponent who has seen the Fireball (a
// Thoughtseize-style reveal) gets the card without the clause, the same
// as the other owner-only cast clauses.
func TestTargetCostNotesStampedForTheOwnerOnly(t *testing.T) {
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := make([]game.Card, 10)
		for j := range deck {
			deck[j] = game.NewCard("filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer %d: %v", i, err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	me, them := g.Seats[0], g.Seats[1]

	fireball := game.Card{
		InstanceID: uuid.New(), Name: "Fireball", TypeLine: "Sorcery", ManaCost: "{X}{R}",
		OracleID: "aa7714b0-2bfb-458a-8ebf-37ec2c53383e", Owner: me.ID, Controller: me.ID,
	}
	fireball.AddKnowersAll([]uuid.UUID{me.ID, them.ID})
	bolt := game.Card{
		InstanceID: uuid.New(), Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
		OracleID: "4457ed35-7c10-48c8-9776-456485fdf070", Owner: me.ID, Controller: me.ID,
	}
	bolt.AddKnower(me.ID)
	me.Hand.PushTop(fireball)
	me.Hand.PushTop(bolt)

	const clause = "This spell costs {1} more to cast for each target beyond the first."
	find := func(v GameView, seat uuid.UUID, id uuid.UUID) *CardView {
		for si := range v.Seats {
			if v.Seats[si].ID != seat.String() {
				continue
			}
			for ci := range v.Seats[si].Hand.Cards {
				if v.Seats[si].Hand.Cards[ci].InstanceID == id.String() {
					return &v.Seats[si].Hand.Cards[ci]
				}
			}
		}
		return nil
	}

	mine := FilterViewFor(ViewOfGame(g), me.ID.String())
	fb := find(mine, me.ID, fireball.InstanceID)
	if fb == nil {
		t.Fatal("Fireball missing from its owner's hand view")
	}
	if len(fb.TargetCostNotes) != 1 || fb.TargetCostNotes[0] != clause {
		t.Errorf("owner's Fireball target_cost_notes = %q, want [%q]", fb.TargetCostNotes, clause)
	}
	lb := find(mine, me.ID, bolt.InstanceID)
	if lb == nil {
		t.Fatal("Lightning Bolt missing from its owner's hand view")
	}
	if lb.TargetCostNotes != nil {
		t.Errorf("Lightning Bolt target_cost_notes = %q, want none", lb.TargetCostNotes)
	}

	theirs := FilterViewFor(ViewOfGame(g), them.ID.String())
	seen := find(theirs, me.ID, fireball.InstanceID)
	if seen == nil || seen.Name != "Fireball" {
		t.Fatalf("the revealed Fireball should reach the opponent face up; got %+v", seen)
	}
	if seen.TargetCostNotes != nil {
		t.Errorf("opponent's view of a revealed Fireball carries target_cost_notes %q, want none", seen.TargetCostNotes)
	}
}
