package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b785_cards_test.go — Visions, S82 OLD batch 785 (#5631). Each test
// plays the real catalog entry and asserts what its text says.

const (
	oldB785Matopi   = "0308b9f2-11da-41d0-8d18-18e8c0c600ab"
	oldB785TinWing  = "0d5fe27f-db51-4323-aa74-b6fdb4309425"
	oldB785Aku      = "233c6686-ead8-4569-941d-3bfa768ff270"
	oldB785Brass    = "2e114d15-4356-4cef-8f89-260ab4c8d4b9"
	oldB785Marauder = "405674c5-607a-43d3-8b46-8741d2339f80"
	oldB785Phoenix  = "72586a7a-3c73-4012-8664-1db0f268eb60"
	oldB785Mine     = "8a2726a4-29b7-4836-bd0a-9c0d914ed915"
	oldB785Lead     = "e8fb2323-3eed-45a6-bfc4-44292c8b3697"
	oldB785Lichen   = "ee7f698d-55c5-4a7a-803d-304febe6a758"
	oldB785Iron     = "fe7c70dd-76ff-4958-9973-6452c7f071ed"
)

func TestOldB785CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB785TinWing: "Tin-Wing Chimera", oldB785Aku: "Aku Djinn",
		oldB785Brass: "Brass-Talon Chimera", oldB785Marauder: "Phyrexian Marauder", oldB785Phoenix: "Bogardan Phoenix",
		oldB785Mine: "Magma Mine", oldB785Lead: "Lead-Belly Chimera", oldB785Lichen: "Lichenthrope",
		oldB785Iron: "Iron-Heart Chimera",
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

// Matopi Golem rides develop's regeneration-rider seam, which spends the
// rider shield first, so it declares that one caveat.
func TestOldB785MatopiGolemDeclaresItsCaveat(t *testing.T) {
	spec, ok := effects.Lookup(oldB785Matopi)
	if !ok {
		t.Fatal("Matopi Golem is not registered")
	}
	if spec.Completeness != effects.CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Errorf("Matopi Golem: completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
}

func oldB785Counters(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Counters[kind]
}

func TestOldB785MatopiGolemRegeneratesAndPaysAMinusCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	golem := effectstest.B12Push(g, me.ID, "Matopi Golem", "Artifact Creature — Golem", oldB785Matopi, 3, 3)
	effectstest.B06AddMana(me, "R")
	effectstest.B16Activate(t, g, me.ID, golem, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.RegenShieldsOn(t, g, golem); n != 1 {
		t.Fatalf("%d shields after the ability resolved, want 1", n)
	}
	effectstest.DestroyIt(t, g, golem)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(golem) {
		t.Fatal("the Golem died despite its shield")
	}
	if n := oldB785Counters(t, g, golem, game.CounterMinusOne); n != 1 {
		t.Errorf("%d -1/-1 counters after regenerating, want 1", n)
	}
	if p, tough := effectstest.RfPT(t, g, golem); p != 2 || tough != 2 {
		t.Errorf("the Golem is %d/%d, want 2/2", p, tough)
	}
	// The shield is gone: a second destruction kills it.
	effectstest.DestroyIt(t, g, golem)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) {
		t.Error("the Golem survived a second destruction with no shield")
	}
}

func TestOldB785MatopiGolemAnotherSourcesRegenerationPutsNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	golem := effectstest.B12Push(g, me.ID, "Matopi Golem", "Artifact Creature — Golem", oldB785Matopi, 3, 3)
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(golem); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.DestroyIt(t, g, golem)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(golem) {
		t.Fatal("the Golem died despite its shield")
	}
	if n := oldB785Counters(t, g, golem, game.CounterMinusOne); n != 0 {
		t.Errorf("%d -1/-1 counters after another source's regeneration, want 0", n)
	}
}

func TestOldB785AkuDjinnPutsACounterOnEachOpposingCreatureAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Aku Djinn", "Creature — Djinn", oldB785Aku, 5, 6)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, g.Seats[2].ID, "B", "Creature — Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if n := oldB785Counters(t, g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("an opposing creature has %d +1/+1 counters, want 1", n)
		}
	}
	if n := oldB785Counters(t, g, mine, game.CounterPlusOne); n != 0 {
		t.Errorf("your own creature got %d counters, want 0", n)
	}
	// Not at an opponent's upkeep.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB785Counters(t, g, a, game.CounterPlusOne); n != 1 {
		t.Errorf("an opponent's upkeep gave %d counters in all, want 1", n)
	}
}

func TestOldB785PhyrexianMarauderEntersWithXCantBlockAndTaxesItsAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastXSpell(t, g, "Phyrexian Marauder", "Artifact Creature — Phyrexian Construct", oldB785Marauder, "{X}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB785Counters(t, g, id, game.CounterPlusOne); n != 2 {
		t.Fatalf("%d +1/+1 counters on entering with X=2, want 2", n)
	}
	// Its attack costs {1} per counter.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	m := effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "Phyrexian Marauder", TypeLine: "Artifact Creature — Phyrexian Construct",
		OracleID: oldB785Marauder, Power: 0, Toughness: 0, Owner: me2.ID, Controller: me2.ID, Counters: map[string]int{"+1/+1": 2}})
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(m, opp2.ID); err == nil {
		t.Fatal("the Marauder attacked without paying")
	}
	effectstest.B06AddMana(me2, "R")
	if err := g2.DeclareAttacker(m, opp2.ID); err == nil {
		t.Fatal("the Marauder attacked having paid {1} for two counters")
	}
	effectstest.B06AddMana(me2, "R")
	if err := g2.DeclareAttacker(m, opp2.ID); err != nil {
		t.Fatalf("the Marauder could not attack with {2} available: %v", err)
	}
}

