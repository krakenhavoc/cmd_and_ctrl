package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// first_life_loss_test.go — "whenever you lose life for the first time
// each turn" (#2540, Gonti's Machinations). The reading is the turn
// tally's LifeLost: the tally listener is registered ahead of the
// trigger harvester, so by the time a trigger's AppliesTo runs for a
// loss, PlayerTurnTally.LifeLost ALREADY includes that loss. A loss is
// therefore the first of the turn exactly when the tally equals the
// loss. The probe below pins that ordering before any card relies on it.

const gontisMachinationsOracle = "4e6238d8-bd6a-4952-a7dd-9c41a78a2cba"

func firstLossProbeReset() {
	effectstest.FirstLossProbeMu.Lock()
	effectstest.FirstLossProbeSamples = nil
	effectstest.FirstLossProbeMu.Unlock()
}

func firstLossProbeSeen() []effectstest.FirstLossSample {
	effectstest.FirstLossProbeMu.Lock()
	defer effectstest.FirstLossProbeMu.Unlock()
	return append([]effectstest.FirstLossSample(nil), effectstest.FirstLossProbeSamples...)
}

// A trigger's AppliesTo sees the tally AFTER the event that is being
// asked about, for both life-loss event kinds, and a later event in the
// same combat damage step sees the running total, not the step's total
// and not the total before it.
func TestATriggerSeesTheTallyIncludingTheLossItIsAskedAbout(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "First Life Loss Probe", "Enchantment", effectstest.FirstLossProbeOracle, false)
	firstLossProbeReset()

	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -3); err != nil {
			t.Fatal(err)
		}
	})
	src := effectstest.Pr7Creature(g, foe.ID, "Source", 2, "R")
	effectstest.Pr7Hit(t, g, src, me.ID, 4)

	want := []effectstest.FirstLossSample{
		{Kind: game.EventChangeLife, Lost: 3, Tally: 3},
		{Kind: game.EventDealDamage, Lost: 4, Tally: 7},
	}
	got := firstLossProbeSeen()
	if len(got) != len(want) {
		t.Fatalf("probe saw %d life-loss events, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("event %d: saw %+v, want %+v (the tally includes the loss being asked about)", i, got[i], want[i])
		}
	}
}

// Two creatures dealing combat damage to the same player in one step
// are two events; the second one sees both losses in the tally.
func TestSimultaneousCombatDamageIsOneEventPerSourceWithARunningTally(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, foe.ID, "First Life Loss Probe", "Enchantment", effectstest.FirstLossProbeOracle, false)
	a := effectstest.Pr7Creature(g, me.ID, "Attacker A", 2, "R")
	b := effectstest.Pr7Creature(g, me.ID, "Attacker B", 3, "R")
	firstLossProbeReset()

	effectstest.DeclareAttack(t, g, foe.ID, a, b)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)

	got := firstLossProbeSeen()
	if len(got) != 2 {
		t.Fatalf("probe saw %d life-loss events from one combat damage step, want 2: %+v", len(got), got)
	}
	if got[0].Tally != got[0].Lost || got[1].Tally != got[0].Lost+got[1].Lost {
		t.Errorf("tallies %+v: the first event should see only itself and the second both", got)
	}
}

// --- the card -----------------------------------------------------------

func gontiBoard(t *testing.T) (g *game.Game, me, foe *game.Player, gonti uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, foe = g.Seats[0], g.Seats[1]
	// Gonti's controller is the one losing life, so it sits across from
	// the creatures.
	gonti = effectstest.PushCatalogPermanent(g, foe.ID, "Gonti's Machinations", "Enchantment", gontisMachinationsOracle, false)
	return g, me, foe, gonti
}

func TestGontisMachinationsDeclaresItsAbilities(t *testing.T) {
	spec, ok := effects.Lookup(gontisMachinationsOracle)
	if !ok {
		t.Fatal("Gonti's Machinations is not registered")
	}
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("completeness %v, want Full", spec.Completeness)
	}
	if len(spec.Triggered) != 1 || len(spec.Activated) != 1 {
		t.Errorf("%d triggered and %d activated rows, want 1 and 1", len(spec.Triggered), len(spec.Activated))
	}
}

// Simultaneous combat damage from several creatures to the same player
// is several events and one first loss: one energy.
func TestGontisMachinationsSimultaneousCombatDamageTriggersOnce(t *testing.T) {
	g, me, foe, _ := gontiBoard(t)
	a := effectstest.Pr7Creature(g, me.ID, "Attacker A", 2, "R")
	b := effectstest.Pr7Creature(g, me.ID, "Attacker B", 3, "R")
	c := effectstest.Pr7Creature(g, me.ID, "Attacker C", 1, "R")

	effectstest.DeclareAttack(t, g, foe.ID, a, b, c)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)

	if foe.Life != game.StartingLife-6 {
		t.Fatalf("life %d, want %d: all three creatures connected", foe.Life, game.StartingLife-6)
	}
	if got := effectstest.EnergyOf(foe); got != 1 {
		t.Errorf("energy %d after three simultaneous hits, want 1", got)
	}
}

