package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kicker_x_test.go — #3232: a kicker whose mana half is {X}, with the
// printed floor "X can't be 0" (Thieving Skydiver) or none (Verdeloth
// the Ancient, Kangee, Aerie Keeper). The enumerator and the view
// halves are in internal/legal and internal/protocol; this file is the
// announce gate, the registration rules and the cards.

const (
	thievingSkydiverOracle = "f509ac02-f22e-4d1a-9fe1-afa96dd7806d"
	verdelothOracle        = "8222321f-b74d-4f74-9ee7-79455433345a"
	kangeeOracle           = "555f9e75-8892-4c8e-9c5e-8ab14a589118"
)

// kxHand puts a creature card with the real cost into me's hand.
func kxHand(g *game.Game, me *game.Player, name, oracle, typeLine, mana string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, OracleID: oracle, TypeLine: typeLine,
			ManaCost: mana, Owner: me.ID, Controller: me.ID, Power: 2, Toughness: 2})
	})
	return id
}

func kxPoolSize(g *game.Game, me *game.Player) int {
	n := 0
	g.WithWriteLock(func() { n = len(me.ManaPool) })
	return n
}

func kxOnBattlefield(g *game.Game, id uuid.UUID) (game.Card, bool) { return battlefieldCard(g, id) }

func kxInHand(me *game.Player, id uuid.UUID) bool {
	for _, c := range me.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// A kicked cast with X = 0 is refused at announce: "X can't be 0" is
// the kicker's printed floor, and nothing is paid or moved.
func TestThievingSkydiverKickedWithXZeroIsRefused(t *testing.T) {
	g, me, _ := p7Table(t)
	sky := kxHand(g, me, "Thieving Skydiver", thievingSkydiverOracle, "Creature — Merfolk Rogue", "{1}{U}")
	floatMana(t, g, me, "{U}{C}{C}{C}")
	before := kxPoolSize(g, me)

	err := g.CastSpell(me.ID, sky, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, XValue: 0})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("kicked X=0 = %v, want ErrInvalidParam", err)
	}
	if !kxInHand(me, sky) || kxPoolSize(g, me) != before {
		t.Fatal("a refused cast moved the card or spent mana")
	}
}

// X = 2 kicked: the kicker costs {X} on top of {1}{U}, the trigger
// reads MV <= 2 (not 3), steals the artifact, and attaches an Equipment
// to the Skydiver.
func TestThievingSkydiverKickedStealsAndAttachesAnEquipment(t *testing.T) {
	g, me, opp := p7Table(t)
	sword := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Short Sword", TypeLine: "Artifact — Equipment", ManaCost: "{1}"})
	apaPush(g, opp.ID, opp.ID, game.Card{Name: "Costly Trinket", TypeLine: "Artifact", ManaCost: "{3}"})
	sky := kxHand(g, me, "Thieving Skydiver", thievingSkydiverOracle, "Creature — Merfolk Rogue", "{1}{U}")
	floatMana(t, g, me, "{U}{C}{C}{C}") // {1}{U} + X=2: four mana, exactly

	if err := g.CastSpell(me.ID, sky, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, XValue: 2}); err != nil {
		t.Fatalf("kicked X=2: %v", err)
	}
	if n := kxPoolSize(g, me); n != 0 {
		t.Fatalf("%d mana left in the pool: X was not charged", n)
	}
	if got := settleAnsweringPicks(t, g, me.ID, opp.ID); got != 1 {
		t.Fatalf("%d target prompts, want 1", got)
	}
	stolen, ok := kxOnBattlefield(g, sword)
	if !ok || stolen.Controller != me.ID {
		t.Fatalf("Short Sword controller = %v (found %v), want the Skydiver's controller", stolen.Controller, ok)
	}
	if stolen.AttachedTo.ID != sky {
		t.Fatalf("Short Sword attached to %v, want the Skydiver %v", stolen.AttachedTo.ID, sky)
	}
}

