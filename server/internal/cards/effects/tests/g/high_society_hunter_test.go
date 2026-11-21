package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// high_society_hunter_test.go — #2955: the attack trigger's sacrifice
// is chosen as the trigger resolves, targets nothing, and the counter
// follows only a sacrifice that happened.

func hunterAttack(g *game.Game) (me, opp *game.Player, hunter uuid.UUID) {
	me, opp = g.Seats[0], g.Seats[1]
	hunter = effectstest.B42Permanent(g, me.ID, "High-Society Hunter", "Creature — Vampire Noble", effectstest.B42HighSocietyHunterOracle, 5, 3)
	return me, opp, hunter
}

func TestHunterAttackTriggerSacrificesNothingWhileOnTheStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, hunter := hunterAttack(g)
	snack := effectstest.PushVanillaCreature(g, me.ID, "Snack", 1, 1)

	effectstest.DeclareAttack(t, g, opp.ID, hunter)
	if got := effectstest.TriggersOnStackFrom(g, hunter); got != 1 {
		t.Fatalf("triggers on the stack = %d, want 1", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("the trigger asked something before resolving: %+v", g.PendingChoices)
	}
	if !g.Battlefield.Contains(snack) {
		t.Fatal("a creature was sacrificed before the trigger resolved")
	}
	if p := effectstest.CurrentPower(t, g, hunter); p != 5 {
		t.Errorf("power %d before resolution, want 5", p)
	}
}

func TestHunterSacrificesAtResolutionAndGetsTheCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, hunter := hunterAttack(g)
	snack := effectstest.PushVanillaCreature(g, me.ID, "Snack", 1, 1)
	other := effectstest.PushVanillaCreature(g, me.ID, "Other", 1, 1)

	effectstest.DeclareAttack(t, g, opp.ID, hunter)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no creature choice at resolution")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, snack)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(snack) {
		t.Error("the chosen creature survived")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("an unchosen creature was sacrificed")
	}
	if p := effectstest.CurrentPower(t, g, hunter); p != 6 {
		t.Errorf("power %d, want 6 after the counter", p)
	}
}

func TestHunterDecliningLeavesItUnchanged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, hunter := hunterAttack(g)
	snack := effectstest.PushVanillaCreature(g, me.ID, "Snack", 1, 1)

	effectstest.DeclareAttack(t, g, opp.ID, hunter)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(snack) {
		t.Error("declining still sacrificed")
	}
	if p := effectstest.CurrentPower(t, g, hunter); p != 5 {
		t.Errorf("power %d, want 5 with no counter", p)
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a decline still asked for a creature")
	}
}

func TestHunterAloneIsNotAskedAndGetsNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, hunter := hunterAttack(g)
	// An opponent's creature is not yours to sacrifice.
	effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 1, 1)

	effectstest.DeclareAttack(t, g, opp.ID, hunter)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("asked to sacrifice with nothing to sacrifice")
	}
	if p := effectstest.CurrentPower(t, g, hunter); p != 5 {
		t.Errorf("power %d, want 5: no sacrifice, no counter", p)
	}
}

// The old picker was a target clause, so a hexproof creature could not
// be chosen. A sacrifice targets nothing.
func TestHunterMayEatAHexproofCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, hunter := hunterAttack(g)
	bogle := effectstest.PushKeywordCreature(g, me.ID, "Bogle", "hexproof")

	effectstest.DeclareAttack(t, g, opp.ID, hunter)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerChooseCards(t, g, me.ID, bogle)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bogle) {
		t.Error("the hexproof creature was not sacrificed")
	}
	if p := effectstest.CurrentPower(t, g, hunter); p != 6 {
		t.Errorf("power %d, want 6", p)
	}
}

// Removing the creature you would have chosen, in response, cannot
// cost you the sacrifice any more: the choice is made at resolution.
func TestHunterResponseRemovalLeavesAnotherChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, hunter := hunterAttack(g)
	first := effectstest.PushVanillaCreature(g, me.ID, "First", 1, 1)
	second := effectstest.PushVanillaCreature(g, me.ID, "Second", 1, 1)

	effectstest.DeclareAttack(t, g, opp.ID, hunter)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(first) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerChooseCards(t, g, me.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(second) {
		t.Error("the remaining creature was not sacrificed")
	}
	if p := effectstest.CurrentPower(t, g, hunter); p != 6 {
		t.Errorf("power %d, want 6", p)
	}
}

// The counter goes on the same object: a Hunter that left in response
// gets nothing, and the trigger does not error.
func TestHunterGoneBeforeResolutionGetsNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, hunter := hunterAttack(g)
	snack := effectstest.PushVanillaCreature(g, me.ID, "Snack", 1, 1)

	effectstest.DeclareAttack(t, g, opp.ID, hunter)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(hunter) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerChooseCards(t, g, me.ID, snack)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(snack) {
		t.Error("the sacrifice still happens")
	}
	if g.Battlefield.Contains(hunter) {
		t.Error("the Hunter came back")
	}
}

func TestHunterDrawsForOthersNontokenDeathsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, hunter := hunterAttack(g)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	token := effectstest.PushTokens(g, opp.ID, effects.TokenCard("1/1 red Goblin"), 1)[0]

	hand := len(me.Hand.Cards)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(token) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand {
		t.Errorf("a token death drew: hand %d -> %d", hand, got)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("an opponent's nontoken death should draw once: hand %d -> %d", hand, got)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(hunter) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("the Hunter's own death drew: hand %d -> %d", hand+1, got)
	}
}
