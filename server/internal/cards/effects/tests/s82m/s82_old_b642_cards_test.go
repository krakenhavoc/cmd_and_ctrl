package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b642_cards_test.go — Planeshift's domain and life cards, batch 642 (#5615).

const (
	b642Stratadon  = "3b78b672-7773-44c6-9a21-f89c058966ff"
	b642Pilgrim    = "542c5eaf-8b03-4406-a794-30c56b61d637"
	b642Disease    = "9480e142-289a-439d-b6a5-2666fb0e5be3"
	b642Allied     = "a5ce2ccd-8c78-4749-8854-cb5fa5f4738e"
	b642Treefolk   = "b153e4c4-6557-43a6-8404-e4450f879945"
	b642Advice     = "14c1adf8-950b-4f64-9c84-956a623448f4"
	b642Scout      = "16345f75-f33e-4a31-adb1-d3563358fcec"
	b642GoblinGame = "4a10c945-ad8b-47b6-8cbd-7b4e2612a9b0"
	b642Mirth      = "cebc1e6d-ab14-4e54-9906-e8a7cf4c2043"
)

func TestOldB642EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b642Stratadon: "Stratadon", b642Pilgrim: "Samite Pilgrim", b642Disease: "Exotic Disease",
		b642Allied: "Allied Strategies", b642Treefolk: "Magnigoth Treefolk", b642Advice: "Malicious Advice",
		b642Scout: "Honorable Scout", b642GoblinGame: "Goblin Game", b642Mirth: "Planeswalker's Mirth",
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

func b642Table(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	return g, me, opp
}

func b642Land(g *game.Game, owner uuid.UUID, typ string) uuid.UUID {
	return effectstest.B12Permanent(g, owner, typ, "Basic Land — "+typ)
}

func b642Hand(g *game.Game, me *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	return effectstest.HandCardOf(g, me, name, typeLine, cost, oracle)
}

func TestOldB642StratadonCostsOneLessPerBasicLandType(t *testing.T) {
	g, me, _ := b642Table(t)
	b642Land(g, me.ID, "Plains")
	b642Land(g, me.ID, "Plains")
	b642Land(g, me.ID, "Island")
	card := b642Hand(g, me, "Stratadon", "Artifact Creature — Beast", "{10}", b642Stratadon)
	effectstest.FillPool(me, 7)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("two basic land types should leave {8}; {7} was accepted")
	}
	effectstest.FillPool(me, 1)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Stratadon for {8} with two types: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(card) {
		t.Error("Stratadon did not enter")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, card, "trample") {
		t.Error("Stratadon has no trample")
	}
}

func TestOldB642SamitePilgrimPreventsXDamage(t *testing.T) {
	g, me, _ := b642Table(t)
	b642Land(g, me.ID, "Plains")
	b642Land(g, me.ID, "Island")
	b642Land(g, me.ID, "Island")
	pilgrim := effectstest.HmlCreature(g, me.ID, "Samite Pilgrim", b642Pilgrim, 1, 1)
	bear := effectstest.HmlCreature(g, me.ID, "Bear", "", 2, 6)
	effectstest.B16Activate(t, g, me.ID, pilgrim, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if !effectstest.TappedForTest(t, g, pilgrim) {
		t.Error("the Pilgrim did not tap")
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 5) })
	if d := effectstest.E2Card(t, g, bear).DamageMarked; d != 3 {
		t.Errorf("5 damage through a 2-shield left %d marked, want 3", d)
	}
}

func TestOldB642ExoticDiseaseDrainsByDomain(t *testing.T) {
	g, me, opp := b642Table(t)
	b642Land(g, me.ID, "Swamp")
	b642Land(g, me.ID, "Forest")
	b642Land(g, me.ID, "Forest")
	card := b642Hand(g, me, "Exotic Disease", "Sorcery", "{4}{B}", b642Disease)
	effectstest.B06AddMana(me, "B", "C", "C", "C", "C")
	l1, l2 := me.Life, opp.Life
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Targets: effectstest.PlayerRefs(opp.ID), Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != l1+2 || opp.Life != l2-2 {
		t.Errorf("life me %d→%d opp %d→%d, want +2/-2", l1, me.Life, l2, opp.Life)
	}
}