func TestOldB785PhyrexianMarauderCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Raider", "Creature — Bear", 1, 1)
	m := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Phyrexian Marauder", TypeLine: "Artifact Creature — Phyrexian Construct",
		OracleID: oldB785Marauder, Power: 0, Toughness: 0, Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{"+1/+1": 2}})
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(m, attacker); err == nil {
		t.Fatal("the Marauder blocked")
	} else if errors.Is(err, game.ErrBlockTaxUnpaid) {
		t.Fatalf("refused for the wrong reason: %v", err)
	}
}

func TestOldB785LichenthropeTurnsDamageIntoMinusCountersAndShedsOneEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	l := effectstest.B12Push(g, me.ID, "Lichenthrope", "Creature — Plant Fungus", oldB785Lichen, 5, 5)
	src := effectstest.B12Creature(g, g.Seats[1].ID, "Pinger", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, l, 3) })
	effectstest.CheckState(t, g)
	if n := oldB785Counters(t, g, l, game.CounterMinusOne); n != 3 {
		t.Fatalf("%d -1/-1 counters after 3 damage, want 3", n)
	}
	if c, _ := g.LookupCardForEffect(l); c.DamageMarked != 0 {
		t.Errorf("%d damage marked, want none", c.DamageMarked)
	}
	if p, tough := effectstest.RfPT(t, g, l); p != 2 || tough != 2 {
		t.Errorf("Lichenthrope is %d/%d, want 2/2", p, tough)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB785Counters(t, g, l, game.CounterMinusOne); n != 2 {
		t.Errorf("%d -1/-1 counters after the upkeep, want 2", n)
	}
}

func TestOldB785BogardanPhoenixReturnsOnceThenIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	phoenix := effectstest.B12Push(g, me.ID, "Bogardan Phoenix", "Creature — Phoenix", oldB785Phoenix, 3, 3)
	effectstest.DestroyIt(t, g, phoenix)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(phoenix) {
		t.Fatal("the Phoenix did not return after its first death")
	}
	if n := oldB785Counters(t, g, phoenix, "death"); n != 1 {
		t.Fatalf("%d death counters on the returned Phoenix, want 1", n)
	}
	effectstest.DestroyIt(t, g, phoenix)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(phoenix) {
		t.Fatal("the Phoenix returned a second time")
	}
	inExile := false
	for _, c := range g.Exile.Cards {
		if c.InstanceID == phoenix {
			inExile = true
		}
	}
	if !inExile {
		t.Error("the Phoenix with a death counter was not exiled")
	}
}

func TestOldB785MagmaMineBuildsPressureAndDealsItWhenSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Magma Mine", "Artifact", oldB785Mine, false)
	for i := 0; i < 3; i++ {
		effectstest.FloatForTest(g, me, "CCCC")
		effectstest.B16Activate(t, g, me.ID, mine, 0, game.ActivateAbilityParams{Strict: true})
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := oldB785Counters(t, g, mine, "pressure"); n != 3 {
		t.Fatalf("%d pressure counters, want 3", n)
	}
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, mine, 1, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the Mine was not sacrificed")
	}
	if got := life - opp.Life; got != 3 {
		t.Errorf("the Mine dealt %d, want 3", got)
	}
}

func TestOldB785ChimerasGiveACounterAndAKeywordThatLastsIndefinitely(t *testing.T) {
	for _, tc := range []struct{ name, oracle, keyword string }{
		{"Tin-Wing Chimera", oldB785TinWing, "flying"},
		{"Brass-Talon Chimera", oldB785Brass, "first strike"},
		{"Lead-Belly Chimera", oldB785Lead, "trample"},
		{"Iron-Heart Chimera", oldB785Iron, "vigilance"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			src := effectstest.B12Push(g, me.ID, tc.name, "Artifact Creature — Chimera", tc.oracle, 2, 2)
			other := effectstest.B12Push(g, me.ID, "Other Chimera", "Artifact Creature — Chimera", "", 2, 2)
			bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			// A non-Chimera is not a legal target.
			if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Strict: true,
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
				t.Fatal("a Bear was targeted by a Chimera ability")
			}
			effectstest.B16Activate(t, g, me.ID, src, 0, game.ActivateAbilityParams{Strict: true,
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}})
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(src) {
				t.Error("the Chimera was not sacrificed")
			}
			if n := oldB785Counters(t, g, other, "+2/+2"); n != 1 {
				t.Errorf("%d +2/+2 counters on the target, want 1", n)
			}
			if !effectstest.HasKeywordOf(g, other, tc.keyword) {
				t.Errorf("the target did not gain %s", tc.keyword)
			}
			if bearHas := effectstest.HasKeywordOf(g, bear, tc.keyword); bearHas {
				t.Errorf("the Bear gained %s", tc.keyword)
			}
			// Indefinitely: still there after the turn ends.
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			if !effectstest.HasKeywordOf(g, other, tc.keyword) {
				t.Errorf("the keyword %s did not last past the turn", tc.keyword)
			}
		})
	}
}
