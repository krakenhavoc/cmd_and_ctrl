package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_copies_together_test.go — #4594: a "for each creature, create a
// token that's a copy of it" effect creates every copy at the same time,
// in one creation (CR 608.2f), so the copies enter the battlefield in
// one event and each newcomer is checked for the others' entry
// (CR 603.6a). The probe is exile_return_together_test.go's: a Soul
// Warden ("whenever another creature enters, you gain 1 life") token
// copy that enters with another sees it enter; one created before the
// other does not. Mister Fantastic's "whenever one or more tokens you
// control enter" is the batch trigger: one event, one trigger
// (CR 603.2c).

// tctSoulWarden puts a Soul Warden under p's control.
func tctSoulWarden(g *game.Game, p uuid.UUID) uuid.UUID {
	return b27Push(g, p, "Soul Warden", "Creature — Human Cleric", b04SoulWardenOracle, "{W}", 1, 1, "W")
}

// tctMisterFantastic puts Mister Fantastic under p's control.
func tctMisterFantastic(g *game.Game, p uuid.UUID) uuid.UUID {
	return b27Push(g, p, "Mister Fantastic, Reed Richards", "Legendary Creature — Human Scientist Hero", mshA5FantasticOracle, "{3}{U}", 2, 4, "U")
}

// tctSettle resolves the stack, saying yes to every optional trigger
// chooser is asked about and taking every trigger order as offered. A
// dozen triggers is more than assistCardSettle and passPriorityAroundTable
// are bounded for, so it passes priority one seat at a time.
func tctSettle(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 256; i++ {
		if c := assistCardTriggerPrompt(g, chooser); c != nil {
			if err := g.ResolveTriggerPrompt(c.ID, c.Chooser, true); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			continue
		}
		if triggerOrderPrompt(g) != nil {
			answerTriggerOrderInOfferedOrder(t, g)
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	var open []game.PendingChoiceKind
	for _, c := range g.PendingChoices {
		if c != nil {
			open = append(open, c.Kind)
		}
	}
	t.Fatalf("the stack did not settle: stack %d, pending triggers %d, open prompts %v", g.Stack.Size(), len(g.PendingTriggers), open)
}

// tctTokens lists the tokens controller controls.
func tctTokens(g *game.Game, controller uuid.UUID) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && IsToken(c) {
				out = append(out, c)
			}
		}
	})
	return out
}

func TestKindredChargeCreatesItsCopiesAtTheSameTime(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tctSoulWarden(g, me.ID)
	tctSoulWarden(g, me.ID)
	tctMisterFantastic(g, me.ID) // a Human Scientist Hero, not a Cleric: watches, is not copied
	life := me.Life

	castAndPause(t, g, "Kindred Charge", "Sorcery", kindredChargeOracle)
	hand := me.Hand.Size()
	answerCreatureType(t, g, me.ID, "Cleric")
	tctSettle(t, g, me.ID)

	tokens := tctTokens(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	tctSoulWarden(g, me.ID)
	tctSoulWarden(g, me.ID)
	tctMisterFantastic(g, me.ID)
	life := me.Life

	castCatalogSpell(t, g, "Multiversal Incursion", "Sorcery", mshA6IncursionOracle, nil)
	hand := me.Hand.Size()
	tctSettle(t, g, me.ID)

	tokens := tctTokens(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, game.Card{
			Name: "Soul Warden", OracleID: b04SoulWardenOracle, TypeLine: "Token Creature — Human Cleric", Power: 1, Toughness: 1,
		}, 2); err != nil {
			t.Fatal(err)
		}
	})
	tctSettle(t, g, me.ID)
	life := me.Life

	castCatalogSpell(t, g, "Second Harvest", "Instant", "61a6341e-0d36-44b6-8026-9c87c8ef61cb", nil)
	tctSettle(t, g, me.ID)

	if got := len(tctTokens(g, me.ID)); got != 4 {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	tctSoulWarden(g, me.ID)

	castAndPause(t, g, "Kindred Charge", "Sorcery", kindredChargeOracle)
	answerCreatureType(t, g, me.ID, "Cleric")
	tctSettle(t, g, me.ID)

	tokens := tctTokens(g, me.ID)
	if len(tokens) != 1 {
		t.Fatalf("%d tokens, want 1", len(tokens))
	}
	tok := tokens[0]
	if !hasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
		t.Error("the copy has no haste")
	}
	if hasKw(tok.Keywords, "haste") {
		t.Error("haste became the copy's printed text, so a copy of it would have haste")
	}
}

// A copy whose entry asks a question (Adaptive Automaton's creature type)
// pauses the creation; when it is answered, the copies still gain haste
// and are still exiled at the next end step. Undoing the answer and giving
// it again lands the same way.
func TestKindredChargeFinishesAfterACopysEntryChoiceAndUndo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Adaptive Automaton", "Artifact Creature — Construct", "53c730c6-2f8c-4af8-b400-b9d573a71e60", false)
	pushCatalogPermanent(g, me.ID, "Adaptive Automaton", "Artifact Creature — Construct", "53c730c6-2f8c-4af8-b400-b9d573a71e60", false)
	for i := range g.Battlefield.Cards {
		g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 2
	}

	castAndPause(t, g, "Kindred Charge", "Sorcery", kindredChargeOracle)
	answerCreatureType(t, g, me.ID, "Construct")
	if pendingOfKind(g, game.PendingChoiceCreatureType) == nil {
		t.Fatal("the copy's entry did not ask for its creature type")
	}
	snapshot := g.Clone()
	for run := 0; run < 2; run++ {
		if run > 0 {
			g.WithWriteLock(func() { g.RestoreFrom(snapshot) })
		}
		for i := 0; i < 2 && pendingOfKind(g, game.PendingChoiceCreatureType) != nil; i++ {
			answerCreatureType(t, g, me.ID, "Construct")
		}
		tokens := tctTokens(g, me.ID)
		if len(tokens) != 2 {
			t.Fatalf("run %d: %d tokens, want 2", run, len(tokens))
		}
		for _, tok := range tokens {
			if !hasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
				t.Errorf("run %d: a copy has no haste", run)
			}
		}
		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		if left := tctTokens(g, me.ID); len(left) != 0 {
			t.Errorf("run %d: %d copies survived the end step", run, len(left))
		}
	}
}
