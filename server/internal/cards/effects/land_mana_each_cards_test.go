package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// land_mana_each_cards_test.go — ADR 0200, #3952: Drain Power and Pygmy
// Hippo, where a player activates a mana ability of each land they control
// and loses all unspent mana.

const (
	lmDrainPower   = "0669172d-396b-4f5a-9703-129c5c849b55"
	lmPygmyHippo   = "0d39513a-1f05-4d9a-a772-745f4f1a3471"
	lmCityOfBrass  = "f25351e3-539b-4bbc-b92d-6480acf4d722"
	lmGravenCairns = "5004b84a-33b7-4f6f-b2c2-7086b9087535"
)

func lmBasic(g *game.Game, owner uuid.UUID, kind string, tapped bool) uuid.UUID {
	id := apaPush(g, owner, owner, game.Card{Name: kind, TypeLine: "Basic Land — " + kind})
	if tapped {
		findBattlefieldCardForTest(g, id).Tapped = true
	}
	return id
}

func lmTapped(g *game.Game, id uuid.UUID) bool {
	c := findBattlefieldCardForTest(g, id)
	return c != nil && c.Tapped
}

// lmOptionIndex is the index of the option whose label ends with `suffix`
// on the seat's open option pick.
func lmOptionIndex(t *testing.T, g *game.Game, chooser uuid.UUID, suffix string) int {
	t.Helper()
	c := latestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no option pick for the land: %+v", g.PendingChoices)
	}
	var labels []string
	for i, o := range c.PickOptions {
		if strings.HasSuffix(o.Label, suffix) {
			return i
		}
		labels = append(labels, o.Label)
	}
	t.Fatalf("no option ending %q in %v", suffix, labels)
	return -1
}

// Drain Power: every untapped land of the target is tapped for its mana,
// the target loses that mana and what was already floating, and the caster
// adds all of it. A tapped land adds nothing.
func TestDrainPowerTakesTheTargetsLandManaAndPool(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := lmBasic(g, opp.ID, "Forest", false)
	plains := lmBasic(g, opp.ID, "Plains", false)
	lmBasic(g, opp.ID, "Island", true)
	mine := lmBasic(g, me.ID, "Swamp", false)

	castCatalogSpell(t, g, "Drain Power", "Sorcery", lmDrainPower,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	// Mana the target already has floating as it resolves is lost too
	// (added after the cast: pools empty between steps, CR 106.4).
	g.WithWriteLock(func() { opp.ManaPool.AddMana(game.ManaToken{Color: "R"}) })
	passPriorityAroundTable(t, g)

	if !lmTapped(g, forest) || !lmTapped(g, plains) {
		t.Errorf("the target's untapped lands were not tapped for mana")
	}
	if lmTapped(g, mine) {
		t.Errorf("the caster's own land was tapped")
	}
	if n := opp.ManaPool.Total(); n != 0 {
		t.Errorf("target still has %d mana", n)
	}
	for _, c := range []string{"G", "W", "R"} {
		if me.ManaPool.CountColor(c) != 1 {
			t.Errorf("caster has %d {%s}, want 1 (pool %v)", me.ManaPool.CountColor(c), c, me.ManaPool)
		}
	}
	if n := me.ManaPool.Total(); n != 3 {
		t.Errorf("caster has %d mana, want 3", n)
	}
}

// The target chooses: City of Brass asks which colour, and its own rider
// still deals its damage.
func TestDrainPowerLetsTheTargetChooseTheColour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	apaPush(g, opp.ID, opp.ID, game.Card{Name: "City of Brass", TypeLine: "Land", OracleID: lmCityOfBrass})
	life := opp.Life
	castCatalogSpell(t, g, "Drain Power", "Sorcery", lmDrainPower,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, opp.ID, lmOptionIndex(t, g, opp.ID, "({U})"))
	passPriorityAroundTable(t, g)
	if me.ManaPool.CountColor("U") != 1 || me.ManaPool.Total() != 1 {
		t.Errorf("caster pool = %v, want one {U}", me.ManaPool)
	}
	if opp.Life != life-1 {
		t.Errorf("City of Brass dealt %d damage, want 1", life-opp.Life)
	}
}

// A filter land is asked after the other lands, so their mana can pay for
// its filter ability if the player chooses it.
func TestDrainPowerAsksAFilterLandLast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cairns := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Graven Cairns", TypeLine: "Land", OracleID: lmGravenCairns})
	mountain := lmBasic(g, opp.ID, "Mountain", false)
	castCatalogSpell(t, g, "Drain Power", "Sorcery", lmDrainPower,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if !lmTapped(g, mountain) {
		t.Fatal("the Mountain was not activated before the Cairns was asked")
	}
	answerOptionPick(t, g, opp.ID, lmOptionIndex(t, g, opp.ID, "({R}{R})"))
	passPriorityAroundTable(t, g)
	if !lmTapped(g, cairns) {
		t.Error("the Cairns was not activated")
	}
	if me.ManaPool.CountColor("R") != 2 || me.ManaPool.Total() != 2 {
		t.Errorf("caster pool = %v, want {R}{R} (the Mountain's {R} paid the filter)", me.ManaPool)
	}
}

// Pygmy Hippo: yes taps the defender's lands, empties their pool, stops the
// Hippo's combat damage, and adds that much {C} at the next main phase.
func TestPygmyHippoDrainsTheDefendersManaIntoColorless(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	f1 := lmBasic(g, opp.ID, "Forest", false)
	f2 := lmBasic(g, opp.ID, "Island", false)
	hippo := ncdAttacker(g, "Pygmy Hippo", lmPygmyHippo, 2, 2)
	life := opp.Life
	ncdPlay(t, g, hippo, true)
	if !lmTapped(g, f1) || !lmTapped(g, f2) {
		t.Error("the defender's lands were not tapped")
	}
	if n := opp.ManaPool.Total(); n != 0 {
		t.Errorf("defender still has %d mana", n)
	}
	if lost := ncdLifeLost(g, life); lost != 0 || !ncdMarked(g, hippo) {
		t.Errorf("Hippo dealt %d combat damage (want 0), marked %v", lost, ncdMarked(g, hippo))
	}
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if me.ManaPool.CountColor("C") != 2 || me.ManaPool.Total() != 2 {
		t.Errorf("at the next main phase the Hippo's controller has %v, want {C}{C}", me.ManaPool)
	}
}

// No: nothing is tapped and the Hippo deals its damage.
func TestPygmyHippoDeclined(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	forest := lmBasic(g, opp.ID, "Forest", false)
	hippo := ncdAttacker(g, "Pygmy Hippo", lmPygmyHippo, 2, 2)
	life := opp.Life
	ncdPlay(t, g, hippo, false)
	if lmTapped(g, forest) || ncdMarked(g, hippo) {
		t.Errorf("declined: land tapped %v, marked %v", lmTapped(g, forest), ncdMarked(g, hippo))
	}
	if lost := ncdLifeLost(g, life); lost != 2 {
		t.Errorf("defender lost %d, want 2", lost)
	}
}
