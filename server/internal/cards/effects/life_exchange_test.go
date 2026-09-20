package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// life_exchange_test.go — the four "exchange life totals" cards of
// issue #1117. Every assertion is on observable state: the two life
// totals, the cards drawn, and for Tree of Perdition the post-layer
// toughness read the way the wire reads it.

const (
	axisOfMortalityOracle  = "158cb888-a115-4d50-a82f-746622f0578e"
	magusOfTheMirrorOracle = "f587676e-d45f-4b1d-be2c-b404d8f55fb3"
	misterNegativeOracle   = "c8da3262-826d-4ea8-bb22-06374bebefe5"
	treeOfPerditionOracle  = "0cdece2a-0bdf-4e6d-9ddc-4a8d58b2ec29"
)

func TestLifeExchangeCardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		axisOfMortalityOracle:  "Axis of Mortality",
		magusOfTheMirrorOracle: "Magus of the Mirror",
		misterNegativeOracle:   "Mister Negative",
		treeOfPerditionOracle:  "Tree of Perdition",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// lxPickPlayers answers the pick_target prompt addressed to chooser
// with the given player slots, in order.
func lxPickPlayers(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	refs := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		refs = append(refs, game.TargetRef{Kind: game.TargetPlayer, ID: id})
	}
	if err := g.ResolvePickTargets(p.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

// lxUntap clears the tapped flag on a battlefield permanent, so a
// test can activate a tap ability twice without walking a whole turn
// cycle for the untap step.
func lxUntap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = false
				return
			}
		}
	})
}

// --- Axis of Mortality ---------------------------------------------

// TestAxisOfMortalitySwapsTwoOpponents proves the controller need not
// be one of the two players: seats 1 and 2 trade, seat 0 (who owns
// the enchantment) keeps its own total.
func TestAxisOfMortalitySwapsTwoOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	a.Life, b.Life = 31, 12
	meLife := me.Life

	pushPermanentForTest(g, me.ID, "Axis of Mortality", axisOfMortalityOracle, "Enchantment")
	advanceToUpkeepOf(t, g, 0)

	answerLatestTriggerPrompt(t, g, me.ID, true)
	lxPickPlayers(t, g, me.ID, a.ID, b.ID)
	passPriorityAroundTable(t, g)

	if a.Life != 12 || b.Life != 31 {
		t.Errorf("after the exchange: A = %d (want 12), B = %d (want 31)", a.Life, b.Life)
	}
	if me.Life != meLife {
		t.Errorf("the controller's life moved: %d -> %d", meLife, me.Life)
	}
}

// TestAxisOfMortalityDeclinedChangesNothing — "you may" is a real
// choice, and saying no leaves every total where it was.
func TestAxisOfMortalityDeclinedChangesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	a.Life, b.Life = 31, 12

	pushPermanentForTest(g, me.ID, "Axis of Mortality", axisOfMortalityOracle, "Enchantment")
	advanceToUpkeepOf(t, g, 0)

	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if a.Life != 31 || b.Life != 12 {
		t.Errorf("a declined exchange moved life: A = %d, B = %d", a.Life, b.Life)
	}
}

// --- Magus of the Mirror --------------------------------------------

// TestMagusOfTheMirrorSwapsWithOpponent activates in the controller's
// own upkeep, which is the only legal window, and checks both halves
// plus the sacrifice.
func TestMagusOfTheMirrorSwapsWithOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Life, opp.Life = 8, 34

	magus := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Magus of the Mirror",
		OracleID: magusOfTheMirrorOracle, TypeLine: "Creature — Human Wizard",
		Power: 4, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	advanceToUpkeepOf(t, g, 0)

	if err := g.ActivateCatalogAbility(me.ID, magus, 0,
		game.ActivateAbilityParams{Targets: acPlayerRef(opp.ID)}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(magus) {
		t.Error("the sacrifice cost was not paid at announce")
	}
	if me.Life != 8 || opp.Life != 34 {
		t.Fatalf("the exchange happened before the ability resolved: me %d, opp %d", me.Life, opp.Life)
	}
	passPriorityAroundTable(t, g)

	if me.Life != 34 || opp.Life != 8 {
		t.Errorf("after the exchange: me = %d (want 34), opp = %d (want 8)", me.Life, opp.Life)
	}
}

// TestMagusOfTheMirrorRefusedOutsideYourUpkeep — the activation
// condition is both halves: not this step, and not your turn.
func TestMagusOfTheMirrorRefusedOutsideYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	magus := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Magus of the Mirror",
		OracleID: magusOfTheMirrorOracle, TypeLine: "Creature — Human Wizard",
		Power: 4, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	// The harness parks the cursor on seat 0's DRAW step — its own
	// turn, one step too late.
	acRefused(t, g, me.ID, magus, 0, game.ActivateAbilityParams{Targets: acPlayerRef(opp.ID)})

	// An opponent's upkeep is not "your upkeep" either.
	advanceToUpkeepOf(t, g, 1)
	acRefused(t, g, me.ID, magus, 0, game.ActivateAbilityParams{Targets: acPlayerRef(opp.ID)})
}

// --- Mister Negative -------------------------------------------------

