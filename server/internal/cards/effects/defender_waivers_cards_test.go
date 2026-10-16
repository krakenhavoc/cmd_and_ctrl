package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// defender_waivers_cards_test.go — ADR 0106's 2026-10-16 amendment
// (#1936): Weathered Sentinels and Hightide Hermit, driven through whole
// turns. The engine half is pinned in game/defender_waivers_test.go.

const (
	weatheredSentinelsOracle = "0903b941-82f1-4d3e-9717-29ab7d0c8cef"
	hightideHermitOracle     = "3689ef94-2573-4424-b52e-379656340345"
)

// wantDefenderRefused asserts the attack is refused for defender.
func wantDefenderRefused(t *testing.T, g *game.Game, attacker, target uuid.UUID) {
	t.Helper()
	if err := declareResult(g, attacker, target); !errors.Is(err, game.ErrDefender) {
		t.Errorf("attack %s: err = %v, want ErrDefender", target, err)
	}
}

// Seat 1 attacks seat 0 and seat 2 attacks seat 3 in their turns; on
// seat 0's next turn the Sentinels may attack seat 1 and nobody else,
// and its attack trigger makes it a 5/8 indestructible.
func TestWeatheredSentinelsAttacksOnlyAPlayerWhoAttackedYou(t *testing.T) {
	g := newCatalogGame(t)
	me, b, c, d := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	sentinels := pushTestPermanent(g, me.ID, game.Card{Name: "Weathered Sentinels", OracleID: weatheredSentinelsOracle,
		TypeLine: "Artifact Creature — Wall", Power: 2, Toughness: 5})
	bBear := pushVanillaCreature(g, b.ID, "B's Bear", 2, 2)
	cBear := pushVanillaCreature(g, c.ID, "C's Bear", 2, 2)
	for _, kw := range []string{"defender", "reach", "vigilance", "trample"} {
		if !hasEffectiveKeyword(t, g, sentinels, kw) {
			t.Fatalf("Weathered Sentinels has no %s", kw)
		}
	}

	advanceToDeclareAttackersOf(t, g, 0)
	for _, p := range []*game.Player{b, c, d} {
		wantDefenderRefused(t, g, sentinels, p.ID)
	}

	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(bBear, me.ID); err != nil {
		t.Fatalf("B attacks me: %v", err)
	}
	advanceToDeclareAttackersOf(t, g, 2)
	if err := g.DeclareAttacker(cBear, d.ID); err != nil {
		t.Fatalf("C attacks D: %v", err)
	}
	advanceToDeclareAttackersOf(t, g, 0)

	wantDefenderRefused(t, g, sentinels, c.ID)
	wantDefenderRefused(t, g, sentinels, d.ID)
	if err := g.DeclareAttacker(sentinels, b.ID); err != nil {
		t.Fatalf("the Sentinels attack the player who attacked me: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, sentinels), effectiveToughness(t, g, sentinels); p != 5 || tough != 8 {
		t.Errorf("after its attack trigger: %d/%d, want 5/8", p, tough)
	}
	if !hasEffectiveKeyword(t, g, sentinels, "indestructible") {
		t.Error("after its attack trigger it should have indestructible")
	}
	if !hasEffectiveKeyword(t, g, sentinels, "defender") {
		t.Error("it attacked as though it didn't have defender, but it still has it")
	}
	if isTapped(g, sentinels) {
		t.Error("vigilance: it should not tap to attack")
	}
}

// Hightide Hermit gets four energy as it enters; paying two lets it
// attack this turn, at any opponent, and it still has defender.
func TestHightideHermitPaysEnergyToAttackThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	hermit := pushTestPermanent(g, me.ID, game.Card{Name: "Hightide Hermit", OracleID: hightideHermitOracle,
		TypeLine: "Creature — Crab", Power: 4, Toughness: 4})
	setEnergy(t, g, me, 4)
	advanceToMainOf(t, g, 0)
	before := g.Clone()
	advanceTo(t, before, game.StepDeclareAttackers)
	wantDefenderRefused(t, before, hermit, b.ID)

	if err := g.ActivateCatalogAbility(me.ID, hermit, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := energyOf(me); got != 2 {
		t.Errorf("energy %d after paying {E}{E}, want 2", got)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, p := range g.Seats[1:] {
		if err := declareResult(g, hermit, p.ID); err != nil {
			t.Errorf("attack %s after paying: %v", p.Name, err)
		}
	}
	if !hasEffectiveKeyword(t, g, hermit, "defender") {
		t.Error("it still has defender")
	}

	// "This turn": by its controller's next turn the waiver has ended.
	advanceToDeclareAttackersOf(t, g, 1)
	advanceToDeclareAttackersOf(t, g, 0)
	wantDefenderRefused(t, g, hermit, b.ID)
}

func TestHightideHermitEntersWithFourEnergy(t *testing.T) {
	g, me, _ := p7Table(t)
	castCatalogSpell(t, g, "Hightide Hermit", "Creature — Crab", hightideHermitOracle, nil)
	passPriorityAroundTable(t, g)
	if got := energyOf(me); got != 4 {
		t.Errorf("energy %d after it entered, want 4", got)
	}
}
