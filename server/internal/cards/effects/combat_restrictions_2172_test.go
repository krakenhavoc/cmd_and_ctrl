package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// combat_restrictions_2172_test.go — The Black Gate (#2172): a creature
// that can't be blocked by creatures one player controls this turn.

const blackGateOracle = "40eb9904-dea3-47cf-963a-04821f98ba64"

// --- The Black Gate ---------------------------------------------

// activateBlackGate pays {1}{B}, targets `target` and resolves the
// ability up to the player prompt, which it leaves open.
func activateBlackGate(t *testing.T, g *game.Game, gate, target uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, gate, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("activate The Black Gate: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func blackGateTable(t *testing.T, lives [4]int) (g *game.Game, gate, runner uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	for i, l := range lives {
		g.Seats[i].Life = l
	}
	gate = pushCatalogPermanent(g, g.Seats[0].ID, "The Black Gate", "Legendary Land — Gate", blackGateOracle, false)
	runner = b12Creature(g, g.Seats[0].ID, "Runner", "Creature — Human", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	return g, gate, runner
}

func TestBlackGateBarsTheChosenPlayersCreatures(t *testing.T) {
	g, gate, runner := blackGateTable(t, [4]int{40, 50, 45, 30})
	p1 := g.Seats[1]
	blocker := b12Creature(g, p1.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	activateBlackGate(t, g, gate, runner)
	c := latestOptionPickFor(g, g.Seats[0].ID)
	if c == nil || len(c.PickOptions) != 1 || c.PickOptions[0].Label != p1.Name {
		t.Fatalf("the pool is the player with the most life only: %+v", c)
	}
	answerChoosePlayer(t, g, g.Seats[0].ID, p1)
	passPriorityAroundTable(t, g)

	brAttack(t, g, runner)
	r := brRefusal(t, g.DeclareBlocker(blocker, runner), game.BlockReasonCantBeBlockedBy)
	if r.Label != "creatures "+p1.Name+" controls" {
		t.Errorf("label %q", r.Label)
	}
	if brOffers(brOffered(t, g, p1.ID), blocker, runner) {
		t.Error("the enumerator offers a barred blocker")
	}
}

func TestBlackGateOtherPlayersCreaturesStillBlock(t *testing.T) {
	g, gate, runner := blackGateTable(t, [4]int{40, 30, 50, 30})
	p1, p2 := g.Seats[1], g.Seats[2]
	blocker := b12Creature(g, p1.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	activateBlackGate(t, g, gate, runner)
	answerChoosePlayer(t, g, g.Seats[0].ID, p2)
	passPriorityAroundTable(t, g)

	brAttack(t, g, runner)
	if !brOffers(brOffered(t, g, p1.ID), blocker, runner) {
		t.Error("the enumerator withholds a creature the rule does not name")
	}
	if err := g.DeclareBlocker(blocker, runner); err != nil {
		t.Fatalf("a creature of a player who was not chosen blocks: %v", err)
	}
}

func TestBlackGateTiedForMostLifeOffersBoth(t *testing.T) {
	g, gate, runner := blackGateTable(t, [4]int{40, 50, 50, 30})
	activateBlackGate(t, g, gate, runner)
	c := latestOptionPickFor(g, g.Seats[0].ID)
	if c == nil || len(c.PickOptions) != 2 {
		t.Fatalf("tied leaders are both offered: %+v", c)
	}
}

func TestBlackGateRuleEndsWithTheTurn(t *testing.T) {
	g, gate, runner := blackGateTable(t, [4]int{40, 50, 45, 30})
	activateBlackGate(t, g, gate, runner)
	answerChoosePlayer(t, g, g.Seats[0].ID, g.Seats[1])
	passPriorityAroundTable(t, g)
	if n := scopedBlockRuleCount(g); n != 1 {
		t.Fatalf("the ability registers one block rule, got %d", n)
	}
	advanceToUpkeepOf(t, g, 1)
	if n := scopedBlockRuleCount(g); n != 0 {
		t.Errorf("%d block rules outlived the turn", n)
	}
}
