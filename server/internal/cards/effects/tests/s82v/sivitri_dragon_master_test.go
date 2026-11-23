package testss82v

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sivitri_dragon_master_test.go — Sivitri, Dragon Master and the attack
// tax paid in life (#5869, #5487).

const sivitriOracle = "5cf86603-cc47-4644-890a-0997c694817f"

func TestSivitriIsRegisteredAndWhole(t *testing.T) {
	spec, ok := effects.Lookup(sivitriOracle)
	if !ok || spec.Name != "Sivitri, Dragon Master" {
		t.Fatalf("Sivitri is not registered: %+v", spec)
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
}

// sivitriTaxTable activates Sivitri's +1 for the active seat, then
// advances to the next seat's declare attackers step. It returns the
// game, Sivitri's controller, the attacking opponent, and Sivitri.
func sivitriTaxTable(t *testing.T) (*game.Game, int, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, seat, me, opp := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Sivitri, Dragon Master", sivitriOracle, 4)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, pw); got != 5 {
		t.Fatalf("loyalty = %d, want 5", got)
	}
	for !(g.Seats[g.Turn.ActiveSeat].ID == opp.ID && g.Turn.Step == game.StepDeclareAttackers) {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	return g, seat, me, opp, pw
}

func sivitriPrice(g *game.Game, attackers []uuid.UUID, target uuid.UUID) game.AttackTaxPrice {
	decls := make([]game.AttackDeclaration, 0, len(attackers))
	for _, a := range attackers {
		decls = append(decls, game.AttackDeclaration{Attacker: a, Target: target})
	}
	return g.PriceAttackDeclaration(decls)
}

func TestSivitriPlusOneChargesTwoLifePerAttacker(t *testing.T) {
	g, seat, me, opp, pw := sivitriTaxTable(t)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	third := g.Seats[(seat+2)%len(g.Seats)]

	price := sivitriPrice(g, []uuid.UUID{bear}, me.ID)
	if price.Life != 2 || price.Cost != "" || price.IsFree() {
		t.Fatalf("attacking Sivitri's controller costs %q and %d life, want 2 life", price.Cost, price.Life)
	}
	if got := price.Describe(); got != "2 life" {
		t.Errorf("Describe = %q, want \"2 life\"", got)
	}
	// "you or planeswalkers you control".
	if p := sivitriPrice(g, []uuid.UUID{bear}, pw); p.Life != 2 {
		t.Errorf("attacking Sivitri herself costs %d life, want 2", p.Life)
	}
	// Nobody else is defended.
	if p := sivitriPrice(g, []uuid.UUID{bear}, third.ID); !p.IsFree() {
		t.Errorf("an attack on another player costs %q / %d life", p.Cost, p.Life)
	}

	before := opp.Life
	if err := g.DeclareAttacker(bear, me.ID); err != nil {
		t.Fatalf("attacking with 2 life to spare: %v", err)
	}
	if opp.Life != before-2 {
		t.Errorf("attacker's life %d, want %d (paid 2)", opp.Life, before-2)
	}
	if me.Life != 40 {
		t.Errorf("defender's life moved to %d", me.Life)
	}
}

func TestSivitriTaxTotalsAcrossSeveralAttackers(t *testing.T) {
	g, _, me, opp, pw := sivitriTaxTable(t)
	a := effectstest.PushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, opp.ID, "B", 2, 2)
	c := effectstest.PushVanillaCreature(g, opp.ID, "C", 2, 2)

	decls := []game.AttackDeclaration{
		{Attacker: a, Target: me.ID},
		{Attacker: b, Target: me.ID},
		{Attacker: c, Target: pw},
	}
	if p := g.PriceAttackDeclaration(decls); p.Life != 6 || len(p.Lines) != 3 {
		t.Fatalf("three attackers cost %d life in %d lines, want 6 in 3", p.Life, len(p.Lines))
	}
	before := opp.Life
	if _, err := g.DeclareAttackersWith(decls, game.DeclareAttackersParams{}); err != nil {
		t.Fatalf("declaring three attackers: %v", err)
	}
	if opp.Life != before-6 {
		t.Errorf("attacker's life %d, want %d (paid 6)", opp.Life, before-6)
	}
}

