package effects

import (
	"sync"
	"testing"

	"github.com/google/uuid"

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

// --- the ordering probe ---------------------------------------------

const firstLossProbeOracle = "test-first-life-loss-probe"

// firstLossSample is what a trigger's AppliesTo saw for one life-loss
// event: the loss itself and the controller's tally at that moment.
type firstLossSample struct {
	Kind  game.EventKind
	Lost  int
	Tally int
}

var (
	firstLossProbeMu      sync.Mutex
	firstLossProbeSamples []firstLossSample
)

func init() {
	Register(Spec{
		OracleID: firstLossProbeOracle,
		Name:     "First Life Loss Probe",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventChangeLife, game.EventDealDamage},
			Key:     "First Life Loss Probe",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if lost, ok := s22PlayerLostLife(ev, source.Controller, g); ok {
					firstLossProbeMu.Lock()
					firstLossProbeSamples = append(firstLossProbeSamples, firstLossSample{
						Kind: ev.Kind, Lost: lost, Tally: g.TurnTallyFor(source.Controller).LifeLost,
					})
					firstLossProbeMu.Unlock()
				}
				return false
			},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
}

func firstLossProbeReset() {
	firstLossProbeMu.Lock()
	firstLossProbeSamples = nil
	firstLossProbeMu.Unlock()
}

func firstLossProbeSeen() []firstLossSample {
	firstLossProbeMu.Lock()
	defer firstLossProbeMu.Unlock()
	return append([]firstLossSample(nil), firstLossProbeSamples...)
}

// A trigger's AppliesTo sees the tally AFTER the event that is being
// asked about, for both life-loss event kinds, and a later event in the
// same combat damage step sees the running total, not the step's total
// and not the total before it.
func TestATriggerSeesTheTallyIncludingTheLossItIsAskedAbout(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "First Life Loss Probe", "Enchantment", firstLossProbeOracle, false)
	firstLossProbeReset()

	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -3); err != nil {
			t.Fatal(err)
		}
	})
	src := pr7Creature(g, foe.ID, "Source", 2, "R")
	pr7Hit(t, g, src, me.ID, 4)

	want := []firstLossSample{
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
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, foe.ID, "First Life Loss Probe", "Enchantment", firstLossProbeOracle, false)
	a := pr7Creature(g, me.ID, "Attacker A", 2, "R")
	b := pr7Creature(g, me.ID, "Attacker B", 3, "R")
	firstLossProbeReset()

	declareAttack(t, g, foe.ID, a, b)
	advanceTo(t, g, game.StepCombatDamage)

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
	g = newCatalogGame(t)
	me, foe = g.Seats[0], g.Seats[1]
	// Gonti's controller is the one losing life, so it sits across from
	// the creatures.
	gonti = pushCatalogPermanent(g, foe.ID, "Gonti's Machinations", "Enchantment", gontisMachinationsOracle, false)
	return g, me, foe, gonti
}

func TestGontisMachinationsDeclaresItsAbilities(t *testing.T) {
	spec, ok := Lookup(gontisMachinationsOracle)
	if !ok {
		t.Fatal("Gonti's Machinations is not registered")
	}
	if spec.Completeness != CompletenessFull {
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
	a := pr7Creature(g, me.ID, "Attacker A", 2, "R")
	b := pr7Creature(g, me.ID, "Attacker B", 3, "R")
	c := pr7Creature(g, me.ID, "Attacker C", 1, "R")

	declareAttack(t, g, foe.ID, a, b, c)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	if foe.Life != game.StartingLife-6 {
		t.Fatalf("life %d, want %d: all three creatures connected", foe.Life, game.StartingLife-6)
	}
	if got := energyOf(foe); got != 1 {
		t.Errorf("energy %d after three simultaneous hits, want 1", got)
	}
}

// A second loss later in the same turn does not trigger again, and the
// next turn's first loss does.
func TestGontisMachinationsSecondLossSameTurnDoesNotRetriggerButNextTurnDoes(t *testing.T) {
	g, me, foe, _ := gontiBoard(t)
	src := pr7Creature(g, me.ID, "Source", 2, "R")

	pr7Hit(t, g, src, foe.ID, 2)
	passPriorityAroundTable(t, g)
	if got := energyOf(foe); got != 1 {
		t.Fatalf("energy %d after the first loss, want 1", got)
	}

	pr7Hit(t, g, src, foe.ID, 3)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, foe.ID, -1); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := energyOf(foe); got != 1 {
		t.Errorf("energy %d after more losses the same turn, want 1", got)
	}

	advanceToUpkeepOf(t, g, 1)
	pr7Hit(t, g, src, foe.ID, 1)
	passPriorityAroundTable(t, g)
	if got := energyOf(foe); got != 2 {
		t.Errorf("energy %d after the next turn's first loss, want 2", got)
	}
}

