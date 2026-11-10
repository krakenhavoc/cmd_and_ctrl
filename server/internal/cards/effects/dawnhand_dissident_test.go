package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// dawnhand_dissident_test.go — #5406: Dawnhand Dissident's two blight
// abilities, and its cast of creature spells you own exiled with it,
// paid by removing three counters from among creatures you control in
// addition to their other costs (the #5390 counter cost on a permission's
// offer).

const dawnhandDissidentOracle = "2cabf7a4-17af-45c7-b314-a84f5dc0d070"

// ddTable is seat 0's main phase with Dawnhand Dissident, a 3/3 Bear (two
// +1/+1 counters and a shield counter on it) and a 10/10 to blight under
// seat 0's control,
// a creature card and a sorcery card in seat 0's graveyard, and a creature
// card in seat 1's graveyard.
type ddTable struct {
	g               *game.Game
	me, opp         *game.Player
	dissident, bear uuid.UUID
	fodder          uuid.UUID // a 10/10 the blights go on
	myBeast, mySorc uuid.UUID
	theirBeast      uuid.UUID
}

func newDDTable(t *testing.T) *ddTable {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMainOf(t, g, 0)
	d := &ddTable{g: g, me: g.Seats[0], opp: g.Seats[1]}
	d.dissident = b12Push(g, d.me.ID, "Dawnhand Dissident", "Creature — Elf Warlock", dawnhandDissidentOracle, 1, 2)
	d.bear = b12Push(g, d.me.ID, "Bear", "Creature — Bear", "", 3, 3)
	findBattlefieldCardForTest(g, d.bear).Counters = map[string]int{game.CounterPlusOne: 2, "shield": 1}
	d.fodder = b12Push(g, d.me.ID, "Fodder", "Creature — Giant", "", 10, 10)
	d.myBeast = graveyardCardOf(g, d.me, "Grave Beast", "Creature — Beast", "{1}{G}", "")
	d.mySorc = graveyardCardOf(g, d.me, "Grave Rite", "Sorcery", "{1}{B}", "")
	d.theirBeast = graveyardCardOf(g, d.opp, "Their Beast", "Creature — Beast", "{1}{G}", "")
	return d
}

// exile activates the second ability on `card`, blighting the Fodder.
func (d *ddTable) exile(t *testing.T, card uuid.UUID) {
	t.Helper()
	findBattlefieldCardForTest(d.g, d.dissident).Tapped = false
	b16Activate(t, d.g, d.me.ID, d.dissident, 1, game.ActivateAbilityParams{
		Strict:    true,
		BlightIDs: []uuid.UUID{d.fodder},
		Targets:   []game.TargetRef{{Kind: game.TargetCard, ID: card}},
	})
	passPriorityAroundTable(t, d.g)
}

// castFromDissident casts `card` from exile through the Dissident's offer,
// removing the Bear's three counters.
func (d *ddTable) castFromDissident(card uuid.UUID) error {
	return d.g.CastSpell(d.me.ID, card, game.CastSpellParams{
		FromZone: "exile", AlternativeCost: dawnhandDissidentAltCostKey, Strict: true,
		CounterSourceIDs: []uuid.UUID{d.bear, d.bear}, CounterCounts: []int{2, 1},
		CounterKinds: []string{game.CounterPlusOne, "shield"},
	})
}

func TestDawnhandDissidentIsFull(t *testing.T) {
	spec, ok := Lookup(dawnhandDissidentOracle)
	if !ok {
		t.Fatal("Dawnhand Dissident is not in the catalog")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Dawnhand Dissident: %s %v, want Full with no caveats", spec.Completeness, spec.Caveats)
	}
	spec, ok = Lookup(celestialReunionOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Celestial Reunion: %+v, want Full with no caveats", spec)
	}
}

// {T}, Blight 1: the counter goes on the creature named, and the surveil
// prompt opens.
func TestDawnhandDissidentBlightOneSurveils(t *testing.T) {
	d := newDDTable(t)
	b16Activate(t, d.g, d.me.ID, d.dissident, 0, game.ActivateAbilityParams{Strict: true, BlightIDs: []uuid.UUID{d.dissident}})
	if got := minusCounters(d.g, d.dissident); got != 1 {
		t.Errorf("-1/-1 counters on the Dissident: %d, want 1", got)
	}
	passPriorityAroundTable(t, d.g)
	if c := surveilChoiceFor(d.g, d.me.ID); c == nil {
		t.Fatal("no surveil prompt")
	}
}

// {T}, Blight 2 exiles the target card, linked to the Dissident; then the
// creature card can be cast from exile on your turn by removing three
// counters from among your creatures, paying its mana cost too.
func TestDawnhandDissidentCastsAnExiledCreatureByRemovingThreeCounters(t *testing.T) {
	d := newDDTable(t)
	d.exile(t, d.myBeast)
	if got := minusCounters(d.g, d.fodder); got != 2 {
		t.Fatalf("-1/-1 counters on the Fodder: %d, want 2", got)
	}
	if !inExile(d.g, d.myBeast) {
		t.Fatal("the target was not exiled")
	}
	payMana(t, d.g, d.me, "{C}{G}")
	if err := d.castFromDissident(d.myBeast); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	if got := qgCounters(d.g, d.bear); got[game.CounterPlusOne] != 0 || got["shield"] != 0 {
		t.Errorf("the Bear's counters = %v, want all three removed", got)
	}
	if len(d.me.ManaPool) != 0 {
		t.Errorf("the mana cost was not paid: pool %v", d.me.ManaPool)
	}
	if item := d.g.StackMeta[d.myBeast]; item == nil || item.Paid.CountersRemoved != 3 {
		t.Fatalf("paid record = %+v, want three counters removed", item)
	}
	passPriorityAroundTable(t, d.g)
	if findBattlefieldCardForTest(d.g, d.myBeast) == nil {
		t.Fatal("the Beast did not resolve onto the battlefield")
	}
}

