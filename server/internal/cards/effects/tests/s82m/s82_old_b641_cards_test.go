package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b641_cards_test.go — Planeshift's non-mana kickers, batch 641 (#5615).

const (
	b641RushingRiver  = "6e2a695d-305a-4bbe-b302-11fdc5c2a451"
	b641BogDown       = "8574f57a-768d-42a0-ac8c-cc0a1de181e0"
	b641ArcticMerfolk = "c56671e1-28a6-40c3-bad5-b5e617ddb7da"
	b641MagmaBurst    = "f5af6d60-decd-4e04-8841-da2d3000a405"
	b641PollenRemedy  = "ff52f955-98de-4f12-9ec7-f1d73e3edae9"
)

func TestOldB641EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b641RushingRiver: "Rushing River", b641BogDown: "Bog Down", b641ArcticMerfolk: "Arctic Merfolk",
		b641MagmaBurst: "Magma Burst", b641PollenRemedy: "Pollen Remedy",
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

func b641Table(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	return g, me, opp
}

func b641Land(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.B12Permanent(g, owner, name, "Basic Land — "+name)
}

func b641Hand(g *game.Game, me *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	return effectstest.HandCardOf(g, me, name, typeLine, cost, oracle)
}

func TestOldB641RushingRiverKickedBouncesTwo(t *testing.T) {
	g, me, opp := b641Table(t)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	land := b641Land(g, me.ID, "Island")
	spell := b641Hand(g, me, "Rushing River", "Instant", "{2}{U}", b641RushingRiver)
	effectstest.B06AddMana(me, "U", "C", "C")
	err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.CardRefs(a, b), OptionalCosts: []int{0},
		SacrificeIDs: []uuid.UUID{land}, Strict: true})
	if err != nil {
		t.Fatalf("kicked Rushing River: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the sacrificed land is still on the battlefield")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(a) || !opp.Hand.Contains(b) {
		t.Error("a kicked Rushing River did not return both permanents")
	}
}

func TestOldB641RushingRiverUnkickedBouncesOneAndRefusals(t *testing.T) {
	g, me, opp := b641Table(t)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	land := b641Land(g, opp.ID, "Forest")
	spell := b641Hand(g, me, "Rushing River", "Instant", "{2}{U}", b641RushingRiver)
	effectstest.B06AddMana(me, "U", "C", "C")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.CardRefs(land), Strict: true}); err == nil {
		t.Fatal("a land was accepted as the target")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.CardRefs(a, a), OptionalCosts: []int{0},
		SacrificeIDs: []uuid.UUID{b641Land(g, me.ID, "Island")}, Strict: true}); err == nil {
		t.Fatal("the same permanent was accepted as both targets")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.CardRefs(a), Strict: true}); err != nil {
		t.Fatalf("unkicked Rushing River: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(a) || opp.Hand.Contains(b) {
		t.Error("an unkicked Rushing River must return only its one target")
	}
}

func TestOldB641BogDownDiscardsTwoOrThreeWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g, me, opp := b641Table(t)
		opp.Hand.Cards = nil
		effectstest.FillHand(opp, "A", "B", "C", "D")
		l1, l2 := b641Land(g, me.ID, "Swamp"), b641Land(g, me.ID, "Swamp")
		spell := b641Hand(g, me, "Bog Down", "Sorcery", "{2}{B}", b641BogDown)
		effectstest.B06AddMana(me, "B", "C", "C")
		p := game.CastSpellParams{Targets: effectstest.PlayerRefs(opp.ID), Strict: true}
		if kicked {
			p.OptionalCosts = []int{0}
			p.SacrificeIDs = []uuid.UUID{l1}
			if err := g.CastSpell(me.ID, spell, p); err == nil {
				t.Fatal("a kicker paid with one land was accepted")
			}
			p.SacrificeIDs = []uuid.UUID{l1, l2}
		}
		if err := g.CastSpell(me.ID, spell, p); err != nil {
			t.Fatalf("kicked=%v: %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		d := effectstest.DiscardChoiceFor(g, opp.ID)
		if d == nil {
			t.Fatalf("kicked=%v: the target was not asked to discard", kicked)
		}
		want := 2
		if kicked {
			want = 3
			if g.Battlefield.Contains(l1) || g.Battlefield.Contains(l2) {
				t.Error("the two sacrificed lands are still on the battlefield")
			}
		}
		if d.ChooseMin != want {
			t.Errorf("kicked=%v: discard %d, want %d", kicked, d.ChooseMin, want)
		}
	}
}