// A second loss later in the same turn does not trigger again, and the
// next turn's first loss does.
func TestGontisMachinationsSecondLossSameTurnDoesNotRetriggerButNextTurnDoes(t *testing.T) {
	g, me, foe, _ := gontiBoard(t)
	src := effectstest.Pr7Creature(g, me.ID, "Source", 2, "R")

	effectstest.Pr7Hit(t, g, src, foe.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EnergyOf(foe); got != 1 {
		t.Fatalf("energy %d after the first loss, want 1", got)
	}

	effectstest.Pr7Hit(t, g, src, foe.ID, 3)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, foe.ID, -1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EnergyOf(foe); got != 1 {
		t.Errorf("energy %d after more losses the same turn, want 1", got)
	}

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.Pr7Hit(t, g, src, foe.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EnergyOf(foe); got != 2 {
		t.Errorf("energy %d after the next turn's first loss, want 2", got)
	}
}

// Gaining life changes nothing: it is not a loss, so it neither
// triggers nor uses up the turn's first loss.
func TestGontisMachinationsLifeGainInBetweenChangesNothing(t *testing.T) {
	g, me, foe, _ := gontiBoard(t)
	src := effectstest.Pr7Creature(g, me.ID, "Source", 2, "R")

	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, foe.ID, 5); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EnergyOf(foe); got != 0 {
		t.Fatalf("energy %d after gaining life, want 0", got)
	}

	effectstest.Pr7Hit(t, g, src, foe.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EnergyOf(foe); got != 1 {
		t.Fatalf("energy %d after the first loss following a gain, want 1", got)
	}

	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, foe.ID, 4); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.Pr7Hit(t, g, src, foe.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EnergyOf(foe); got != 1 {
		t.Errorf("energy %d: a second loss after a gain is not a first loss, want 1", got)
	}
}

// Damage that costs no life is not a loss: infect and prevented damage
// leave the turn's first loss for a later, real one.
func TestGontisMachinationsDamageThatLosesNoLifeDoesNotCount(t *testing.T) {
	g, me, foe, _ := gontiBoard(t)
	agent := effectstest.B12Push(g, me.ID, "Blighted Agent", "Creature — Phyrexian Human Rogue", effectstest.BlightedAgentOracle, 1, 1)
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")

	effectstest.Pr7Hit(t, g, agent, foe.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PoisonOn(foe) != 3 || foe.Life != game.StartingLife || effectstest.EnergyOf(foe) != 0 {
		t.Fatalf("infect: %d poison, %d life, %d energy; want 3, %d and 0", effectstest.PoisonOn(foe), foe.Life, effectstest.EnergyOf(foe), game.StartingLife)
	}

	g.WithWriteLock(func() {
		if !g.PreventNextDamageThisTurnForEffect(uuid.Nil, foe.ID, 2, false, "test shield") {
			t.Fatal("the shield registered nothing")
		}
	})
	effectstest.Pr7Hit(t, g, bear, foe.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if foe.Life != game.StartingLife || effectstest.EnergyOf(foe) != 0 {
		t.Fatalf("prevented: %d life, %d energy; want %d and 0", foe.Life, effectstest.EnergyOf(foe), game.StartingLife)
	}

	effectstest.Pr7Hit(t, g, bear, foe.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if foe.Life != game.StartingLife-2 {
		t.Fatalf("life %d, want %d: the shield is spent", foe.Life, game.StartingLife-2)
	}
	if got := effectstest.EnergyOf(foe); got != 1 {
		t.Errorf("energy %d after the first real loss, want 1", got)
	}
}

// Only its controller's own loss counts.
func TestGontisMachinationsIgnoresOtherPlayersLosses(t *testing.T) {
	g, me, foe, _ := gontiBoard(t)
	src := effectstest.Pr7Creature(g, me.ID, "Source", 2, "R")

	effectstest.Pr7Hit(t, g, src, g.Seats[2].ID, 2)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -2); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EnergyOf(foe); got != 0 {
		t.Errorf("energy %d for other players' losses, want 0", got)
	}
}

// Pay {E}{E}, Sacrifice: each opponent loses 3 life and you gain life
// equal to the life lost this way.
func TestGontisMachinationsDrainsEachOpponentForThreeAndGainsTheTotal(t *testing.T) {
	g, me, foe, gonti := gontiBoard(t)
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(foe.ID, game.CounterEnergy, 2); err != nil {
			t.Fatal(err)
		}
	})
	before := foe.Life

	effectstest.Pr7Activate(t, g, foe.ID, gonti, 0, game.ActivateAbilityParams{})

	if effectstest.FindBattlefieldCardForTest(g, gonti) != nil {
		t.Error("Gonti's Machinations is still on the battlefield after being sacrificed")
	}
	if got := effectstest.EnergyOf(foe); got != 0 {
		t.Errorf("energy %d, want 0: the cost paid {E}{E}", got)
	}
	for _, opp := range []*game.Player{me, g.Seats[2], g.Seats[3]} {
		if opp.Life != game.StartingLife-3 {
			t.Errorf("%s life %d, want %d", opp.Name, opp.Life, game.StartingLife-3)
		}
	}
	if foe.Life != before+9 {
		t.Errorf("controller life %d, want %d: three opponents lost 3 each", foe.Life, before+9)
	}
}

func TestGontisMachinationsDrainNeedsTheEnergy(t *testing.T) {
	g, _, foe, gonti := gontiBoard(t)
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(foe.ID, game.CounterEnergy, 1); err != nil {
			t.Fatal(err)
		}
	})
	if err := g.ActivateCatalogAbility(foe.ID, gonti, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activating with one {E} succeeded")
	}
}
