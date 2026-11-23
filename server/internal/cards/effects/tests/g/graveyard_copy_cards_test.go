package testsg

import (
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// graveyard_copy_cards_test.go — #4797: "you may have this creature enter
// as a copy of any creature card in a graveyard" (CR 707.2, 614.1c),
// through CastSpell, the copy_target prompt and the stack.

const (
	bodyDoubleOracle        = "895527a6-e3b2-454d-a3be-c0036630ec73"
	voqTorchbearerOracle    = "1b8be000-e2fd-4157-ace2-7b8603c03fba"
	superiorSpiderManOracle = "636cc915-9d1b-4ffe-9e74-795b78663911"
)

func TestGraveyardCopyBodyDoubleOffersCreatureCardsInEveryGraveyardAndCopiesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	giant := effectstest.GraveyardCreature(opp, "Hill Giant", "Giant", 3, 3)
	bolt := effectstest.GraveyardCardOf(g, opp, "Lightning Bolt", "Instant", "{R}", "")
	onField := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)

	double := effectstest.CastGraveyardCopier(t, g, "Body Double", "Creature — Shapeshifter", bodyDoubleOracle, 0, 0)
	prompt := effectstest.WaitForCopyPrompt(t, g)
	if !slices.Contains(prompt.CopyOptions, giant) {
		t.Errorf("the creature card in a graveyard should be offered: %v", prompt.CopyOptions)
	}
	if slices.Contains(prompt.CopyOptions, bolt) || slices.Contains(prompt.CopyOptions, onField) {
		t.Errorf("only creature CARDS in graveyards are offered, not an instant or a permanent: %v", prompt.CopyOptions)
	}
	effectstest.ResolveWithCopyChoice(t, g, giant)

	got := effectstest.CopyBattlefieldCard(t, g, double)
	if got.Name != "Hill Giant" || got.CurrentPower() != 3 || got.CurrentToughness() != 3 {
		t.Errorf("Body Double = %q %d/%d, want a 3/3 Hill Giant", got.Name, got.CurrentPower(), got.CurrentToughness())
	}
	if !opp.Graveyard.Contains(giant) {
		t.Error("Body Double copies the card; it stays in the graveyard")
	}
}

func TestGraveyardCopyVoqCopiesAGraveyardCardAndIsAKlingonSpyWithMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	giant := effectstest.GraveyardCreature(me, "Hill Giant", "Giant", 3, 3)

	voq := effectstest.CastGraveyardCopier(t, g, "Voq, Torchbearer", "Legendary Creature — Klingon Spy", voqTorchbearerOracle, 4, 3)
	effectstest.ResolveWithCopyChoice(t, g, giant)

	got := effectstest.CopyBattlefieldCard(t, g, voq)
	if got.Name != "Hill Giant" || got.CurrentPower() != 3 {
		t.Errorf("Voq = %q power %d, want a copy of the 3/3 Hill Giant", got.Name, got.CurrentPower())
	}
	for _, sub := range []string{"Giant", "Klingon", "Spy"} {
		if !got.HasSubtype(sub) {
			t.Errorf("the copy should be a %s", sub)
		}
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, voq), "menace") {
		t.Error("the copy should have menace")
	}
	if !me.Graveyard.Contains(giant) {
		t.Error("Voq exiles nothing")
	}
}

func TestGraveyardCopySuperiorSpiderManCopiesAGraveyardCardThenExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	giant := effectstest.GraveyardCreature(opp, "Hill Giant", "Giant", 3, 3)

	sup := effectstest.CastGraveyardCopier(t, g, "Superior Spider-Man", "Legendary Creature — Spider Human Hero", superiorSpiderManOracle, 4, 4)
	prompt := effectstest.WaitForCopyPrompt(t, g)
	if err := g.ResolveCopyTarget(prompt.ID, prompt.Chooser, giant); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}

	got := effectstest.CopyBattlefieldCard(t, g, sup)
	if got.Name != "Superior Spider-Man" || got.CurrentPower() != 4 || got.CurrentToughness() != 4 {
		t.Errorf("entered as %q %d/%d, want Superior Spider-Man 4/4", got.Name, got.CurrentPower(), got.CurrentToughness())
	}
	for _, sub := range []string{"Giant", "Spider", "Human", "Hero"} {
		if !got.HasSubtype(sub) {
			t.Errorf("the copy should be a %s", sub)
		}
	}
	// "When you do" is a trigger: the card is still in the graveyard
	// until it resolves.
	if !opp.Graveyard.Contains(giant) {
		t.Fatal("the card was exiled before the reflexive trigger resolved")
	}
	for i := 0; i < 16 && !effectstest.StackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if opp.Graveyard.Contains(giant) || !g.Exile.Contains(giant) {
		t.Error("the copied card should have been exiled")
	}
	if c := effectstest.CopyBattlefieldCard(t, g, sup); c.Name != "Superior Spider-Man" || c.CurrentPower() != 4 {
		t.Errorf("exiling the card changed the copy: %q %d", c.Name, c.CurrentPower())
	}
}

func TestGraveyardCopySuperiorSpiderManDeclinedEntersAsHimselfAndExilesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	giant := effectstest.GraveyardCreature(opp, "Hill Giant", "Giant", 3, 3)

	sup := effectstest.CastGraveyardCopier(t, g, "Superior Spider-Man", "Legendary Creature — Spider Human Hero", superiorSpiderManOracle, 4, 4)
	effectstest.ResolveWithCopyChoice(t, g, uuid.Nil)

	got := effectstest.CopyBattlefieldCard(t, g, sup)
	if got.Name != "Superior Spider-Man" || got.HasSubtype("Giant") {
		t.Errorf("declined, he should enter as his printed self: %q", got.Name)
	}
	if !opp.Graveyard.Contains(giant) {
		t.Error("nothing was copied, so nothing is exiled")
	}
}
