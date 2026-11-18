package effects

import (
	"github.com/google/uuid"

	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const shannaSisaysLegacyOracle = "546201d0-3fa5-44e4-a526-ab0945711e49"

func TestShannaSisaysLegacyGetsPlusOnePerCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shanna := b27Push(g, me.ID, "Shanna, Sisay's Legacy", "Legendary Creature — Human Warrior", shannaSisaysLegacyOracle, "{G}{W}", 0, 0, "G", "W")
	if p, tt := effectivePower(t, g, shanna), effectiveToughness(t, g, shanna); p != 1 || tt != 1 {
		t.Errorf("alone Shanna is %d/%d, want 1/1", p, tt)
	}
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "My Cat", "Creature — Cat", 1, 1)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p, tt := effectivePower(t, g, shanna), effectiveToughness(t, g, shanna); p != 3 || tt != 3 {
		t.Errorf("with two more creatures Shanna is %d/%d, want 3/3 (opponents' creatures don't count)", p, tt)
	}
}

func TestShannaSisaysLegacyRefusesOpponentsAbilitiesButNotSpellsOrYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shanna := b27Push(g, me.ID, "Shanna, Sisay's Legacy", "Legendary Creature — Human Warrior", shannaSisaysLegacyOracle, "{G}{W}", 0, 0, "G", "W")
	theirBallista := b27Push(g, opp.ID, "Walking Ballista", "Artifact Creature — Construct", walkingBallistaOracle, "{X}{X}", 0, 0)
	myBallista := b27Push(g, me.ID, "Walking Ballista", "Artifact Creature — Construct", walkingBallistaOracle, "{X}{X}", 0, 0)
	for _, id := range []uuid.UUID{theirBallista, myBallista} {
		setCountersOn(g, id, map[string]int{game.CounterPlusOne: 2})
	}
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(opp.ID, theirBallista, 1, game.ActivateAbilityParams{Targets: cardRefs(shanna), Strict: true}); err == nil {
		t.Error("an opponent's ability targeted Shanna")
	}
	if err := g.ActivateCatalogAbility(me.ID, myBallista, 1, game.ActivateAbilityParams{Targets: cardRefs(shanna), Strict: true}); err != nil {
		t.Errorf("my own ability could not target Shanna: %v", err)
	}
	passPriorityAroundTable(t, g)
	// An opponent's spell may still target her (it is not hexproof).
	stoke := putInHand(opp, game.Card{Name: "Stoke the Flames", TypeLine: "Instant", OracleID: rprB03StokeOracle, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, stoke, game.CastSpellParams{Targets: cardTarget(shanna), Strict: true}); err != nil {
		t.Errorf("an opponent's spell could not target Shanna: %v", err)
	}
}
