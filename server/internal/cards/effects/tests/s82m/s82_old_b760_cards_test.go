package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 760 (Tempest, #5628): Energizer, Essence Bottle, Fevered
// Convulsions, Magnetic Web, Magmasaur, Spike Drone and Segmented Wurm.

const (
	oldB760Energizer = "708b0321-7de9-4ffd-9d7a-9f4813dd542e"
	oldB760Bottle    = "76f839b0-1d4c-4ad9-b44f-a8a13d6d109f"
	oldB760Fevered   = "83f35293-39e9-4c78-9c9f-25a1fc905676"
	oldB760Web       = "c62c8006-5650-43cb-8247-da097aa42376"
	oldB760Magmasaur = "da38bfdf-b4c8-4ffe-ad53-b42c317cae5f"
	oldB760Drone     = "dd6d15c0-c294-42c9-a6cf-57b0b5d40823"
	oldB760Wurm      = "eb1baf72-4ef6-4967-8a22-2022c633f3d3"
)

func oldB760AddCounter(t *testing.T, g *game.Game, id uuid.UUID, kind string, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, kind, n); err != nil {
			t.Fatal(err)
		}
	})
}

func TestOldB760EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB760Energizer: "Energizer", oldB760Bottle: "Essence Bottle", oldB760Fevered: "Fevered Convulsions",
		oldB760Web: "Magnetic Web", oldB760Magmasaur: "Magmasaur", oldB760Drone: "Spike Drone", oldB760Wurm: "Segmented Wurm",
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

func TestOldB760EnergizerGrowsByTapping(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Energizer", "Artifact Creature — Juggernaut", oldB760Energizer, false)
	effectstest.FillPool(me, 2)
	if err := effectstest.InvB1Activate(g, me.ID, id, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, id) != 1 || !effectstest.TappedNow(g, id) {
		t.Errorf("counters %d, tapped %v; want one counter and tapped", effectstest.PlusOnes(g, id), effectstest.TappedNow(g, id))
	}
	effectstest.FillPool(me, 2)
	if err := effectstest.InvB1Activate(g, me.ID, id, 0); err == nil {
		t.Error("activated while tapped")
	}
}

func TestOldB760EssenceBottleStoresAndCashesInElixir(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Essence Bottle", "Artifact", oldB760Bottle, false)
	for i := 0; i < 3; i++ {
		effectstest.FillPool(me, 3)
		if err := effectstest.InvB1Activate(g, me.ID, id, 0); err != nil {
			t.Fatalf("add %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		g.WithWriteLock(func() {
			for j := range g.Battlefield.Cards {
				if g.Battlefield.Cards[j].InstanceID == id {
					g.Battlefield.Cards[j].Tapped = false
				}
			}
		})
	}
	if got := effectstest.CounterCount(g, id, "elixir"); got != 3 {
		t.Fatalf("elixir counters %d, want 3", got)
	}
	life := me.Life
	if err := effectstest.InvB1Activate(g, me.ID, id, 1); err != nil {
		t.Fatalf("cash in: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d → %d, want +6", life, me.Life)
	}
	if got := effectstest.CounterCount(g, id, "elixir"); got != 0 {
		t.Errorf("elixir counters left %d", got)
	}
}

func TestOldB760FeveredConvulsionsPutsAMinusCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Fevered Convulsions", "Enchantment", oldB760Fevered, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FillPoolColored(me, "B", 4)
	if err := effectstest.InvB1Activate(g, me.ID, id, 0, effectstest.CardTarget(bear)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters %d, want 1", got)
	}
	if got := effectstest.B12Card(t, g, bear).CurrentPower(); got != 1 {
		t.Errorf("power %d, want 1", got)
	}
}

func TestOldB760SpikeDroneEntersWithACounterAndMovesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drone := effectstest.CastCatalogSpell(t, g, "Spike Drone", "Creature — Spike Drone", oldB760Drone, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, drone); got != 1 {
		t.Fatalf("the Drone entered with %d counters, want 1", got)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FillPool(me, 2)
	if err := effectstest.InvB1Activate(g, me.ID, drone, 0, effectstest.CardTarget(bear)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, bear); got != 1 {
		t.Errorf("the target has %d counters, want 1", got)
	}
}

func TestOldB760SegmentedWurmShrinksEachTimeItIsTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wurm := effectstest.B12Push(g, opp.ID, "Segmented Wurm", "Creature — Wurm", oldB760Wurm, 5, 5)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: wurm}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, wurm, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters %d, want 1", got)
	}
}