// An artifact above X is not a legal target; with none that cheap the
// trigger never reaches the stack.
func TestThievingSkydiverTriggerNeedsAnArtifactAtMostX(t *testing.T) {
	g, me, opp := p7Table(t)
	trinket := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Costly Trinket", TypeLine: "Artifact", ManaCost: "{3}"})
	sky := kxHand(g, me, "Thieving Skydiver", thievingSkydiverOracle, "Creature — Merfolk Rogue", "{1}{U}")
	floatMana(t, g, me, "{U}{C}{C}")

	if err := g.CastSpell(me.ID, sky, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, XValue: 1}); err != nil {
		t.Fatalf("kicked X=1: %v", err)
	}
	if got := settleAnsweringPicks(t, g, me.ID, opp.ID); got != 0 {
		t.Fatalf("%d target prompts for an MV 3 artifact at X=1, want none", got)
	}
	if c, _ := kxOnBattlefield(g, trinket); c.Controller != opp.ID {
		t.Fatal("an artifact above X was stolen")
	}
	if _, ok := kxOnBattlefield(g, sky); !ok {
		t.Fatal("the Skydiver did not resolve")
	}
}

// Unkicked: no X is announced or charged, there is no trigger, and an
// X announced without the kicker is refused.
func TestThievingSkydiverUnkickedChargesNoXAndHasNoTrigger(t *testing.T) {
	g, me, opp := p7Table(t)
	sword := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Short Sword", TypeLine: "Artifact — Equipment", ManaCost: "{1}"})
	sky := kxHand(g, me, "Thieving Skydiver", thievingSkydiverOracle, "Creature — Merfolk Rogue", "{1}{U}")
	floatMana(t, g, me, "{U}{C}")

	err := g.CastSpell(me.ID, sky, game.CastSpellParams{Strict: true, XValue: 2})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("unkicked X=2 = %v, want ErrInvalidParam", err)
	}
	if err := g.CastSpell(me.ID, sky, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("unkicked cast: %v", err)
	}
	if n := kxPoolSize(g, me); n != 0 {
		t.Fatalf("%d mana left: the printed {1}{U} alone should empty the pool", n)
	}
	if got := settleAnsweringPicks(t, g, me.ID, opp.ID); got != 0 {
		t.Fatalf("an unkicked Skydiver asked for %d targets", got)
	}
	if c, _ := kxOnBattlefield(g, sword); c.Controller != opp.ID {
		t.Fatal("an unkicked Skydiver stole an artifact")
	}
}

// Verdeloth's kicker has no floor: X = 0 is a legal kicked cast and
// makes no tokens; X = 3 makes three Saprolings, each 2/2 under the lord.
func TestVerdelothKickedMakesXSaprolings(t *testing.T) {
	for _, tc := range []struct {
		name string
		x    int
		pool string
		want int
	}{
		{"X=0 is legal", 0, "{G}{G}{C}{C}{C}{C}", 0},
		{"X=3", 3, "{G}{G}{C}{C}{C}{C}{C}{C}{C}", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := p7Table(t)
			v := kxHand(g, me, "Verdeloth the Ancient", verdelothOracle, "Legendary Creature — Treefolk", "{4}{G}{G}")
			floatMana(t, g, me, tc.pool)
			if err := g.CastSpell(me.ID, v, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, XValue: tc.x}); err != nil {
				t.Fatalf("kicked X=%d: %v", tc.x, err)
			}
			settleAnsweringPicks(t, g, me.ID, opp.ID)
			n := 0
			g.WithWriteLock(func() {
				for _, c := range g.Battlefield.Cards {
					if c.Controller == me.ID && c.HasSubtype("Saproling") {
						n++
						if c.Effective().Power != 2 || c.Effective().Toughness != 2 {
							t.Errorf("a Saproling is %d/%d, want 2/2 under Verdeloth", c.Effective().Power, c.Effective().Toughness)
						}
					}
				}
			})
			if n != tc.want {
				t.Fatalf("%d Saprolings, want %d", n, tc.want)
			}
		})
	}
}

