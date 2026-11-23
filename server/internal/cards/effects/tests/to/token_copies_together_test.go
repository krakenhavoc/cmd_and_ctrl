package teststo

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestKindredChargeCreatesItsCopiesAtTheSameTime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.TctSoulWarden(g, me.ID)
	effectstest.TctSoulWarden(g, me.ID)
	effectstest.TctMisterFantastic(g, me.ID) // a Human Scientist Hero, not a Cleric: watches, is not copied
	life := me.Life

	effectstest.CastAndPause(t, g, "Kindred Charge", "Sorcery", effectstest.KindredChargeOracle)
	hand := me.Hand.Size()
	effectstest.AnswerCreatureType(t, g, me.ID, "Cleric")
	effectstest.TctSettle(t, g, me.ID)

	tokens := effectstest.TctTokens(g, me.ID)
	if len(tokens) != 2 {
		t.Fatalf("%d tokens, want 2 (one per Cleric)", len(tokens))
	}
	// Each original sees both copies enter (2 x 2), and each copy sees
	// the other, because they enter together (2 x 1).
	if got := me.Life - life; got != 6 {
		t.Errorf("gained %d life, want 6: each Soul Warden copy sees the other enter (CR 603.6a)", got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1: two tokens entering together are one batch (CR 603.2c)", got)
	}
}

func TestMultiversalIncursionCreatesItsCopiesAtTheSameTime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.TctSoulWarden(g, me.ID)
	effectstest.TctSoulWarden(g, me.ID)
	effectstest.TctMisterFantastic(g, me.ID)
	life := me.Life

	effectstest.CastCatalogSpell(t, g, "Multiversal Incursion", "Sorcery", effectstest.MshA6IncursionOracle, nil)
	hand := me.Hand.Size()
	effectstest.TctSettle(t, g, me.ID)

	tokens := effectstest.TctTokens(g, me.ID)
	if len(tokens) != 3 {
		t.Fatalf("%d tokens, want 3 (one per nontoken creature)", len(tokens))
	}
	// Three creatures enter together. Each original Soul Warden sees all
	// three (2 x 3); each Soul Warden copy sees the two others (2 x 2).
	if got := me.Life - life; got != 10 {
		t.Errorf("gained %d life, want 10: each Soul Warden copy sees the others enter (CR 603.6a)", got)
	}
	// Mister Fantastic triggers once for the batch, and so does his
	// copy, which is one of the tokens entering (CR 603.6a, 603.2c).
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2: one trigger for each Mister Fantastic, for the one batch", got)
	}
}

// The other "for each" token copiers share the creation: Second Harvest's
// copies of two Soul Warden tokens see each other enter.
func TestSecondHarvestCreatesItsCopiesAtTheSameTime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, game.Card{
			Name: "Soul Warden", OracleID: effectstest.B04SoulWardenOracle, TypeLine: "Token Creature — Human Cleric", Power: 1, Toughness: 1,
		}, 2); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.TctSettle(t, g, me.ID)
	life := me.Life

	effectstest.CastCatalogSpell(t, g, "Second Harvest", "Instant", "61a6341e-0d36-44b6-8026-9c87c8ef61cb", nil)
	effectstest.TctSettle(t, g, me.ID)

	if got := len(effectstest.TctTokens(g, me.ID)); got != 4 {
		t.Fatalf("%d tokens, want 4", got)
	}
	if got := me.Life - life; got != 6 {
		t.Errorf("gained %d life, want 6: each Soul Warden copy sees the other enter (CR 603.6a)", got)
	}
}

// Kindred Charge's haste is "those tokens gain haste", an effect of the
// spell and not an exception to the copy (CR 707.9a), so a copy of one of
// those tokens has no haste.
func TestKindredChargeHasteIsAGrantNotACopiableValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.TctSoulWarden(g, me.ID)

	effectstest.CastAndPause(t, g, "Kindred Charge", "Sorcery", effectstest.KindredChargeOracle)
	effectstest.AnswerCreatureType(t, g, me.ID, "Cleric")
	effectstest.TctSettle(t, g, me.ID)

	tokens := effectstest.TctTokens(g, me.ID)
	if len(tokens) != 1 {
		t.Fatalf("%d tokens, want 1", len(tokens))
	}
	tok := tokens[0]
	if !effectstest.HasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
		t.Error("the copy has no haste")
	}
	if effectstest.HasKw(tok.Keywords, "haste") {
		t.Error("haste became the copy's printed text, so a copy of it would have haste")
	}
}

// A copy whose entry asks a question (Adaptive Automaton's creature type)
// pauses the creation; when it is answered, the copies still gain haste
// and are still exiled at the next end step. Undoing the answer and giving
// it again lands the same way.
func TestKindredChargeFinishesAfterACopysEntryChoiceAndUndo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Adaptive Automaton", "Artifact Creature — Construct", "53c730c6-2f8c-4af8-b400-b9d573a71e60", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Adaptive Automaton", "Artifact Creature — Construct", "53c730c6-2f8c-4af8-b400-b9d573a71e60", false)
	for i := range g.Battlefield.Cards {
		g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 2
	}

	effectstest.CastAndPause(t, g, "Kindred Charge", "Sorcery", effectstest.KindredChargeOracle)
	effectstest.AnswerCreatureType(t, g, me.ID, "Construct")
	if effectstest.PendingOfKind(g, game.PendingChoiceCreatureType) == nil {
		t.Fatal("the copy's entry did not ask for its creature type")
	}
	snapshot := g.Clone()
	for run := 0; run < 2; run++ {
		if run > 0 {
			g.WithWriteLock(func() { g.RestoreFrom(snapshot) })
		}
		for i := 0; i < 2 && effectstest.PendingOfKind(g, game.PendingChoiceCreatureType) != nil; i++ {
			effectstest.AnswerCreatureType(t, g, me.ID, "Construct")
		}
		tokens := effectstest.TctTokens(g, me.ID)
		if len(tokens) != 2 {
			t.Fatalf("run %d: %d tokens, want 2", run, len(tokens))
		}
		for _, tok := range tokens {
			if !effectstest.HasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
				t.Errorf("run %d: a copy has no haste", run)
			}
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		if left := effectstest.TctTokens(g, me.ID); len(left) != 0 {
			t.Errorf("run %d: %d copies survived the end step", run, len(left))
		}
	}
}
