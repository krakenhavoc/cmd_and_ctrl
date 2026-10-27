package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_10c_land_mana_restore_test.go — #4253 slice 10c: the walk over a
// player's lands (ActivateAManaAbilityOfEachLand) carries the lands still
// to visit as data on its option pick.

// Drain Power over two City of Brass with a Forest between them: the
// walk's first question is a restore point, and so is the second, asked
// only after the restored answer has activated the Forest. The caster ends
// with the two colours chosen, and each City dealt its damage.
func TestDrainPowerWalkIsARestorePointAtEachQuestion(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	city1 := apaPush(g, opp.ID, opp.ID, game.Card{Name: "City of Brass", TypeLine: "Land", OracleID: lmCityOfBrass})
	forest := lmBasic(g, opp.ID, "Forest", false)
	city2 := apaPush(g, opp.ID, opp.ID, game.Card{Name: "City of Brass", TypeLine: "Land", OracleID: lmCityOfBrass})
	life := opp.Life
	castCatalogSpell(t, g, "Drain Power", "Sorcery", lmDrainPower,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	g = restoreAtKeyedPrompt(t, g, "option-pick/land-mana-each")
	opp = g.Seats[1]
	if lmTapped(g, forest) || lmTapped(g, city2) {
		t.Fatal("the restored table walked past the first City")
	}
	answerOptionPick(t, g, opp.ID, lmOptionIndex(t, g, opp.ID, "({U})"))
	if !lmTapped(g, city1) || !lmTapped(g, forest) {
		t.Fatalf("after the first answer: City %v, Forest %v tapped, want both", lmTapped(g, city1), lmTapped(g, forest))
	}
	if lmTapped(g, city2) {
		t.Fatal("the second City was tapped before it was asked")
	}

	g = restoreAtKeyedPrompt(t, g, "option-pick/land-mana-each")
	me, opp := g.Seats[0], g.Seats[1]
	answerOptionPick(t, g, opp.ID, lmOptionIndex(t, g, opp.ID, "({R})"))
	passPriorityAroundTable(t, g)

	if n := opp.ManaPool.Total(); n != 0 {
		t.Errorf("the target still has %d mana", n)
	}
	if me.ManaPool.CountColor("U") != 1 || me.ManaPool.CountColor("R") != 1 || me.ManaPool.CountColor("G") != 1 || me.ManaPool.Total() != 3 {
		t.Errorf("caster pool = %v, want {U}{R}{G}", me.ManaPool)
	}
	if opp.Life != life-2 {
		t.Errorf("the Cities dealt %d damage, want 2", life-opp.Life)
	}
}

// A filter land waits for the others: its question is asked on the second
// pass, and the table waiting on it is a restore point whose walk has
// nothing left but the lost mana.
func TestDrainPowerFilterLandQuestionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	cairns := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Graven Cairns", TypeLine: "Land", OracleID: lmGravenCairns})
	mountain := lmBasic(g, opp.ID, "Mountain", false)
	castCatalogSpell(t, g, "Drain Power", "Sorcery", lmDrainPower,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if !lmTapped(g, mountain) {
		t.Fatal("the Mountain was not activated before the Cairns was asked")
	}

	g = restoreAtKeyedPrompt(t, g, "option-pick/land-mana-each")
	me, opp := g.Seats[0], g.Seats[1]
	answerOptionPick(t, g, opp.ID, lmOptionIndex(t, g, opp.ID, "({R}{R})"))
	passPriorityAroundTable(t, g)
	if !lmTapped(g, cairns) {
		t.Error("the Cairns was not activated after the restored answer")
	}
	if me.ManaPool.CountColor("R") != 2 || me.ManaPool.Total() != 2 {
		t.Errorf("caster pool = %v, want {R}{R}", me.ManaPool)
	}
}
