package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b48_cards_test.go — GAP batch 48 (#5948, S82): Modern Horizons 3
// artifacts and creatures. Each test plays the real catalog entry.

const (
	gapB48DisruptorFlute = "d6c46906-1339-4440-8cd0-2bb9890d4975"
	gapB48Dreadmobile    = "51cc270d-55c8-4b6b-a6d7-f98d7b8e5185"
	gapB48Drossclaw      = "cda1ac77-3986-4600-9835-b0c0063ecc2f"
	gapB48CranialRam     = "e958b3c5-0666-4a99-a933-e5cb8b0da9da"
	gapB48MeteoricMace   = "c6a607f2-4228-4bcc-84ed-37bd362d7542"
	gapB48ArgentDais     = "3269f589-00fc-4f0b-8693-656f2ab0670d"
	gapB48Reliquary      = "528039f3-e8e6-499d-add4-43cf2d731b65"
	gapB48ItThatHeralds  = "c14ac951-329d-452b-be75-5c72166cd9c7"
	gapB48ArcboundCondor = "d484be42-5dd4-49e0-afcd-2d3c1a2c3037"
)

func TestGapB48CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB48DisruptorFlute: "Disruptor Flute",
		gapB48Dreadmobile:    "Dreadmobile",
		gapB48Drossclaw:      "Drossclaw",
		gapB48CranialRam:     "Cranial Ram",
		gapB48MeteoricMace:   "Meteoric Mace",
		gapB48ArgentDais:     "Argent Dais",
		gapB48Reliquary:      "Essence Reliquary",
		gapB48ItThatHeralds:  "It That Heralds the End",
		gapB48ArcboundCondor: "Arcbound Condor",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestGapB48CranialRamMakesAGermThatGrowsWithArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	_, germ := effectstest.CastLivingWeapon(t, g, "Cranial Ram", gapB48CranialRam)
	// Two artifacts: the Rock and the Ram; +2/+1 on a 0/0 Germ.
	if p, tg := effectstest.PtOf(t, g, germ); p != 2 || tg != 1 {
		t.Errorf("the Germ is %d/%d, want 2/1", p, tg)
	}
}

func TestGapB48DrossclawDrainsEachOpponentWhenTheEquippedCreatureAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.M3cb1Opponent(g)
	_, germ := effectstest.CastLivingWeapon(t, g, "Drossclaw", gapB48Drossclaw)
	if p, tg := effectstest.PtOf(t, g, germ); p != 1 || tg != 1 {
		t.Fatalf("the Germ is %d/%d, want 1/1", p, tg)
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, germ).SummonedThisTurn = false })
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, germ)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d -> %d, want a point lost", life, opp.Life)
	}
	_ = me
}

func TestGapB48MeteoricMaceEquipsForPlusFourAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	mace := effectstest.PushCatalogPermanent(g, me.ID, "Meteoric Mace", "Artifact — Equipment", gapB48MeteoricMace, false)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, mace, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.PtOf(t, g, bear); p != 6 || tg != 2 {
		t.Errorf("the Bear is %d/%d, want 6/2", p, tg)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "trample") {
		t.Error("the Bear lacks trample")
	}
}

func TestGapB48DreadmobileGrowsBySacrificingAnotherArtifactOrCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	car := effectstest.PushCatalogPermanent(g, me.ID, "Dreadmobile", "Artifact — Vehicle", gapB48Dreadmobile, false)
	fodder := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Fodder", TypeLine: "Creature — Goblin", Power: 1, Toughness: 1})
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, car, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{car}, Strict: true}); err == nil {
		t.Fatal("sacrificed the Dreadmobile to itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, car, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, car, game.CounterPlusOne) != 1 {
		t.Errorf("%d +1/+1 counters, want 1", effectstest.CounterCount(g, car, game.CounterPlusOne))
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still there")
	}
}

