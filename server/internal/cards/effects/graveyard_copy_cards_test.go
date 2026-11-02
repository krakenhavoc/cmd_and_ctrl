package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_copy_cards_test.go — #4797: "you may have this creature enter
// as a copy of any creature card in a graveyard" (CR 707.2, 614.1c),
// through CastSpell, the copy_target prompt and the stack.

const (
	bodyDoubleOracle        = "895527a6-e3b2-454d-a3be-c0036630ec73"
	voqTorchbearerOracle    = "1b8be000-e2fd-4157-ace2-7b8603c03fba"
	superiorSpiderManOracle = "636cc915-9d1b-4ffe-9e74-795b78663911"
)

// castGraveyardCopier puts the creature in the active player's hand with
// its printed body and casts it.
func castGraveyardCopier(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// graveyardCreature puts a vanilla creature card in p's graveyard.
func graveyardCreature(p *game.Player, name, subtype string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — " + subtype,
		ManaCost: "{2}{G}", Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID})
	return id
}

// waitForCopyPrompt passes priority until the copy picker opens.
func waitForCopyPrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 16 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	p := copyPrompt(g)
	if p == nil {
		t.Fatal("no copy-target prompt")
	}
	return p
}

func TestGraveyardCopyBodyDoubleOffersCreatureCardsInEveryGraveyardAndCopiesOne(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	giant := graveyardCreature(opp, "Hill Giant", "Giant", 3, 3)
	bolt := graveyardCardOf(g, opp, "Lightning Bolt", "Instant", "{R}", "")
	onField := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)

	double := castGraveyardCopier(t, g, "Body Double", "Creature — Shapeshifter", bodyDoubleOracle, 0, 0)
	prompt := waitForCopyPrompt(t, g)
	if !slices.Contains(prompt.CopyOptions, giant) {
		t.Errorf("the creature card in a graveyard should be offered: %v", prompt.CopyOptions)
	}
	if slices.Contains(prompt.CopyOptions, bolt) || slices.Contains(prompt.CopyOptions, onField) {
		t.Errorf("only creature CARDS in graveyards are offered, not an instant or a permanent: %v", prompt.CopyOptions)
	}
	resolveWithCopyChoice(t, g, giant)

	got := copyBattlefieldCard(t, g, double)
	if got.Name != "Hill Giant" || got.CurrentPower() != 3 || got.CurrentToughness() != 3 {
		t.Errorf("Body Double = %q %d/%d, want a 3/3 Hill Giant", got.Name, got.CurrentPower(), got.CurrentToughness())
	}
	if !opp.Graveyard.Contains(giant) {
		t.Error("Body Double copies the card; it stays in the graveyard")
	}
}

func TestGraveyardCopyVoqCopiesAGraveyardCardAndIsAKlingonSpyWithMenace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	giant := graveyardCreature(me, "Hill Giant", "Giant", 3, 3)

	voq := castGraveyardCopier(t, g, "Voq, Torchbearer", "Legendary Creature — Klingon Spy", voqTorchbearerOracle, 4, 3)
	resolveWithCopyChoice(t, g, giant)

	got := copyBattlefieldCard(t, g, voq)
	if got.Name != "Hill Giant" || got.CurrentPower() != 3 {
		t.Errorf("Voq = %q power %d, want a copy of the 3/3 Hill Giant", got.Name, got.CurrentPower())
	}
	for _, sub := range []string{"Giant", "Klingon", "Spy"} {
		if !got.HasSubtype(sub) {
			t.Errorf("the copy should be a %s", sub)
		}
	}
	if !eotHasAbility(effectiveAbilities(t, g, voq), "menace") {
		t.Error("the copy should have menace")
	}
	if !me.Graveyard.Contains(giant) {
		t.Error("Voq exiles nothing")
	}
}

func TestGraveyardCopySuperiorSpiderManCopiesAGraveyardCardThenExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	giant := graveyardCreature(opp, "Hill Giant", "Giant", 3, 3)

	sup := castGraveyardCopier(t, g, "Superior Spider-Man", "Legendary Creature — Spider Human Hero", superiorSpiderManOracle, 4, 4)
	prompt := waitForCopyPrompt(t, g)
	if err := g.ResolveCopyTarget(prompt.ID, prompt.Chooser, giant); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}

	got := copyBattlefieldCard(t, g, sup)
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
	for i := 0; i < 16 && !stackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if opp.Graveyard.Contains(giant) || !g.Exile.Contains(giant) {
		t.Error("the copied card should have been exiled")
	}
	if c := copyBattlefieldCard(t, g, sup); c.Name != "Superior Spider-Man" || c.CurrentPower() != 4 {
		t.Errorf("exiling the card changed the copy: %q %d", c.Name, c.CurrentPower())
	}
}

func TestGraveyardCopySuperiorSpiderManDeclinedEntersAsHimselfAndExilesNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	giant := graveyardCreature(opp, "Hill Giant", "Giant", 3, 3)

	sup := castGraveyardCopier(t, g, "Superior Spider-Man", "Legendary Creature — Spider Human Hero", superiorSpiderManOracle, 4, 4)
	resolveWithCopyChoice(t, g, uuid.Nil)

	got := copyBattlefieldCard(t, g, sup)
	if got.Name != "Superior Spider-Man" || got.HasSubtype("Giant") {
		t.Errorf("declined, he should enter as his printed self: %q", got.Name)
	}
	if !opp.Graveyard.Contains(giant) {
		t.Error("nothing was copied, so nothing is exiled")
	}
}