func TestOldB641ArcticMerfolkKickedEntersWithACounter(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g, me, _ := b641Table(t)
		other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		spell := b641Hand(g, me, "Arctic Merfolk", "Creature — Merfolk", "{1}{U}", b641ArcticMerfolk)
		effectstest.B06AddMana(me, "U", "C")
		p := game.CastSpellParams{Strict: true}
		if kicked {
			p.OptionalCosts = []int{0}
			p.ReturnIDs = []uuid.UUID{other}
		}
		if err := g.CastSpell(me.ID, spell, p); err != nil {
			t.Fatalf("kicked=%v: %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.E2Card(t, g, spell)
		want := 0
		if kicked {
			want = 1
			if !me.Hand.Contains(other) {
				t.Error("the returned creature is not in hand")
			}
		}
		if got := c.Counters[game.CounterPlusOne]; got != want {
			t.Errorf("kicked=%v: %d +1/+1 counters, want %d", kicked, got, want)
		}
	}
}

func TestOldB641MagmaBurstKickedHitsTwoDifferentTargets(t *testing.T) {
	g, me, opp := b641Table(t)
	bear := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	l1, l2 := b641Land(g, me.ID, "Mountain"), b641Land(g, me.ID, "Mountain")
	spell := b641Hand(g, me, "Magma Burst", "Instant", "{3}{R}", b641MagmaBurst)
	effectstest.B06AddMana(me, "R", "C", "C", "C")
	refs := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: bear}}
	life := opp.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: refs, OptionalCosts: []int{0},
		SacrificeIDs: []uuid.UUID{l1, l2}, Strict: true}); err != nil {
		t.Fatalf("kicked Magma Burst: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, opp.Life)
	}
	if d := effectstest.DamageOn(t, g, bear); d != 3 {
		t.Errorf("the Ogre has %d damage, want 3", d)
	}
}

func TestOldB641MagmaBurstUnkickedHitsOne(t *testing.T) {
	g, me, opp := b641Table(t)
	spell := b641Hand(g, me, "Magma Burst", "Instant", "{3}{R}", b641MagmaBurst)
	effectstest.B06AddMana(me, "R", "C", "C", "C")
	life := opp.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.PlayerRefs(opp.ID), Strict: true}); err != nil {
		t.Fatalf("Magma Burst: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, opp.Life)
	}
}

func TestOldB641PollenRemedyPreventsThreeOrSixDivided(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g, me, opp := b641Table(t)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		land := b641Land(g, me.ID, "Plains")
		spell := b641Hand(g, me, "Pollen Remedy", "Instant", "{W}", b641PollenRemedy)
		effectstest.B06AddMana(me, "W")
		total := 3
		p := game.CastSpellParams{Targets: append(effectstest.CardRefs(bear), game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}),
			Distribution: map[uuid.UUID]int{bear: 2, me.ID: 1}, Strict: true}
		if kicked {
			total = 6
			p.OptionalCosts = []int{0}
			p.SacrificeIDs = []uuid.UUID{land}
			if err := g.CastSpell(me.ID, spell, p); err == nil {
				t.Fatal("a kicked division of three was accepted")
			}
			p.Distribution = map[uuid.UUID]int{bear: 4, me.ID: 2}
		} else if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: p.Targets, Distribution: map[uuid.UUID]int{bear: 4, me.ID: 2}, Strict: true}); err == nil {
			t.Fatal("an unkicked division of six was accepted")
		}
		if err := g.CastSpell(me.ID, spell, p); err != nil {
			t.Fatalf("kicked=%v: %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		_ = opp
		life := me.Life
		g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 9) })
		g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.New(), me.ID, 9) })
		bearShield := 9 - effectstest.E2Card(t, g, bear).DamageMarked
		meShield := 9 - (life - me.Life)
		if bearShield+meShield != total {
			t.Errorf("kicked=%v: prevented %d, want %d", kicked, bearShield+meShield, total)
		}
	}
}