// Unkicked Verdeloth creates nothing and charges only {4}{G}{G}.
func TestVerdelothUnkickedMakesNothing(t *testing.T) {
	g, me, opp := p7Table(t)
	v := kxHand(g, me, "Verdeloth the Ancient", verdelothOracle, "Legendary Creature — Treefolk", "{4}{G}{G}")
	floatMana(t, g, me, "{G}{G}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, v, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("unkicked: %v", err)
	}
	settleAnsweringPicks(t, g, me.ID, opp.ID)
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.HasSubtype("Saproling") {
				t.Fatal("an unkicked Verdeloth made a Saproling")
			}
		}
	})
}

// Kangee: kicker {X}{2}, so X = 2 costs four more than {2}{W}{U}; the
// counters land through the enters trigger and pump OTHER Birds.
func TestKangeeKickedPutsXFeatherCountersAndPumpsOtherBirds(t *testing.T) {
	g, me, opp := p7Table(t)
	bird := apaPush(g, me.ID, me.ID, game.Card{Name: "Test Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1})
	k := kxHand(g, me, "Kangee, Aerie Keeper", kangeeOracle, "Legendary Creature — Bird Wizard", "{2}{W}{U}")
	floatMana(t, g, me, "{W}{U}{C}{C}{C}{C}{C}{C}") // {2}{W}{U} + {X}{2} at X=2

	if err := g.CastSpell(me.ID, k, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, XValue: 2}); err != nil {
		t.Fatalf("kicked X=2: %v", err)
	}
	if n := kxPoolSize(g, me); n != 0 {
		t.Fatalf("%d mana left: {X}{2} was not charged in full", n)
	}
	settleAnsweringPicks(t, g, me.ID, opp.ID)
	kc, ok := kxOnBattlefield(g, k)
	if !ok || kc.Counters["feather"] != 2 {
		t.Fatalf("Kangee has %d feather counters (on battlefield %v), want 2", kc.Counters["feather"], ok)
	}
	b, _ := kxOnBattlefield(g, bird)
	if p, tt := b.Effective().Power, b.Effective().Toughness; p != 3 || tt != 3 {
		t.Fatalf("the other Bird is %d/%d, want 3/3", p, tt)
	}
	if kc.Effective().Power != 2 {
		t.Fatalf("Kangee pumped itself: %d power", kc.Effective().Power)
	}
}

// Register holds an optional {X} cost to the shape the engine reads.
func TestRegisterHoldsAKickerWithXToItsShape(t *testing.T) {
	registerForTest(t, Spec{OracleID: "test-kicker-x-ok", Name: "Kicker X OK",
		OptionalCosts: []game.AdditionalCost{KickerX("{X}", 1)}})
	mustPanic(t, "has no {X}", func() {
		Register(Spec{OracleID: "test-kicker-x-floor-no-x", Name: "Floor No X",
			OptionalCosts: []game.AdditionalCost{KickerX("{2}", 1)}})
	})
	mustPanic(t, "negative MinX", func() {
		Register(Spec{OracleID: "test-kicker-x-negative", Name: "Negative Floor",
			OptionalCosts: []game.AdditionalCost{KickerX("{X}", -1)}})
	})
	mustPanic(t, "repeats", func() {
		c := Multikicker("{X}", 3)
		Register(Spec{OracleID: "test-kicker-x-repeat", Name: "Repeating X",
			OptionalCosts: []game.AdditionalCost{c}})
	})
	for _, o := range []string{thievingSkydiverOracle, verdelothOracle, kangeeOracle} {
		costs := game.OptionalCostsFor(o)
		if len(costs) != 1 || costs[0].Key != game.KickerKey || !game.OptionalCostHasX(costs[0]) {
			t.Errorf("%s: optional costs %+v, want one kicker with an {X}", o, costs)
		}
	}
	if c := game.OptionalCostsFor(thievingSkydiverOracle)[0]; c.MinX != 1 || !strings.Contains(c.Label, "{X}") {
		t.Errorf("Thieving Skydiver's kicker = %+v, want MinX 1 labelled with {X}", c)
	}
}