func TestSivitriTaxRefusesALifeTotalTooLowToPay(t *testing.T) {
	g, _, me, opp, _ := sivitriTaxTable(t)
	a := effectstest.PushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, opp.ID, "B", 2, 2)
	opp.Life = 3

	// 4 life from a total of 3 is not payable (CR 119.4): all or nothing.
	_, err := g.DeclareAttackersWith([]game.AttackDeclaration{
		{Attacker: a, Target: me.ID},
		{Attacker: b, Target: me.ID},
	}, game.DeclareAttackersParams{})
	if !errors.Is(err, game.ErrAttackTaxUnpaid) {
		t.Fatalf("two attackers on 3 life: %v, want ErrAttackTaxUnpaid", err)
	}
	if opp.Life != 3 {
		t.Errorf("a refused declaration took life: %d", opp.Life)
	}
	// One attacker is 2 of the 3.
	if err := g.DeclareAttacker(a, me.ID); err != nil {
		t.Fatalf("one attacker on 3 life: %v", err)
	}
	if opp.Life != 1 {
		t.Errorf("life %d, want 1", opp.Life)
	}
	// And 1 life cannot pay 2.
	if err := g.DeclareAttacker(b, me.ID); !errors.Is(err, game.ErrAttackTaxUnpaid) {
		t.Fatalf("an attacker on 1 life: %v, want ErrAttackTaxUnpaid", err)
	}
	if opp.Life != 1 {
		t.Errorf("a refused attack took life: %d", opp.Life)
	}
}

func TestSivitriTaxEndsAsHerControllersNextTurnBegins(t *testing.T) {
	g, seat, me, opp, _ := sivitriTaxTable(t)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	for g.Turn.ActiveSeat != seat {
		if p := sivitriPrice(g, []uuid.UUID{bear}, me.ID); p.Life != 2 {
			t.Fatalf("on seat %d's turn, step %s, the tax is %d life, want 2", g.Turn.ActiveSeat, g.Turn.Step, p.Life)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if p := sivitriPrice(g, []uuid.UUID{bear}, me.ID); !p.IsFree() {
		t.Errorf("the tax outlived 'until your next turn': %d life", p.Life)
	}
}

func TestSivitriMinusThreeFetchesADragon(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Sivitri, Dragon Master", sivitriOracle, 4)
	dragon := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Spark Dragon", TypeLine: "Creature — Dragon"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, pw); got != 1 {
		t.Errorf("loyalty = %d, want 1", got)
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a non-Dragon was offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, dragon)
	if !me.Hand.Contains(dragon) {
		t.Error("the Dragon is not in hand")
	}
}

func TestSivitriMinusSevenSparesOnlyDragons(t *testing.T) {
	g, seat, me, opp := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Sivitri, Dragon Master", sivitriOracle, 7)
	myDragon := effectstest.PushDiesCreatureForTest(g, me.ID, "My Dragon", "", "Creature — Dragon", 4, 4)
	theirDragon := effectstest.PushDiesCreatureForTest(g, opp.ID, "Their Dragon", "", "Creature — Dragon", 4, 4)
	myBear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirBear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−7: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{myBear, theirBear} {
		if g.Battlefield.Contains(id) {
			t.Error("a non-Dragon creature survived")
		}
	}
	for _, id := range []uuid.UUID{myDragon, theirDragon} {
		if !g.Battlefield.Contains(id) {
			t.Error("a Dragon was destroyed")
		}
	}
}

func TestSivitriMinusSevenNeedsSevenLoyalty(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Sivitri, Dragon Master", sivitriOracle, 6)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("−7 activated from 6 loyalty")
	}
}