func TestOldB642AlliedStrategiesDrawsByTheTargetsDomain(t *testing.T) {
	g, me, opp := b642Table(t)
	b642Land(g, me.ID, "Island")
	b642Land(g, opp.ID, "Swamp")
	b642Land(g, opp.ID, "Mountain")
	b642Land(g, opp.ID, "Mountain")
	for i := 0; i < 5; i++ {
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	card := b642Hand(g, me, "Allied Strategies", "Sorcery", "{4}{U}", b642Allied)
	effectstest.B06AddMana(me, "U", "C", "C", "C", "C")
	before, mine := opp.Hand.Size(), me.Hand.Size()
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Targets: effectstest.PlayerRefs(opp.ID), Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - before; got != 2 {
		t.Errorf("the target drew %d, want 2 (their Swamp and Mountain, not my Island)", got)
	}
	if me.Hand.Size() != mine-1 {
		t.Error("the caster drew")
	}
}

func TestOldB642MagnigothTreefolkHasLandwalkOfEachType(t *testing.T) {
	g, me, _ := b642Table(t)
	tree := effectstest.PushCatalogPermanent(g, me.ID, "Magnigoth Treefolk", "Creature — Treefolk", b642Treefolk, false)
	for _, kw := range []string{"swampwalk", "forestwalk", "islandwalk"} {
		if effectstest.HasKeywordOnBattlefield(t, g, tree, kw) {
			t.Errorf("no lands, yet %s", kw)
		}
	}
	b642Land(g, me.ID, "Swamp")
	b642Land(g, me.ID, "Swamp")
	b642Land(g, me.ID, "Forest")
	if !effectstest.HasKeywordOnBattlefield(t, g, tree, "swampwalk") || !effectstest.HasKeywordOnBattlefield(t, g, tree, "forestwalk") {
		t.Error("missing swampwalk or forestwalk")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, tree, "islandwalk") || effectstest.HasKeywordOnBattlefield(t, g, tree, "mountainwalk") {
		t.Error("has landwalk of a type it does not control")
	}
}

func TestOldB642MaliciousAdviceTapsXAndLosesX(t *testing.T) {
	g, me, opp := b642Table(t)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	card := b642Hand(g, me, "Malicious Advice", "Instant", "{X}{U}{B}", b642Advice)
	effectstest.B06AddMana(me, "U", "B", "C", "C")
	life := me.Life
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Targets: effectstest.CardRefs(a, b, c), XValue: 2, Strict: true}); err == nil {
		t.Fatal("three targets for X=2 were accepted")
	}
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Targets: effectstest.CardRefs(a, b), XValue: 2, Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, a) || !effectstest.TappedForTest(t, g, b) || effectstest.TappedForTest(t, g, c) {
		t.Error("exactly the two targets should be tapped")
	}
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want -2", life, me.Life)
	}
}

func TestOldB642HonorableScoutGainsTwoPerBlackOrRedCreature(t *testing.T) {
	g, me, opp := b642Table(t)
	effectstest.B16Creature(g, opp.ID, "Black", "Creature — Zombie", 1, 1, "B")
	effectstest.B16Creature(g, opp.ID, "Both", "Creature — Goblin", 1, 1, "B", "R")
	effectstest.B16Creature(g, opp.ID, "White", "Creature — Soldier", 1, 1, "W")
	effectstest.B16Creature(g, me.ID, "MyRed", "Creature — Goblin", 1, 1, "R")
	card := b642Hand(g, me, "Honorable Scout", "Creature — Human Soldier Scout", "{W}", b642Scout)
	effectstest.B06AddMana(me, "W")
	life := me.Life
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if me.Life != life+4 {
		t.Errorf("life %d → %d, want +4 (two creatures, 2 each)", life, me.Life)
	}
}

func TestOldB642PlaneswalkersMirthGainsTheRevealedCardsManaValue(t *testing.T) {
	g, me, opp := b642Table(t)
	mirth := effectstest.PushCatalogPermanent(g, me.ID, "Planeswalker's Mirth", "Enchantment", b642Mirth, false)
	opp.Hand.Cards = nil
	b642Hand(g, opp, "Five", "Sorcery", "{4}{R}", "")
	effectstest.B06AddMana(me, "W", "C", "C", "C")
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, mirth, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want +5", life, me.Life)
	}
	// An empty hand reveals nothing and gains nothing.
	opp.Hand.Cards = nil
	effectstest.B06AddMana(me, "W", "C", "C", "C")
	life = me.Life
	effectstest.B16Activate(t, g, me.ID, mirth, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	if me.Life != life {
		t.Errorf("an empty hand changed my life %d → %d", life, me.Life)
	}
}

func TestOldB642GoblinGameFewestLosesHalfAndTiesAllDo(t *testing.T) {
	for _, tie := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		n := len(g.Seats)
		me := g.Seats[g.Turn.ActiveSeat]
		b, c, d := g.Seats[(g.Turn.ActiveSeat+1)%n], g.Seats[(g.Turn.ActiveSeat+2)%n], g.Seats[(g.Turn.ActiveSeat+3)%n]
		effectstest.AdvanceToMain(t, g)
		effectstest.FillPoolColored(me, "R", 2)
		effectstest.FillPool(me, 5)
		effectstest.CastFromHandForTest(t, g, me, "Goblin Game", "Sorcery", "{5}{R}{R}", b642GoblinGame, game.CastSpellParams{Strict: true})
		effectstest.PassPriorityAroundTable(t, g)
		life := map[uuid.UUID]int{me.ID: me.Life, b.ID: b.Life, c.ID: c.Life, d.ID: d.Life}
		pa := effectstest.AnswerNumber(t, g, me.ID, 1)
		if pa.NoMax != true || pa.Min != 1 {
			t.Errorf("the number is 1 or greater with no ceiling: %+v", pa)
		}
		effectstest.AnswerNumber(t, g, b.ID, 3)
		second := 1
		if !tie {
			second = 2
		}
		effectstest.AnswerNumber(t, g, c.ID, second)
		effectstest.AnswerNumber(t, g, d.ID, 4)
		effectstest.PassPriorityAroundTable(t, g)
		half := func(l int) int { return (l + 1) / 2 }
		want := map[uuid.UUID]int{}
		nums := map[uuid.UUID]int{me.ID: 1, b.ID: 3, c.ID: second, d.ID: 4}
		for id, k := range nums {
			l := life[id] - k
			if k == 1 {
				l -= half(l)
			}
			want[id] = l
		}
		for _, p := range []*game.Player{me, b, c, d} {
			if p.Life != want[p.ID] {
				t.Errorf("tie=%v: %s life %d, want %d", tie, p.Name, p.Life, want[p.ID])
			}
		}
	}
}