// Refusals: a noncreature card, a creature card you don't own, two
// counters, no counters named, a plain cast from exile without the offer,
// an opponent's turn, and the Dissident gone.
func TestDawnhandDissidentCastRefusals(t *testing.T) {
	d := newDDTable(t)
	d.exile(t, d.mySorc)
	d.exile(t, d.theirBeast)
	d.exile(t, d.myBeast)
	reset := func() {
		findBattlefieldCardForTest(d.g, d.bear).Counters = map[string]int{game.CounterPlusOne: 2, "shield": 1}
	}
	reset()
	payMana(t, d.g, d.me, "{C}{G}{B}")
	for name, try := range map[string]func() error{
		"a sorcery":       func() error { return d.castFromDissident(d.mySorc) },
		"a card they own": func() error { return d.castFromDissident(d.theirBeast) },
		"no counters": func() error {
			return d.g.CastSpell(d.me.ID, d.myBeast, game.CastSpellParams{FromZone: "exile", AlternativeCost: dawnhandDissidentAltCostKey, Strict: true})
		},
		"no offer claimed": func() error {
			return d.g.CastSpell(d.me.ID, d.myBeast, game.CastSpellParams{FromZone: "exile", Strict: true})
		},
		"two counters removed": func() error {
			return d.g.CastSpell(d.me.ID, d.myBeast, game.CastSpellParams{
				FromZone: "exile", AlternativeCost: dawnhandDissidentAltCostKey, Strict: true,
				CounterSourceIDs: []uuid.UUID{d.bear}, CounterCounts: []int{2}, CounterKinds: []string{game.CounterPlusOne},
			})
		},
	} {
		if err := try(); err == nil {
			t.Errorf("%s: the cast was accepted", name)
		}
		if d.g.Stack.Size() != 0 {
			t.Fatalf("%s: a refused cast left the stack at %d", name, d.g.Stack.Size())
		}
		reset()
	}
	// Two counters in all: the offer cannot be claimed (CR 118.3).
	findBattlefieldCardForTest(d.g, d.bear).Counters = map[string]int{game.CounterPlusOne: 2}
	if err := d.g.CastSpell(d.me.ID, d.myBeast, game.CastSpellParams{
		FromZone: "exile", AlternativeCost: dawnhandDissidentAltCostKey, Strict: true,
		CounterSourceIDs: []uuid.UUID{d.bear}, CounterCounts: []int{2}, CounterKinds: []string{game.CounterPlusOne},
	}); err == nil {
		t.Error("two counters in all paid a three-counter cost")
	}
	reset()
	// The Dissident leaves: the permission goes with it.
	d.g.WithWriteLock(func() { _ = d.g.DestroyPermanentForEffect(d.dissident, game.DestroyOptions{}) })
	if err := d.castFromDissident(d.myBeast); err == nil {
		t.Error("the cast was accepted with the Dissident gone")
	}
}

// Not during an opponent's turn.
func TestDawnhandDissidentCastIsYourTurnOnly(t *testing.T) {
	d := newDDTable(t)
	d.exile(t, d.myBeast)
	advanceToMainOf(t, d.g, 1)
	findBattlefieldCardForTest(d.g, d.bear).Counters = map[string]int{game.CounterPlusOne: 2, "shield": 1}
	payMana(t, d.g, d.me, "{C}{G}")
	if err := d.castFromDissident(d.myBeast); err == nil {
		t.Fatal("the cast was accepted on an opponent's turn")
	}
}

// The bot is offered the cast from exile with one payment of exactly three
// counters, and the payment is one the engine accepts.
func TestDawnhandDissidentBotPath(t *testing.T) {
	d := newDDTable(t)
	d.exile(t, d.myBeast)
	findBattlefieldCardForTest(d.g, d.bear).Counters = map[string]int{game.CounterPlusOne: 2, "shield": 1}
	payMana(t, d.g, d.me, "{C}{G}")
	var p qgMove
	found := false
	for _, m := range legal.EnumerateFor(d.g, d.me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != d.myBeast {
			continue
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		found = true
		break
	}
	if !found {
		t.Fatal("the enumerator offers no cast of the exiled Beast")
	}
	total := 0
	for _, n := range p.CounterCounts {
		total += n
	}
	if p.FromZone != "exile" || p.AlternativeCost != dawnhandDissidentAltCostKey || total != 3 {
		t.Fatalf("move params = %+v, want the Dissident's offer removing three counters", p)
	}
	ids := make([]uuid.UUID, len(p.CounterSourceIDs))
	for i, s := range p.CounterSourceIDs {
		ids[i] = uuid.MustParse(s)
	}
	if err := d.g.CastSpell(d.me.ID, d.myBeast, game.CastSpellParams{
		FromZone: p.FromZone, AlternativeCost: p.AlternativeCost, Strict: true,
		CounterSourceIDs: ids, CounterCounts: p.CounterCounts, CounterKind: p.CounterKind, CounterKinds: p.CounterKinds,
	}); err != nil {
		t.Fatalf("the enumerated payment was refused: %v", err)
	}
}
