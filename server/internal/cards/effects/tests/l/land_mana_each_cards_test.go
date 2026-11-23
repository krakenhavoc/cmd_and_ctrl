package testsl

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drain Power: every untapped land of the target is tapped for its mana,
// the target loses that mana and what was already floating, and the caster
// adds all of it. A tapped land adds nothing.
func TestDrainPowerTakesTheTargetsLandManaAndPool(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.LmBasic(g, opp.ID, "Forest", false)
	plains := effectstest.LmBasic(g, opp.ID, "Plains", false)
	effectstest.LmBasic(g, opp.ID, "Island", true)
	mine := effectstest.LmBasic(g, me.ID, "Swamp", false)

	effectstest.CastCatalogSpell(t, g, "Drain Power", "Sorcery", effectstest.LmDrainPower,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	// Mana the target already has floating as it resolves is lost too
	// (added after the cast: pools empty between steps, CR 106.4).
	g.WithWriteLock(func() { opp.ManaPool.AddMana(game.ManaToken{Color: "R"}) })
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.LmTapped(g, forest) || !effectstest.LmTapped(g, plains) {
		t.Errorf("the target's untapped lands were not tapped for mana")
	}
	if effectstest.LmTapped(g, mine) {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "City of Brass", TypeLine: "Land", OracleID: effectstest.LmCityOfBrass})
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Drain Power", "Sorcery", effectstest.LmDrainPower,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, opp.ID, effectstest.LmOptionIndex(t, g, opp.ID, "({U})"))
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cairns := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Graven Cairns", TypeLine: "Land", OracleID: effectstest.LmGravenCairns})
	mountain := effectstest.LmBasic(g, opp.ID, "Mountain", false)
	effectstest.CastCatalogSpell(t, g, "Drain Power", "Sorcery", effectstest.LmDrainPower,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.LmTapped(g, mountain) {
		t.Fatal("the Mountain was not activated before the Cairns was asked")
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, effectstest.LmOptionIndex(t, g, opp.ID, "({R}{R})"))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.LmTapped(g, cairns) {
		t.Error("the Cairns was not activated")
	}
	if me.ManaPool.CountColor("R") != 2 || me.ManaPool.Total() != 2 {
		t.Errorf("caster pool = %v, want {R}{R} (the Mountain's {R} paid the filter)", me.ManaPool)
	}
}

// Pygmy Hippo: yes taps the defender's lands, empties their pool, stops the
// Hippo's combat damage, and adds that much {C} at the next main phase.
func TestPygmyHippoDrainsTheDefendersManaIntoColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	f1 := effectstest.LmBasic(g, opp.ID, "Forest", false)
	f2 := effectstest.LmBasic(g, opp.ID, "Island", false)
	hippo := effectstest.NcdAttacker(g, "Pygmy Hippo", effectstest.LmPygmyHippo, 2, 2)
	life := opp.Life
	effectstest.NcdPlay(t, g, hippo, true)
	if !effectstest.LmTapped(g, f1) || !effectstest.LmTapped(g, f2) {
		t.Error("the defender's lands were not tapped")
	}
	if n := opp.ManaPool.Total(); n != 0 {
		t.Errorf("defender still has %d mana", n)
	}
	if lost := effectstest.NcdLifeLost(g, life); lost != 0 || !effectstest.NcdMarked(g, hippo) {
		t.Errorf("Hippo dealt %d combat damage (want 0), marked %v", lost, effectstest.NcdMarked(g, hippo))
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if me.ManaPool.CountColor("C") != 2 || me.ManaPool.Total() != 2 {
		t.Errorf("at the next main phase the Hippo's controller has %v, want {C}{C}", me.ManaPool)
	}
}

// No: nothing is tapped and the Hippo deals its damage.
func TestPygmyHippoDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	forest := effectstest.LmBasic(g, opp.ID, "Forest", false)
	hippo := effectstest.NcdAttacker(g, "Pygmy Hippo", effectstest.LmPygmyHippo, 2, 2)
	life := opp.Life
	effectstest.NcdPlay(t, g, hippo, false)
	if effectstest.LmTapped(g, forest) || effectstest.NcdMarked(g, hippo) {
		t.Errorf("declined: land tapped %v, marked %v", effectstest.LmTapped(g, forest), effectstest.NcdMarked(g, hippo))
	}
	if lost := effectstest.NcdLifeLost(g, life); lost != 2 {
		t.Errorf("defender lost %d, want 2", lost)
	}
}
