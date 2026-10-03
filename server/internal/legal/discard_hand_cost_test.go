package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// discard_hand_cost_test.go — the enumerator half of #1600. #544's
// promise, for a cost that names nothing: offer the activation exactly
// when the engine takes it, with no `discard_ids` (the engine refuses
// them for a hand clause), and never offer Lion's Eye Diamond outside
// the window "Activate only as an instant" leaves open.

const (
	oracleLionsEyeDiamond = "ee6099b0-fb1f-42f1-b862-7708c6e36d05"
	oracleNullBrooch      = "6f885041-3e57-4a69-84f2-fd207ff9f31b"
	oracleSlateOfAncestry = "a07483b4-c04f-42a4-b979-8b77c11fa8f5"
)

func lionsEyeDiamond() game.Card {
	return game.Card{Name: "Lion's Eye Diamond", TypeLine: "Artifact", ManaCost: "{0}", OracleID: oracleLionsEyeDiamond}
}

// One mana move, naming no card, and the dispatcher accepts it — with a
// full hand and with an empty one.
func TestLionsEyeDiamondIsOfferedWithNoDiscardIDs(t *testing.T) {
	for _, empty := range []bool{false, true} {
		g := newTable(t)
		active := g.Seats[g.Turn.ActiveSeat]
		advanceTo(t, g, game.StepPrecombatMain)
		if empty {
			clearHand(active)
		} else {
			handCard(active, basic("Forest", "Forest"))
		}
		led := battlefieldCard(g, active, lionsEyeDiamond())

		moves := legal.EnumerateFor(g, active.ID)
		got := movesFrom(moves, led, legal.KindMana)
		if len(got) != 1 {
			t.Fatalf("empty hand %v: want one mana move from the Diamond, got %v", empty, labels(got))
		}
		if strings.Contains(string(got[0].Params), "discard_ids") {
			t.Errorf("empty hand %v: the move names cards to discard: %s", empty, got[0].Params)
		}
		// The price the params cannot name rides MoveCost.Hand.
		want := active.Hand.Size()
		have := 0
		if got[0].Cost != nil {
			have = got[0].Cost.Hand
		}
		if have != want {
			t.Errorf("empty hand %v: Cost.Hand = %d, want the hand's %d cards", empty, have, want)
		}
		if empty && got[0].Cost != nil {
			t.Errorf("an empty hand costs nothing, but the move carries %+v", *got[0].Cost)
		}
		dispatchAll(t, g, active.ID, got)
	}
}

// Not offered to a seat that does not hold priority, and not while the
// seat owes a prompt (the pay-unless tax the classic illegal play would
// pay with it) — the engine refuses both, so the enumerator must not
// offer either.
func TestLionsEyeDiamondIsNotOfferedOutsideTheInstantWindow(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	led := battlefieldCard(g, active, lionsEyeDiamond())

	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if got := movesFrom(legal.EnumerateFor(g, active.ID), led, legal.KindMana); len(got) != 0 {
		t.Errorf("offered without priority: %v", labels(got))
	}
	for i := 0; i < len(g.Seats)-1; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	holder := g.Seats[g.Turn.PriorityHolder]
	if holder == nil || holder.ID != active.ID {
		t.Skip("priority did not come back to the active seat in this step")
	}
	g.WithWriteLock(func() {
		g.PendingChoices = append(g.PendingChoices, &game.PendingChoice{
			ID: uuid.New(), Kind: game.PendingChoicePayUnless, Chooser: active.ID, Reason: "Mana Leak",
		})
	})
	if got := movesFrom(legal.EnumerateFor(g, active.ID), led, legal.KindMana); len(got) != 0 {
		t.Errorf("offered while a pay-unless prompt is owed: %v", labels(got))
	}
}

// The CR 602 owner: Slate of Ancestry's draw is offered with no
// `discard_ids`, the hand it throws away on MoveCost.Hand and in the
// label, and the move is one the dispatcher accepts. Null Brooch, with
// no spell on the stack to counter, is not offered at all.
func TestDiscardYourHandActivatedAbilitiesAreOfferedWithNoDiscardIDs(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	clearHand(active)
	handCard(active, basic("Island", "Island"))
	handCard(active, basic("Forest", "Forest"))
	slate := battlefieldCard(g, active, game.Card{Name: "Slate of Ancestry", TypeLine: "Artifact", ManaCost: "{4}", OracleID: oracleSlateOfAncestry})
	brooch := battlefieldCard(g, active, game.Card{Name: "Null Brooch", TypeLine: "Artifact", ManaCost: "{4}", OracleID: oracleNullBrooch})
	for i := 0; i < 6; i++ {
		active.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}

	moves := legal.EnumerateFor(g, active.ID)
	got := movesFrom(moves, slate, legal.KindActivate)
	if len(got) != 1 {
		t.Fatalf("want one Slate activation, got %v", labels(got))
	}
	if strings.Contains(string(got[0].Params), "discard_ids") {
		t.Errorf("the Slate move names cards to discard: %s", got[0].Params)
	}
	if got[0].Cost == nil || got[0].Cost.Hand != 2 {
		t.Errorf("Cost = %+v, want Hand: 2", got[0].Cost)
	}
	if !strings.Contains(got[0].Label, "discarding your hand (2 cards)") {
		t.Errorf("label %q does not say the hand goes", got[0].Label)
	}
	// No spell on the stack: the Brooch has nothing to target.
	if b := movesFrom(moves, brooch, legal.KindActivate); len(b) != 0 {
		t.Errorf("Null Brooch offered with no spell to counter: %v", labels(b))
	}
	dispatchAll(t, g, active.ID, got)
}
