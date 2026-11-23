package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB83LaserScrewdriver = "9602b5f7-65f9-4447-815e-f60d3e4f7290"
	gapB83Diary            = "9bdd6fea-d436-412b-a7b7-bbdba73dc986"
	gapB83Fireplace        = "f07a8c8d-73a1-44cc-9313-05f3bcf1a67c"
	gapB83Bessie           = "776aa036-08aa-4fa0-82bd-c09e6e3e2c5e"
	gapB83Moment           = "8a5259cc-224d-4748-b764-63ad0e4e02f1"
)

func TestGapB83EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB83LaserScrewdriver: "Laser Screwdriver",
		gapB83Diary:            "Five Hundred Year Diary",
		gapB83Fireplace:        "Rotating Fireplace",
		gapB83Bessie:           "Bessie, the Doctor's Roadster",
		gapB83Moment:           "The Moment",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, %+v", name, ok, spec)
		}
	}
}

func TestGapB83LaserScrewdriverManaTapSurveilAndGoad(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ls := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Laser Screwdriver", TypeLine: "Artifact", OracleID: gapB83LaserScrewdriver})
	rock := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	bear := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, ls, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool %d, want 1 mana", me.ManaPool.Total())
	}
	effectstest.UntapForTest(g, ls)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, ls, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Fatal("a creature was a legal target for tap target artifact")
	}
	if err := g.ActivateCatalogAbility(me.ID, ls, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(rock)}); err != nil {
		t.Fatalf("tap target artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, rock) {
		t.Error("the target artifact is untapped")
	}
	effectstest.UntapForTest(g, ls)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, ls, 2, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("goad: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got, ok := g.LookupCardForEffect(bear); !ok || got.LatestGoader() != me.ID {
		t.Error("the creature was not goaded")
	}
	effectstest.UntapForTest(g, ls)
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, ls, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("surveil: %v", err)
	}
}

func TestGapB83FiveHundredYearDiaryEntersTappedAndCountsClues(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := gapB83Cast(t, g, me, game.Card{Name: "Five Hundred Year Diary", TypeLine: "Legendary Artifact — Book Clue", ManaCost: "{3}{U}", OracleID: gapB83Diary}, "UCCC", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, id) {
		t.Fatal("the Diary did not enter tapped")
	}
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Clue", TypeLine: "Token Artifact — Clue"})
	effectstest.UntapForTest(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if got := me.ManaPool.CountColor("U"); got != 2 {
		t.Errorf("{U} added: %d, want 2 (the Diary is a Clue and so is the token)", got)
	}
	// Sacrifice: draw a card.
	hand := me.Hand.Size()
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || g.Battlefield.Contains(id) {
		t.Errorf("hand %d (want %d), still on battlefield %v", me.Hand.Size(), hand+1, g.Battlefield.Contains(id))
	}
}

func TestGapB83RotatingFireplaceEntersTappedWithATimeCounterAndMakesThatMuchMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := gapB83Cast(t, g, me, game.Card{Name: "Rotating Fireplace", TypeLine: "Artifact", ManaCost: "{3}", OracleID: gapB83Fireplace}, "CCC", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, id) || effectstest.CounterOnBattlefieldCard(g, id, game.CounterTime) != 1 {
		t.Fatalf("tapped %v, time counters %d, want tapped with 1", effectstest.B16Tapped(t, g, id), effectstest.CounterOnBattlefieldCard(g, id, game.CounterTime))
	}
	effectstest.UntapForTest(g, id)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Counters[game.CounterTime] = 3 })
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if got := me.ManaPool.CountColor("C"); got != 3 {
		t.Errorf("{C} added: %d, want 3", got)
	}
	// Time travel is a sorcery-speed ability.
	effectstest.UntapForTest(g, id)
	effectstest.FloatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("time travel: %v", err)
	}
}

func TestGapB83BessieAttackMakesAnotherLegendaryCreatureUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bessie := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bessie, the Doctor's Roadster", TypeLine: "Legendary Artifact Creature — Vehicle", OracleID: gapB83Bessie,
		Power: 3, Toughness: 4, Keywords: []string{"haste"}})
	hero := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Hero", TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2})
	plain := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Plain", TypeLine: "Creature — Human", Power: 2, Toughness: 2})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bessie, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	gapB83Settle(t, g, me.ID, effectstest.B16TargetCard(hero)[0])
	cant := func(id uuid.UUID) bool {
		c, _ := g.LookupCardForEffect(id)
		return c.Effective().Restrictions&game.CantBeBlocked != 0
	}
	if !cant(hero) || cant(plain) || cant(bessie) {
		t.Errorf("can't be blocked: hero %v plain %v bessie %v, want true false false", cant(hero), cant(plain), cant(bessie))
	}
}

func TestGapB83TheMomentCountsTimeAndSweeps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	moment := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "The Moment", TypeLine: "Legendary Artifact", ManaCost: "{2}", OracleID: gapB83Moment,
		Counters: map[string]int{game.CounterTime: 2}})
	two := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Two", TypeLine: "Artifact", ManaCost: "{2}"})
	four := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Four", TypeLine: "Creature — Giant", ManaCost: "{3}{G}", Power: 4, Toughness: 4})
	land := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, moment, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(two) {
		t.Error("the mana value 2 artifact survived two time counters")
	}
	if !g.Battlefield.Contains(four) || !g.Battlefield.Contains(land) {
		t.Error("the mana value 4 creature or the land was destroyed")
	}
	if g.Battlefield.Contains(moment) {
		t.Error("The Moment was not sacrificed")
	}
}

func TestGapB83TheMomentPhasesOutACreatureUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	moment := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "The Moment", TypeLine: "Legendary Artifact", ManaCost: "{2}", OracleID: gapB83Moment})
	bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, moment, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("phase out: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.PhasedOut.Contains(bear) {
		t.Fatal("the creature did not phase out")
	}
	effectstest.EndTurn(t, g)
	effectstest.EndTurn(t, g)
	if !g.PhasedOut.Contains(bear) {
		t.Error("the creature phased back in while The Moment was still on the battlefield")
	}
}

// gapB83Settle passes priority until the stack is empty, answering each
// target prompt `me` owes with the next of `picks` (none left: no target).
func gapB83Settle(t *testing.T, g *game.Game, me uuid.UUID, picks ...game.TargetRef) {
	t.Helper()
	next := 0
	for i := 0; i < 64; i++ {
		if p := effectstest.LatestPickTarget(g, me); p != nil {
			var err error
			if next < len(picks) {
				err = g.ResolvePickTarget(p.ID, me, picks[next])
				next++
			} else {
				err = g.ResolvePickTargets(p.ID, me, nil)
			}
			if err != nil {
				t.Fatalf("answering the target prompt: %v", err)
			}
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
}

// gapB83Cast puts c in me's hand, floats mana, and casts it in a main phase.
func gapB83Cast(t *testing.T, g *game.Game, me *game.Player, c game.Card, mana string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := effectstest.AvB24Hand(g, me, c)
	if mana != "" {
		effectstest.FloatForTest(g, me, mana)
	}
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return id
}