func TestGapB48ArgentDaisGrowsAndExilesAPermanentGivingItsControllerTwoCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.M3cb1Opponent(g)
	dais := effectstest.TrekB31Enter(t, g, me, game.Card{Name: "Argent Dais", TypeLine: "Artifact", OracleID: gapB48ArgentDais, ManaCost: "{1}{W}"})
	if got := effectstest.CounterCount(g, dais, "oil"); got != 2 {
		t.Fatalf("%d oil counters on entry, want 2", got)
	}
	victim := effectstest.M3cb1Push(g, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.FloatMana(t, g, me, "{C}{C}")
	hand := opp.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, dais, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the Bear was not exiled")
	}
	if got := opp.Hand.Size(); got != hand+2 {
		t.Errorf("their hand %d -> %d, want +2", hand, got)
	}
	if got := effectstest.CounterCount(g, dais, "oil"); got != 0 {
		t.Errorf("%d oil counters left, want 0", got)
	}
}

func TestGapB48ArgentDaisGetsAnOilCounterWhenTwoCreaturesAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.M3cb1Opponent(g)
	dais := effectstest.TrekB31Enter(t, g, me, game.Card{Name: "Argent Dais", TypeLine: "Artifact", OracleID: gapB48ArgentDais, ManaCost: "{1}{W}"})
	a := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "A", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	b := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "B", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, dais, "oil"); got != 3 {
		t.Errorf("%d oil counters, want 3 (one trigger for the declaration)", got)
	}
}

func TestGapB48EssenceReliquaryReturnsAPermanentAndItsAuras(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rel := effectstest.PushCatalogPermanent(g, me.ID, "Essence Reliquary", "Artifact", gapB48Reliquary, false)
	bear := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	aura := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Pacifism", TypeLine: "Enchantment — Aura"})
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, aura).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: bear}
	})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, rel, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(rel), Strict: true}); err == nil {
		t.Fatal("targeted itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, rel, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(aura) {
		t.Error("the Bear or its Aura is still on the battlefield")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want +2", hand, got)
	}
}

func TestGapB48ItThatHeraldsTheEndPumpsOtherColorlessCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	it := effectstest.PushCatalogPermanent(g, me.ID, "It That Heralds the End", "Creature — Eldrazi Drone", gapB48ItThatHeralds, false)
	golem := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 3, Toughness: 3})
	bear := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2})
	if p, _ := effectstest.PtOf(t, g, golem); p != 4 {
		t.Errorf("the colorless Golem has power %d, want 4", p)
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("the green Bear has power %d, want 2", p)
	}
	if p, _ := effectstest.PtOf(t, g, it); p != 1 {
		t.Errorf("It That Heralds the End has power %d, want its own 1 (no bonus for itself)", p)
	}
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Seven Rock", TypeLine: "Artifact", ManaCost: "{7}", Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the {7} colorless spell for six mana: %v", err)
	}
}

func TestGapB48ArcboundCondorShrinksAnOpponentsCreatureWhenAnotherArtifactEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.M3cb1Opponent(g)
	effectstest.TrekB31Enter(t, g, me, game.Card{Name: "Arcbound Condor", TypeLine: "Artifact Creature — Bird", OracleID: gapB48ArcboundCondor, ManaCost: "{2}{B}{B}",
		Keywords: []string{"flying", "modular 3"}})
	effectstest.PassPriorityAroundTable(t, g)
	victim := effectstest.M3cb1Push(g, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.TrekB31Enter(t, g, me, game.Card{Name: "Rock", TypeLine: "Artifact"})
	effectstest.AnswerPickTargetIfOpen(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.PtOf(t, g, victim); p != 1 || tg != 1 {
		t.Errorf("the Bear is %d/%d, want 1/1", p, tg)
	}
}

func TestGapB48DisruptorFluteTaxesAndBansTheChosenName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.TrekB31Enter(t, g, me, game.Card{Name: "Disruptor Flute", TypeLine: "Artifact", OracleID: gapB48DisruptorFlute, ManaCost: "{2}"})
	effectstest.ActNameFor(t, g, me.ID, "Test Bauble")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Bauble", TypeLine: "Artifact", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast the named spell for three mana; it costs {3} more")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the named spell for four mana: %v", err)
	}
}

func TestGapB48MeteoricMaceHasCascade(t *testing.T) {
	spec, ok := effects.Lookup(gapB48MeteoricMace)
	if !ok {
		t.Fatal("Meteoric Mace is not registered")
	}
	found := false
	for _, tr := range spec.Triggered {
		if tr.Keyword == "cascade" {
			found = true
		}
	}
	if !found {
		t.Error("Meteoric Mace has no cascade trigger")
	}
}