// Gaining life changes nothing: it is not a loss, so it neither
// triggers nor uses up the turn's first loss.
func TestGontisMachinationsLifeGainInBetweenChangesNothing(t *testing.T) {
	g, me, foe, _ := gontiBoard(t)
	src := pr7Creature(g, me.ID, "Source", 2, "R")

	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, foe.ID, 5); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := energyOf(foe); got != 0 {
		t.Fatalf("energy %d after gaining life, want 0", got)
	}

	pr7Hit(t, g, src, foe.ID, 2)
	passPriorityAroundTable(t, g)
	if got := energyOf(foe); got != 1 {
		t.Fatalf("energy %d after the first loss following a gain, want 1", got)
	}

	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, foe.ID, 4); err != nil {
			t.Fatal(err)
		}
	})
	pr7Hit(t, g, src, foe.ID, 2)
	passPriorityAroundTable(t, g)
	if got := energyOf(foe); got != 1 {
		t.Errorf("energy %d: a second loss after a gain is not a first loss, want 1", got)
	}
}

// Damage that costs no life is not a loss: infect and prevented damage
// leave the turn's first loss for a later, real one.
func TestGontisMachinationsDamageThatLosesNoLifeDoesNotCount(t *testing.T) {
	g, me, foe, _ := gontiBoard(t)
	agent := b12Push(g, me.ID, "Blighted Agent", "Creature — Phyrexian Human Rogue", blightedAgentOracle, 1, 1)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")

	pr7Hit(t, g, agent, foe.ID, 3)
	passPriorityAroundTable(t, g)
	if poisonOn(foe) != 3 || foe.Life != game.StartingLife || energyOf(foe) != 0 {
		t.Fatalf("infect: %d poison, %d life, %d energy; want 3, %d and 0", poisonOn(foe), foe.Life, energyOf(foe), game.StartingLife)
	}

	g.WithWriteLock(func() {
		if !g.PreventNextDamageThisTurnForEffect(uuid.Nil, foe.ID, 2, false, "test shield") {
			t.Fatal("the shield registered nothing")
		}
	})
	pr7Hit(t, g, bear, foe.ID, 2)
	passPriorityAroundTable(t, g)
	if foe.Life != game.StartingLife || energyOf(foe) != 0 {
		t.Fatalf("prevented: %d life, %d energy; want %d and 0", foe.Life, energyOf(foe), game.StartingLife)
	}

	pr7Hit(t, g, bear, foe.ID, 2)
	passPriorityAroundTable(t, g)
	if foe.Life != game.StartingLife-2 {
		t.Fatalf("life %d, want %d: the shield is spent", foe.Life, game.StartingLife-2)
	}
	if got := energyOf(foe); got != 1 {
		t.Errorf("energy %d after the first real loss, want 1", got)
	}
}

// Only its controller's own loss counts.
func TestGontisMachinationsIgnoresOtherPlayersLosses(t *testing.T) {
	g, me, foe, _ := gontiBoard(t)
	src := pr7Creature(g, me.ID, "Source", 2, "R")

	pr7Hit(t, g, src, g.Seats[2].ID, 2)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -2); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := energyOf(foe); got != 0 {
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

	pr7Activate(t, g, foe.ID, gonti, 0, game.ActivateAbilityParams{})

	if findBattlefieldCardForTest(g, gonti) != nil {
		t.Error("Gonti's Machinations is still on the battlefield after being sacrificed")
	}
	if got := energyOf(foe); got != 0 {
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
