package testsst

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	city1 := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "City of Brass", TypeLine: "Land", OracleID: effectstest.LmCityOfBrass})
	forest := effectstest.LmBasic(g, opp.ID, "Forest", false)
	city2 := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "City of Brass", TypeLine: "Land", OracleID: effectstest.LmCityOfBrass})
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Drain Power", "Sorcery", effectstest.LmDrainPower,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.RestoreAtKeyedPrompt(t, g, "option-pick/land-mana-each")
	opp = g.Seats[1]
	if effectstest.LmTapped(g, forest) || effectstest.LmTapped(g, city2) {
		t.Fatal("the restored table walked past the first City")
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, effectstest.LmOptionIndex(t, g, opp.ID, "({U})"))
	if !effectstest.LmTapped(g, city1) || !effectstest.LmTapped(g, forest) {
		t.Fatalf("after the first answer: City %v, Forest %v tapped, want both", effectstest.LmTapped(g, city1), effectstest.LmTapped(g, forest))
	}
	if effectstest.LmTapped(g, city2) {
		t.Fatal("the second City was tapped before it was asked")
	}

	g = effectstest.RestoreAtKeyedPrompt(t, g, "option-pick/land-mana-each")
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AnswerOptionPick(t, g, opp.ID, effectstest.LmOptionIndex(t, g, opp.ID, "({R})"))
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	cairns := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Graven Cairns", TypeLine: "Land", OracleID: effectstest.LmGravenCairns})
	mountain := effectstest.LmBasic(g, opp.ID, "Mountain", false)
	effectstest.CastCatalogSpell(t, g, "Drain Power", "Sorcery", effectstest.LmDrainPower,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.LmTapped(g, mountain) {
		t.Fatal("the Mountain was not activated before the Cairns was asked")
	}

	g = effectstest.RestoreAtKeyedPrompt(t, g, "option-pick/land-mana-each")
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AnswerOptionPick(t, g, opp.ID, effectstest.LmOptionIndex(t, g, opp.ID, "({R}{R})"))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.LmTapped(g, cairns) {
		t.Error("the Cairns was not activated after the restored answer")
	}
	if me.ManaPool.CountColor("R") != 2 || me.ManaPool.Total() != 2 {
		t.Errorf("caster pool = %v, want {R}{R}", me.ManaPool)
	}
}
