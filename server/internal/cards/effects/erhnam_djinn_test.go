package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// erhnam_djinn_test.go — #5737: "until your next upkeep", on Erhnam Djinn
// and Xenic Poltergeist, driven through the real turn machine on a
// four-seat table.

const (
	erhnamDjinnOracle      = "d48a38c9-3dcd-4c18-8840-1b057ede3ff0"
	xenicPoltergeistOracle = "fb8f80cc-6214-4ab9-a9a9-1873ab9feb0c"
)

// erhnamTable puts Erhnam Djinn and a Forest under seat 1, a creature
// under each of seats 2 and 3, a Wall under seat 2 and another creature
// under seat 1, and walks to seat 1's upkeep with the trigger asking for
// its target.
func erhnamTable(t *testing.T) (g *game.Game, djinn, bear2, bear3, wall, mine uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me := g.Seats[1]
	djinn = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Erhnam Djinn", TypeLine: "Creature — Djinn",
		Power: 4, Toughness: 5, OracleID: erhnamDjinnOracle, Owner: me.ID, Controller: me.ID,
	})
	seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	mine = pushTribalCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	bear2 = pushTribalCreature(g, g.Seats[2].ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	bear3 = pushTribalCreature(g, g.Seats[3].ID, "Runeclaw Bear", "Creature — Bear", 2, 2)
	wall = pushTribalCreature(g, g.Seats[2].ID, "Wall of Wood", "Creature — Wall", 0, 3)
	advanceToUpkeepOf(t, g, 1)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("Erhnam Djinn's upkeep trigger asked for no target")
	}
	return g, djinn, bear2, bear3, wall, mine
}

func TestErhnamDjinnTargetsOnlyANonWallCreatureAnOpponentControls(t *testing.T) {
	g, _, _, _, wall, mine := erhnamTable(t)
	me := g.Seats[1]
	for name, id := range map[string]uuid.UUID{"a Wall": wall, "a creature its controller controls": mine} {
		c := g.Clone()
		p := latestPickTarget(c, me.ID)
		if err := c.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
			t.Errorf("Erhnam Djinn targeted %s", name)
		}
	}
}

// The grant lasts through every other seat's turn, so the creature walks
// past Erhnam Djinn's controller's Forest, and ends as that player's next
// upkeep begins: when the next trigger asks for its target, last turn's
// forestwalk is already gone.
func TestErhnamDjinnGrantsForestwalkUntilYourNextUpkeep(t *testing.T) {
	g, djinn, bear2, bear3, _, _ := erhnamTable(t)
	me := g.Seats[1]
	pickCard(t, g, me.ID, bear2)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Fatal("the target did not gain forestwalk")
	}
	if hasEffectiveKeyword(t, g, bear3, "forestwalk") {
		t.Fatal("a creature that was not the target gained forestwalk")
	}

	advanceToMainOf(t, g, 2)
	if !hasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Fatal("forestwalk ended before the bear's controller's turn")
	}
	attackIntoBlocks(t, g, me.ID, bear2)
	if r := blockRefusal(t, g, djinn, bear2); r != game.BlockReasonLandwalk {
		t.Fatalf("Erhnam Djinn blocking the forestwalker: reason %q, want landwalk", r)
	}

	advanceToMainOf(t, g, 0)
	if !hasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Fatal("forestwalk ended before Erhnam Djinn's controller's next upkeep")
	}
	advanceToUpkeepOf(t, g, 1)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("the next upkeep's trigger asked for no target")
	}
	if hasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Error("forestwalk outlived the beginning of the next upkeep")
	}
	pickCard(t, g, me.ID, bear3)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear3, "forestwalk") || hasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Error("the second trigger's grant is not on its own target alone")
	}
}

// A table holding the grant is a restore point, and the restored grant
// still ends at the right upkeep.
func TestErhnamDjinnsGrantSurvivesARestore(t *testing.T) {
	g, _, bear2, _, _, _ := erhnamTable(t)
	pickCard(t, g, g.Seats[1].ID, bear2)
	passPriorityAroundTable(t, g)
	advanceToMainOf(t, g, 3)

	g = restoreRoundTrip(t, g, true)
	if !hasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Fatal("the restore dropped the grant")
	}
	advanceToMainOf(t, g, 0)
	if !hasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Fatal("the restored grant ended early")
	}
	advanceToUpkeepOf(t, g, 1)
	if hasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Error("the restored grant outlived its controller's next upkeep")
	}
}

// CR 800.4m: if Erhnam Djinn's controller leaves the game, the grant on
// the other player's creature lasts until that player's next turn would
// have begun, and no longer.
func TestErhnamDjinnsGrantEndsWhenItsControllersTurnWouldHaveBegun(t *testing.T) {
	g, _, bear2, _, _, _ := erhnamTable(t)
	me := g.Seats[1]
	pickCard(t, g, me.ID, bear2)
	passPriorityAroundTable(t, g)
	advanceToMainOf(t, g, 2)
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if !hasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Fatal("the grant ended as its controller left; CR 800.4m says it waits")
	}
	advanceToMainOf(t, g, 0)
	if !hasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Fatal("the grant ended before its controller's turn would have begun")
	}
	advanceToMainOf(t, g, 2) // seat 1 is stepped over
	if hasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Error("the grant outlived the turn its departed controller would have taken")
	}
}

// Xenic Poltergeist: the artifact is a creature of its mana value until
// the Poltergeist's controller's next upkeep, and not a moment longer.
func TestXenicPoltergeistAnimatesAnArtifactUntilYourNextUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	ghost := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Xenic Poltergeist", TypeLine: "Creature — Spirit",
		Power: 1, Toughness: 1, OracleID: xenicPoltergeistOracle, Owner: me.ID, Controller: me.ID,
	})
	relic := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Relic", TypeLine: "Artifact", ManaCost: "{3}",
		Owner: them.ID, Controller: them.ID,
	})
	golem := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{4}",
		Power: 3, Toughness: 3, Owner: them.ID, Controller: them.ID,
	})
	if err := g.Clone().ActivateCatalogAbility(me.ID, ghost, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: golem}}}); err == nil {
		t.Fatal("the Poltergeist targeted an artifact creature")
	}
	pr7Activate(t, g, me.ID, ghost, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}}})
	check := func(when string, want bool) {
		t.Helper()
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		c := findBattlefieldCardForTest(g, relic)
		if c == nil {
			t.Fatalf("%s: the relic is gone", when)
		}
		eff := c.Effective()
		if c.IsCreature() != want || !c.IsArtifact() {
			t.Fatalf("%s: creature %v artifact %v, want creature %v", when, c.IsCreature(), c.IsArtifact(), want)
		}
		if want && (eff.Power != 3 || eff.Toughness != 3) {
			t.Errorf("%s: %d/%d, want 3/3 (its mana value)", when, eff.Power, eff.Toughness)
		}
	}
	check("as it resolves", true)
	advanceToMainOf(t, g, 1)
	check("on the artifact's controller's turn", true)
	advanceToMainOf(t, g, 3)
	check("on the last opponent's turn", true)
	advanceToUpkeepOf(t, g, 0)
	check("at the Poltergeist's controller's next upkeep", false)
}

func TestUntilYourNextUpkeepCardsAreFull(t *testing.T) {
	for name, oracle := range map[string]string{"Erhnam Djinn": erhnamDjinnOracle, "Xenic Poltergeist": xenicPoltergeistOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", name)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s completeness = %v, want Full", name, spec.Completeness)
		}
	}
}