// TestMisterNegativeExchangesAndDraws is the card in the direction it
// is cast for: you are ahead, you hand the total over, and the life
// you lost is the cards you draw.
func TestMisterNegativeExchangesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Life, opp.Life = 30, 25
	handBefore := me.Hand.Size()

	negative := castCatalogSpell(t, g, "Mister Negative", "Legendary Creature — Human Villain",
		misterNegativeOracle, nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				break
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Battlefield.Contains(negative) {
		t.Fatalf("Mister Negative did not resolve to the battlefield")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	lxPickPlayers(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	if me.Life != 25 || opp.Life != 30 {
		t.Errorf("after the exchange: me = %d (want 25), opp = %d (want 30)", me.Life, opp.Life)
	}
	// The cast itself drew nothing; the hand delta is the spell
	// leaving (castCatalogSpell casts from hand) plus the draws.
	if got := me.Hand.Size() - handBefore; got != 5 {
		t.Errorf("hand delta = %d, want 5 (the 5 life lost)", got)
	}
}

// TestMisterNegativeGainingLifeDrawsNothing — "if you LOST life this
// way". Swapping upward is still a legal, useful line and it draws
// no cards.
func TestMisterNegativeGainingLifeDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Life, opp.Life = 11, 33
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Mister Negative", "Legendary Creature — Human Villain",
		misterNegativeOracle, nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				break
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	lxPickPlayers(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	if me.Life != 33 || opp.Life != 11 {
		t.Errorf("after the exchange: me = %d (want 33), opp = %d (want 11)", me.Life, opp.Life)
	}
	if got := me.Hand.Size() - handBefore; got != 0 {
		t.Errorf("hand delta = %d, want 0 — gaining life draws nothing", got)
	}
}

// TestMisterNegativeHasItsPrintedKeywords pins the badge half of the
// card: vigilance and lifelink reach the effective ability list.
func TestMisterNegativeHasItsPrintedKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mister Negative",
		OracleID: misterNegativeOracle, TypeLine: "Legendary Creature — Human Villain",
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	abilities := effectiveAbilities(t, g, id)
	for _, want := range []string{"vigilance", "lifelink"} {
		found := false
		for _, a := range abilities {
			if a == want {
				found = true
			}
		}
		if !found {
			t.Errorf("Mister Negative is missing %q: %v", want, abilities)
		}
	}
}

// --- Tree of Perdition ------------------------------------------------

func lxTree(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tree of Perdition",
		OracleID: treeOfPerditionOracle, TypeLine: "Creature — Plant",
		Power: 0, Toughness: 13, Owner: owner, Controller: owner,
	})
}

// TestTreeOfPerditionExchangesLifeAndToughness is the whole card in
// one activation: the opponent drops to 13, and the Tree's toughness
// becomes what they had — post-layer, read the way the wire reads it.
func TestTreeOfPerditionExchangesLifeAndToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Life = 37

	tree := lxTree(g, me.ID)
	if got := effectiveToughness(t, g, tree); got != 13 {
		t.Fatalf("printed toughness = %d, want 13", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, tree, 0,
		game.ActivateAbilityParams{Targets: acPlayerRef(opp.ID)}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if opp.Life != 37 {
		t.Fatalf("the exchange happened before the ability resolved: opp = %d", opp.Life)
	}
	passPriorityAroundTable(t, g)

	if opp.Life != 13 {
		t.Errorf("opponent's life = %d, want 13", opp.Life)
	}
	if got := effectiveToughness(t, g, tree); got != 37 {
		t.Errorf("Tree's toughness = %d, want 37", got)
	}
}

// TestTreeOfPerditionSecondActivationSwapsFromTheNewValues — the set
// has no duration, so the second activation trades the toughness the
// FIRST one installed, not the printed 13. CR 613.6 timestamp order
// is what makes the newer set win.
func TestTreeOfPerditionSecondActivationSwapsFromTheNewValues(t *testing.T) {
	g := newCatalogGame(t)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	first.Life, second.Life = 37, 22

	tree := lxTree(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, tree, 0,
		game.ActivateAbilityParams{Targets: acPlayerRef(first.ID)}); err != nil {
		t.Fatalf("first ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if first.Life != 13 || effectiveToughness(t, g, tree) != 37 {
		t.Fatalf("first exchange: life %d, toughness %d; want 13 / 37",
			first.Life, effectiveToughness(t, g, tree))
	}

	lxUntap(g, tree)
	if err := g.ActivateCatalogAbility(me.ID, tree, 0,
		game.ActivateAbilityParams{Targets: acPlayerRef(second.ID)}); err != nil {
		t.Fatalf("second ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)

	if second.Life != 37 {
		t.Errorf("second opponent's life = %d, want 37 (the toughness the first swap installed)", second.Life)
	}
	if got := effectiveToughness(t, g, tree); got != 22 {
		t.Errorf("Tree's toughness = %d, want 22", got)
	}
	if first.Life != 13 {
		t.Errorf("the first opponent's life moved on the second activation: %d", first.Life)
	}
}

// TestTreeOfPerditionHasDefender — the printed keyword, and the
// reason a 0/13 sits still.
func TestTreeOfPerditionHasDefender(t *testing.T) {
	g := newCatalogGame(t)
	tree := lxTree(g, g.Seats[0].ID)
	found := false
	for _, a := range effectiveAbilities(t, g, tree) {
		if a == "defender" {
			found = true
		}
	}
	if !found {
		t.Error("Tree of Perdition is missing defender")
	}
}
