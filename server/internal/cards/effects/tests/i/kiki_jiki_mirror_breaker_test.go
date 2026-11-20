package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	kikiJikiOracle          = "a34b7416-cfe3-4a1e-a8c1-a3056b747519"
	kikiJikiTypeLine        = "Legendary Creature — Goblin Shaman"
	kikiJikiVisionaryOracle = "c6a3a882-a127-4590-93d7-679ef4313efe" // Elvish Visionary
)

func TestKikiJikiMirrorBreakerIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(kikiJikiOracle)
	if !ok {
		t.Fatal("Kiki-Jiki, Mirror Breaker is not registered")
	}
	if spec.Name != "Kiki-Jiki, Mirror Breaker" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
	if len(spec.Activated) != 1 {
		t.Errorf("%d activated abilities, want the one {T} ability", len(spec.Activated))
	}
}

// TestKikiJikiCopiesWithEntersTriggerHasteAndEndStepSacrifice is the card
// end to end: Kiki-Jiki entered this turn and still taps (its own haste),
// the token copy of Elvish Visionary draws a card as it enters (the copied
// enters trigger), attacks the turn it was made (the "except it has haste"),
// and is sacrificed at the next end step while Kiki-Jiki and the original
// stay.
func TestKikiJikiCopiesWithEntersTriggerHasteAndEndStepSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kiki := effectstest.PushCatalogPermanent(g, me.ID, "Kiki-Jiki, Mirror Breaker", kikiJikiTypeLine, kikiJikiOracle, true)
	visionary := effectstest.B12Push(g, me.ID, "Elvish Visionary", "Creature — Elf Shaman", kikiJikiVisionaryOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	handBefore := me.Hand.Size()

	if err := effectstest.HobB4Activate(t, g, me.ID, kiki, 0, "", effectstest.CardRefs(visionary)); err != nil {
		t.Fatalf("activate the turn Kiki-Jiki entered (it has haste): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.TappedNow(g, kiki) {
		t.Error("Kiki-Jiki did not tap")
	}
	var token uuid.UUID
	for _, id := range effectstest.BattlefieldIDsNamed(g, "Elvish Visionary") {
		if id != visionary {
			token = id
		}
	}
	if token == uuid.Nil {
		t.Fatal("no token copy of Elvish Visionary")
	}
	if c, _ := effectstest.BattlefieldCard(g, token); !c.IsToken() {
		t.Error("the copy is not a token")
	}
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("hand grew by %d, want 1: the copy's enters trigger draws a card", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, token, "haste") {
		t.Fatal("the copy has no haste")
	}

	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, token)
	if opp.Life != life-1 {
		t.Errorf("opponent at %d, want %d: the hasty copy attacks the turn it was made", opp.Life, life-1)
	}

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, token); ok {
		t.Error("the token survived the end step")
	}
	if _, ok := effectstest.BattlefieldCard(g, kiki); !ok {
		t.Error("Kiki-Jiki was sacrificed — \"sacrifice it\" is the token")
	}
	if _, ok := effectstest.BattlefieldCard(g, visionary); !ok {
		t.Error("the copied creature was sacrificed")
	}
}

// TestKikiJikiRefusesALegendaryOrAnOpponentsCreature is the target clause:
// "target nonlegendary creature you control".
func TestKikiJikiRefusesALegendaryOrAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kiki := effectstest.B12Push(g, me.ID, "Kiki-Jiki, Mirror Breaker", kikiJikiTypeLine, kikiJikiOracle, 2, 2)
	legend := effectstest.B12Creature(g, me.ID, "Legendary Bear", "Legendary Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	if err := effectstest.HobB4Activate(t, g, me.ID, kiki, 0, "", effectstest.CardRefs(legend)); err == nil {
		t.Error("targeted a legendary creature")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, kiki, 0, "", effectstest.CardRefs(kiki)); err == nil {
		t.Error("targeted Kiki-Jiki itself, which is legendary")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, kiki, 0, "", effectstest.CardRefs(theirs)); err == nil {
		t.Error("targeted a creature an opponent controls")
	}
	if effectstest.TappedNow(g, kiki) {
		t.Error("a refused activation tapped Kiki-Jiki")
	}
	if n := len(effectstest.BattlefieldIDsNamed(g, "Legendary Bear")); n != 1 {
		t.Errorf("%d Legendary Bears, want 1: nothing was copied", n)
	}
}
