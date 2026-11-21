package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsex "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ex"
	cardshu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/hu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fallen Empires' activation-count mana abilities (#5745): "{1}: Add {W}
// (or {B}). If this ability has been activated four or more times this
// turn, sacrifice this creature at the beginning of the next end step."

const (
	femFarrelitePriestOracle = "3ee18e5f-3d12-489e-b337-29208caaed7a"
	femInitiatesOracle       = "0bf359ef-669f-41dc-8ab5-ba8e58b2f398"
)

func TestFemActivationCountCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		femFarrelitePriestOracle: "Farrelite Priest",
		femInitiatesOracle:       "Initiates of the Ebon Hand",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// femActivate activates the creature's mana ability once, paying {1} from
// a lone floating {G}, and checks it made one mana of `color`. The pool is
// emptied around it, so each activation pays with the {G} and no mana an
// earlier one made.
func femActivate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, color string) {
	t.Helper()
	g.WithWriteLock(func() {
		me.ManaPool.EmptyPool()
		effectstest.FillPoolColored(me, "G", 1)
	})
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{1}: add {%s}: %v", color, err)
	}
	if got := effectstest.McPoolCount(me, color); got != 1 {
		t.Fatalf("{%s} in the pool %d after an activation, want 1", color, got)
	}
	if got := effectstest.McPoolCount(me, "G"); got != 0 {
		t.Fatalf("{G} in the pool %d after an activation, want 0: the {1} was not paid", got)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
}

// femOnlyNamed is the instance ID of the one battlefield card named
// `name`: a flicker returns the card as a new object with a new ID.
func femOnlyNamed(t *testing.T, g *game.Game, name string) uuid.UUID {
	t.Helper()
	var ids []uuid.UUID
	g.ReadSnapshot(func() { ids = effectstest.BattlefieldIDsNamed(g, name) })
	if len(ids) != 1 {
		t.Fatalf("%d %s on the battlefield, want 1", len(ids), name)
	}
	return ids[0]
}

// femActivations is the per-turn count of the ability labelled `label`
// on the object `id` names.
func femActivations(g *game.Game, id uuid.UUID, label string) int {
	n := 0
	g.ReadSnapshot(func() { n = g.ActivatedThisTurn(id, label) })
	return n
}

// femSacrificeTriggers is how many delayed triggers from `id` are waiting.
func femSacrificeTriggers(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, dt := range g.DelayedTriggers {
			if dt != nil && dt.SourceCardID == id {
				n++
			}
		}
	})
	return n
}

// femOnBattlefield reports whether `id` is on the battlefield.
func femOnBattlefield(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				found = true
			}
		}
	})
	return found
}

func TestFemFarrelitePriestIsSacrificedAfterTheFourthActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	priest := effectstest.B12Push(g, me.ID, "Farrelite Priest", "Creature — Human Cleric", femFarrelitePriestOracle, 1, 3)
	effectstest.AdvanceToMain(t, g)
	for i := 1; i <= 3; i++ {
		femActivate(t, g, me, priest, "W")
		if got := femSacrificeTriggers(g, priest); got != 0 {
			t.Fatalf("after activation %d, %d sacrifice triggers are waiting; want none before the fourth", i, got)
		}
	}
	if got := femActivations(g, priest, cardsex.FarrelitePriestManaLabel); got != 3 {
		t.Fatalf("the ability counted %d activations, want 3", got)
	}
	// The fourth still makes its mana, and sets up the sacrifice.
	femActivate(t, g, me, priest, "W")
	if got := femSacrificeTriggers(g, priest); got != 1 {
		t.Fatalf("after the fourth activation %d sacrifice triggers are waiting, want 1", got)
	}
	// "Four or more": the fifth sets up another, and only one sacrifice
	// can happen.
	femActivate(t, g, me, priest, "W")
	if got := femSacrificeTriggers(g, priest); got != 2 {
		t.Fatalf("after the fifth activation %d sacrifice triggers are waiting, want 2", got)
	}
	if !femOnBattlefield(g, priest) {
		t.Fatal("the Priest left before the end step")
	}
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if femOnBattlefield(g, priest) {
		t.Fatal("the Priest survived the end step after four activations")
	}
	if !effectstest.InGraveyard(me, priest) {
		t.Error("the Priest is not in its owner's graveyard")
	}
	if got := femSacrificeTriggers(g, priest); got != 0 {
		t.Errorf("%d sacrifice triggers still waiting after the end step", got)
	}
}

func TestFemInitiatesCountResetsEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	initiates := effectstest.B12Push(g, me.ID, "Initiates of the Ebon Hand", "Creature — Cleric", femInitiatesOracle, 1, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	for i := 0; i < 3; i++ {
		femActivate(t, g, me, initiates, "B")
	}
	// The next turn is a new count: one more activation is its first.
	effectstest.AdvanceToMainOf(t, g, 1)
	if got := femActivations(g, initiates, cardshu.InitiatesOfTheEbonHandManaLabel); got != 0 {
		t.Fatalf("the count is %d at the start of the next turn, want 0", got)
	}
	femActivate(t, g, me, initiates, "B")
	if got := femSacrificeTriggers(g, initiates); got != 0 {
		t.Fatalf("%d sacrifice triggers after three activations last turn and one this turn, want none", got)
	}
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !femOnBattlefield(g, initiates) {
		t.Fatal("the Initiates was sacrificed without four activations in one turn")
	}
}

// CR 400.7: a flickered creature is a new object, so its count starts
// again, and a sacrifice the old object set up does not reach it
// (CR 603.7c).
func TestFemInitiatesFlickerIsANewObject(t *testing.T) {
	const name = "Initiates of the Ebon Hand"
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	first := effectstest.B12Push(g, me.ID, name, "Creature — Cleric", femInitiatesOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		femActivate(t, g, me, first, "B")
	}
	effectstest.FlickerForTest(t, g, me.ID, first)
	second := femOnlyNamed(t, g, name)
	if got := femActivations(g, second, cardshu.InitiatesOfTheEbonHandManaLabel); got != 0 {
		t.Fatalf("the returned Initiates starts with %d activations, want 0", got)
	}
	femActivate(t, g, me, second, "B")
	if got := femSacrificeTriggers(g, second); got != 0 {
		t.Fatalf("the new object's first activation set up %d sacrifice triggers, want none", got)
	}
	for i := 0; i < 3; i++ {
		femActivate(t, g, me, second, "B")
	}
	if got := femSacrificeTriggers(g, second); got != 1 {
		t.Fatalf("the new object's fourth activation set up %d sacrifice triggers, want 1", got)
	}
	// Flickered again before the end step: the waiting trigger names the
	// object that was, and the creature that came back is not it.
	effectstest.FlickerForTest(t, g, me.ID, second)
	third := femOnlyNamed(t, g, name)
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if !femOnBattlefield(g, third) {
		t.Fatal("a sacrifice set up by the old object reached the flickered Initiates")
	}
}

// The auto-tapper does not spend a creature whose rider can sacrifice it:
// a {W} cost with a Sol Ring and the Priest on the board finds no plan,
// and nothing is counted.
func TestFemFarrelitePriestIsNotAutoTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	priest := effectstest.B12Push(g, me.ID, "Farrelite Priest", "Creature — Human Cleric", femFarrelitePriestOracle, 1, 3)
	effectstest.B12Push(g, me.ID, "Sol Ring", "Artifact", effectstest.SolRingOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if plan, ok := effectstest.PlansFor(t, g, me, "{W}"); ok {
		t.Fatalf("the auto-tapper planned %v for {W}; the Priest is activated by hand", plan)
	}
	if got := femActivations(g, priest, cardsex.FarrelitePriestManaLabel); got != 0 {
		t.Errorf("the Priest's ability counted %d activations with none made", got)
	}
}
