package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// erhnamTable puts Erhnam Djinn and a Forest under seat 1, a creature
// under each of seats 2 and 3, a Wall under seat 2 and another creature
// under seat 1, and walks to seat 1's upkeep with the trigger asking for
// its target.
func erhnamTable(t *testing.T) (g *game.Game, djinn, bear2, bear3, wall, mine uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	djinn = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Erhnam Djinn", TypeLine: "Creature — Djinn",
		Power: 4, Toughness: 5, OracleID: effectstest.ErhnamDjinnOracle, Owner: me.ID, Controller: me.ID,
	})
	effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	mine = effectstest.PushTribalCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	bear2 = effectstest.PushTribalCreature(g, g.Seats[2].ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	bear3 = effectstest.PushTribalCreature(g, g.Seats[3].ID, "Runeclaw Bear", "Creature — Bear", 2, 2)
	wall = effectstest.PushTribalCreature(g, g.Seats[2].ID, "Wall of Wood", "Creature — Wall", 0, 3)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("Erhnam Djinn's upkeep trigger asked for no target")
	}
	return g, djinn, bear2, bear3, wall, mine
}

func TestErhnamDjinnTargetsOnlyANonWallCreatureAnOpponentControls(t *testing.T) {
	g, _, _, _, wall, mine := erhnamTable(t)
	me := g.Seats[1]
	for name, id := range map[string]uuid.UUID{"a Wall": wall, "a creature its controller controls": mine} {
		c := g.Clone()
		p := effectstest.LatestPickTarget(c, me.ID)
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
	effectstest.PickCard(t, g, me.ID, bear2)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Fatal("the target did not gain forestwalk")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear3, "forestwalk") {
		t.Fatal("a creature that was not the target gained forestwalk")
	}

	effectstest.AdvanceToMainOf(t, g, 2)
	if !effectstest.HasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Fatal("forestwalk ended before the bear's controller's turn")
	}
	effectstest.AttackIntoBlocks(t, g, me.ID, bear2)
	if r := effectstest.BlockRefusal(t, g, djinn, bear2); r != game.BlockReasonLandwalk {
		t.Fatalf("Erhnam Djinn blocking the forestwalker: reason %q, want landwalk", r)
	}

	effectstest.AdvanceToMainOf(t, g, 0)
	if !effectstest.HasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Fatal("forestwalk ended before Erhnam Djinn's controller's next upkeep")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("the next upkeep's trigger asked for no target")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Error("forestwalk outlived the beginning of the next upkeep")
	}
	effectstest.PickCard(t, g, me.ID, bear3)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear3, "forestwalk") || effectstest.HasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Error("the second trigger's grant is not on its own target alone")
	}
}

// A table holding the grant is a restore point, and the restored grant
// still ends at the right upkeep.
func TestErhnamDjinnsGrantSurvivesARestore(t *testing.T) {
	g, _, bear2, _, _, _ := erhnamTable(t)
	effectstest.PickCard(t, g, g.Seats[1].ID, bear2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToMainOf(t, g, 3)

	g = effectstest.RestoreRoundTrip(t, g, true)
	if !effectstest.HasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Fatal("the restore dropped the grant")
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	if !effectstest.HasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Fatal("the restored grant ended early")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Error("the restored grant outlived its controller's next upkeep")
	}
}

// CR 800.4m: if Erhnam Djinn's controller leaves the game, the grant on
// the other player's creature lasts until that player's next turn would
// have begun, and no longer.
func TestErhnamDjinnsGrantEndsWhenItsControllersTurnWouldHaveBegun(t *testing.T) {
	g, _, bear2, _, _, _ := erhnamTable(t)
	me := g.Seats[1]
	effectstest.PickCard(t, g, me.ID, bear2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToMainOf(t, g, 2)
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Fatal("the grant ended as its controller left; CR 800.4m says it waits")
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	if !effectstest.HasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Fatal("the grant ended before its controller's turn would have begun")
	}
	effectstest.AdvanceToMainOf(t, g, 2) // seat 1 is stepped over
	if effectstest.HasEffectiveKeyword(t, g, bear2, "forestwalk") {
		t.Error("the grant outlived the turn its departed controller would have taken")
	}
}

// Xenic Poltergeist: the artifact is a creature of its mana value until
// the Poltergeist's controller's next upkeep, and not a moment longer.
func TestXenicPoltergeistAnimatesAnArtifactUntilYourNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	ghost := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Xenic Poltergeist", TypeLine: "Creature — Spirit",
		Power: 1, Toughness: 1, OracleID: effectstest.XenicPoltergeistOracle, Owner: me.ID, Controller: me.ID,
	})
	relic := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Relic", TypeLine: "Artifact", ManaCost: "{3}",
		Owner: them.ID, Controller: them.ID,
	})
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{4}",
		Power: 3, Toughness: 3, Owner: them.ID, Controller: them.ID,
	})
	if err := g.Clone().ActivateCatalogAbility(me.ID, ghost, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: golem}}}); err == nil {
		t.Fatal("the Poltergeist targeted an artifact creature")
	}
	effectstest.Pr7Activate(t, g, me.ID, ghost, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}}})
	check := func(when string, want bool) {
		t.Helper()
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		c := effectstest.FindBattlefieldCardForTest(g, relic)
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
	effectstest.AdvanceToMainOf(t, g, 1)
	check("on the artifact's controller's turn", true)
	effectstest.AdvanceToMainOf(t, g, 3)
	check("on the last opponent's turn", true)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	check("at the Poltergeist's controller's next upkeep", false)
}

func TestUntilYourNextUpkeepCardsAreFull(t *testing.T) {
	for name, oracle := range map[string]string{"Erhnam Djinn": effectstest.ErhnamDjinnOracle, "Xenic Poltergeist": effectstest.XenicPoltergeistOracle} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", name)
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s completeness = %v, want Full", name, spec.Completeness)
		}
	}
}