func oldB760AnswerMagmasaur(t *testing.T, g *game.Game, me *game.Player, yes bool) {
	t.Helper()
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID}) })
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestConfirmFor(g, me.ID)
	if c == nil {
		t.Fatal("no confirm prompt")
	}
	if err := g.ResolveConfirm(c.ID, me.ID, yes); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB760MagmasaurShrinksThenBlowsUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mag := effectstest.CastCatalogSpell(t, g, "Magmasaur", "Creature — Elemental Dinosaur", oldB760Magmasaur, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, mag); got != 5 {
		t.Fatalf("entered with %d counters, want 5", got)
	}
	bear := effectstest.B12Creature(g, opp.ID, "Ground Bear", "Creature — Bear", 2, 9)
	flier := effectstest.B12Push(g, opp.ID, "Flier", "Creature — Bird", "", 2, 9)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flier {
				g.Battlefield.Cards[i].Keywords = []string{"flying"}
			}
		}
	})
	oldB760AnswerMagmasaur(t, g, me, true)
	if got := effectstest.PlusOnes(g, mag); got != 4 {
		t.Fatalf("after removing one: %d counters, want 4", got)
	}
	mine, theirs := me.Life, opp.Life
	oldB760AnswerMagmasaur(t, g, me, false)
	effectstest.SettleStateBasedActions(t, g)
	if _, ok := effectstest.CardOnBattlefield(g, mag); ok {
		t.Error("the Magmasaur was not sacrificed")
	}
	if me.Life != mine-4 || opp.Life != theirs-4 {
		t.Errorf("life me %d → %d, opp %d → %d; want 4 damage each", mine, me.Life, theirs, opp.Life)
	}
	if got := effectstest.DamageMarkedOn(g, bear); got != 4 {
		t.Errorf("the ground creature has %d damage, want 4", got)
	}
	if got := effectstest.DamageMarkedOn(g, flier); got != 0 {
		t.Errorf("the flier has %d damage, want 0", got)
	}
}

func oldB760WebTable(t *testing.T) (g *game.Game, me, opp *game.Player, a, b, plain, c uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me = g.Seats[seat]
	opp = g.Seats[(seat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Magnetic Web", "Artifact", oldB760Web, false)
	a = effectstest.B12Creature(g, me.ID, "Magnet A", "Creature — Bear", 2, 2)
	b = effectstest.B12Creature(g, me.ID, "Magnet B", "Creature — Bear", 2, 2)
	plain = effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	c = effectstest.B12Creature(g, opp.ID, "Magnet C", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{a, b, c} {
		oldB760AddCounter(t, g, id, "magnet", 1)
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	return g, me, opp, a, b, plain, c
}

func TestOldB760MagneticWebForcesEveryMagnetCreatureToAttackWithOne(t *testing.T) {
	g, _, opp, a, b, plain, _ := oldB760WebTable(t)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != b {
		t.Fatalf("pass with one magnet creature attacking = %v, want a refusal naming the other", err)
	}
	if err := g.DeclareAttacker(b, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(B): %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("pass with both magnet creatures attacking (the plain one home %v): %v", plain, err)
	}
}

func TestOldB760MagneticWebAsksNothingWithoutAnAttackingMagnet(t *testing.T) {
	g, _, opp, _, _, plain, _ := oldB760WebTable(t)
	if err := g.DeclareAttacker(plain, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("a creature with no magnet counter attacked alone: %v", err)
	}
}

func TestOldB760MagneticWebAMagnetAloneIsFreeWhenNoOtherCanAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Magnetic Web", "Artifact", oldB760Web, false)
	a := effectstest.B12Creature(g, me.ID, "Magnet A", "Creature — Bear", 2, 2)
	oldB760AddCounter(t, g, a, "magnet", 1)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("a lone magnet creature may attack alone: %v", err)
	}
}

func TestOldB760MagneticWebMakesMagnetCreaturesBlockTheAttacker(t *testing.T) {
	g, _, opp, a, b, _, c := oldB760WebTable(t)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(b, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(B): %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.OnToBlockers(t, g)
	must := effectstest.MustBlockNow(g)
	if got := must[c]; got != a && got != b {
		t.Errorf("the opposing magnet creature must block %v, want one of the attackers (%v or %v)", got, a, b)
	}
}

func TestOldB760MagneticWebPutsAMagnetCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	web := effectstest.PushCatalogPermanent(g, me.ID, "Magnetic Web", "Artifact", oldB760Web, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FillPool(me, 1)
	if err := effectstest.InvB1Activate(g, me.ID, web, 0, effectstest.CardTarget(bear)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, "magnet"); got != 1 {
		t.Errorf("magnet counters %d, want 1", got)
	}
}
