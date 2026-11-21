package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b99_cards_test.go — Commander Masters creatures, GAP batch 99
// (#5948). Each test plays the real catalog entry through the engine.

const (
	gapOmarthisOracle      = "50676eba-a55a-4749-bb07-74a30e015781"
	gapSkitteringOracle    = "ac141679-6ada-49a8-b225-d1a20c3838db"
	gapZhulodokOracle      = "ec726c54-987b-48ed-8ffa-ec73a5e35333"
	gapCabalPatriarchOracl = "70d1caee-1c35-4fa3-bbc3-a7dcc226be90"
	gapCarrierThrallOracle = "b3e0b8a6-5de4-4988-b042-2641b1ed3e32"
	gapChainerOracle       = "cb905eb0-bb20-41c0-9695-0f5c71972561"
	gapClotSliverOracle    = "2a885729-0618-4a25-b18d-b0587f14718d"
)

func TestGapB99EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gapOmarthisOracle:      "Omarthis, Ghostfire Initiate",
		gapSkitteringOracle:    "Skittering Cicada",
		gapZhulodokOracle:      "Zhulodok, Void Gorger",
		gapCabalPatriarchOracl: "Cabal Patriarch",
		gapCarrierThrallOracle: "Carrier Thrall",
		gapChainerOracle:       "Chainer, Dementia Master",
		gapClotSliverOracle:    "Clot Sliver",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Omarthis, Ghostfire Initiate ---

func TestGapB99OmarthisEntersWithXCounters(t *testing.T) {
	g := newCatalogGame(t)
	id := b12PlayFromHand(t, g, "Omarthis, Ghostfire Initiate", "Legendary Creature — Spirit Snake", gapOmarthisOracle, game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)
	if n := counterOn(g, id, game.CounterPlusOne); n != 2 {
		t.Errorf("X = 2 put %d +1/+1 counters on it, want 2", n)
	}
}

func TestGapB99OmarthisGrowsWhenYouCounterAnotherColorlessCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	omarthis := pushCatalogPermanent(g, me.ID, "Omarthis, Ghostfire Initiate", "Legendary Creature — Spirit Snake", gapOmarthisOracle, false)
	colorless := pushCatalogPermanent(g, me.ID, "Colorless Golem", "Artifact Creature — Golem", "", false)
	red := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Goblin", TypeLine: "Creature — Goblin",
		Colors: []string{"R"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	theirs := pushCatalogPermanent(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "", false)
	put := func(target uuid.UUID, placer uuid.UUID) {
		g.WithWriteLock(func() {
			if err := g.AddCounterByForEffect(placer, target, game.CounterPlusOne, 1); err != nil {
				t.Fatalf("put counter: %v", err)
			}
		})
		passPriorityAroundTable(t, g)
	}
	put(red, me.ID)
	put(theirs, opp.ID)
	if latestConfirmFor(g, me.ID) != nil || counterOn(g, omarthis, game.CounterPlusOne) != 0 {
		t.Fatal("Omarthis reacted to a coloured creature, or to counters another player put")
	}
	put(colorless, me.ID)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := counterOn(g, omarthis, game.CounterPlusOne); n != 1 {
		t.Errorf("Omarthis has %d +1/+1 counters, want 1", n)
	}
	put(omarthis, me.ID)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("Omarthis was offered a counter for counters put on itself")
	}
}

func TestGapB99OmarthisManifestsOneCardPerCounterWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedCheapLibrary(g, me, 6)
	omarthis := pushCatalogPermanent(g, me.ID, "Omarthis, Ghostfire Initiate", "Legendary Creature — Spirit Snake", gapOmarthisOracle, false)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(omarthis, game.CounterPlusOne, 3); err != nil {
			t.Fatalf("counters: %v", err)
		}
	})
	_ = g.DestroyPermanentForEffect(omarthis)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.FaceDown {
			n++
		}
	}
	if n != 3 {
		t.Errorf("%d manifested creatures, want 3", n)
	}
}

// --- Skittering Cicada ---

func TestGapB99SkitteringCicadaGrowsByTheManaValueOfEachColorlessSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	cicada := pushCatalogPermanent(g, me.ID, "Skittering Cicada", "Creature — Insect", gapSkitteringOracle, false)
	rock := handSpell(me, "Colorless Rock", "Artifact", "{3}")
	red := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: red, Name: "Red Rock", TypeLine: "Artifact", ManaCost: "{2}{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, red, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the red spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, cicada); p != 1 {
		t.Errorf("a coloured spell grew the Cicada to power %d", p)
	}
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the colorless spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, cicada); p != 4 || tough != 4 {
		t.Errorf("Cicada is %d/%d, want 4/4 (1/1 and +3/+3 for a mana value 3 spell)", p, tough)
	}
	assertKeywords(t, g, cicada, "flash", "trample")
}

func TestGapB99SkitteringCicadaLetsYouCastColorlessSpellsAtInstantSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Skittering Cicada", "Creature — Insect", gapSkitteringOracle, false)
	rock := handSpell(me, "Colorless Rock", "Artifact", "{0}")
	red := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: red, Name: "Red Rock", TypeLine: "Artifact", ManaCost: "{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepBeginCombat)
	b06AddMana(me, "R")
	if err := g.CastSpell(me.ID, red, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a coloured artifact was cast outside a main phase")
	}
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a colorless artifact outside a main phase: %v", err)
	}
}

// --- Zhulodok, Void Gorger ---

func TestGapB99ZhulodokGivesBigColorlessSpellsFromHandCascadeTwice(t *testing.T) {
	cast := func(name, typeLine, cost string, colors []string) int {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		seedCheapLibrary(g, me, 6)
		pushCatalogPermanent(g, me.ID, "Zhulodok, Void Gorger", "Legendary Creature — Eldrazi", gapZhulodokOracle, false)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Colors: colors, Owner: me.ID, Controller: me.ID})
		advanceToMain(t, g)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
		offers := 0
		for i := 0; i < 24; i++ {
			var order *game.PendingChoice
			for _, c := range g.PendingChoices {
				if c != nil && c.Kind == game.PendingChoiceTriggerOrder {
					order = c
				}
			}
			if order != nil {
				if err := g.ResolveTriggerOrder(order.ID, order.Chooser, append([]uuid.UUID(nil), order.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				continue
			}
			if n := answerAllMayCast(t, g, me.ID, false); n > 0 {
				offers += n
				continue
			}
			if stackFullyEmpty(g) {
				break
			}
			passPriorityAroundTable(t, g)
		}
		return offers
	}
	if n := cast("Big Rock", "Artifact", "{7}", nil); n != 2 {
		t.Errorf("a colorless 7-drop got %d cascade offers, want 2", n)
	}
	if n := cast("Small Rock", "Artifact", "{6}", nil); n != 0 {
		t.Errorf("a colorless 6-drop got %d cascade offers, want 0", n)
	}
	if n := cast("Big Red Rock", "Artifact", "{6}{R}", []string{"R"}); n != 0 {
		t.Errorf("a red 7-drop got %d cascade offers, want 0", n)
	}
}

// --- Cabal Patriarch ---

func TestGapB99CabalPatriarchShrinksByTwoForASacrificeOrAGraveyardCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	patriarch := pushCatalogPermanent(g, me.ID, "Cabal Patriarch", "Legendary Creature — Human Wizard", gapCabalPatriarchOracl, false)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	victim := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)
	b06AddMana(me, "B", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, patriarch, 0, game.ActivateAbilityParams{Targets: cardRefs(victim), SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("sacrifice mode: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if p, tough := effectivePTOf(t, g, victim); p != 3 || tough != 3 {
		t.Errorf("the target is %d/%d, want 3/3", p, tough)
	}
	yard := pushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	b06AddMana(me, "B", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, patriarch, 1, game.ActivateAbilityParams{Targets: cardRefs(victim), ExileIDs: []uuid.UUID{yard}, Strict: true}); err != nil {
		t.Fatalf("exile mode: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(yard) {
		t.Error("the graveyard creature card was not exiled")
	}
	if p, tough := effectivePTOf(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
}

// --- Carrier Thrall ---

func TestGapB99CarrierThrallLeavesAnEldraziScionWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	thrall := pushCatalogPermanent(g, me.ID, "Carrier Thrall", "Creature — Vampire", gapCarrierThrallOracle, false)
	_ = g.DestroyPermanentForEffect(thrall)
	passPriorityAroundTable(t, g)
	var scion *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Controller == me.ID && c.Name == "Eldrazi Scion" {
			scion = c
		}
	}
	if scion == nil || !scion.IsToken() || scion.Power != 1 || scion.Toughness != 1 || !scion.IsColorless() {
		t.Fatalf("the token is %+v, want a 1/1 colorless Eldrazi Scion", scion)
	}
	if err := g.ActivateManaAbility(me.ID, scion.InstanceID, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("sacrifice the Scion for mana: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
}

// --- Chainer, Dementia Master ---

func TestGapB99ChainerReanimatesBlackNightmaresAndExilesThemWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	chainer := pushCatalogPermanent(g, me.ID, "Chainer, Dementia Master", "Legendary Creature — Human Minion", gapChainerOracle, false)
	dead := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Their Dead Knight", TypeLine: "Creature — Human Knight",
			Colors: []string{"W"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	})
	b06AddMana(me, "B", "B", "B")
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, chainer, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d, want %d", me.Life, life-3)
	}
	c := findBattlefieldCardForTest(g, dead)
	if c == nil || c.Controller != me.ID {
		t.Fatalf("the creature card did not enter under your control: %+v", c)
	}
	live := layeredCard(t, g, dead)
	if cols := live.EffectiveColors(); len(cols) != 1 || cols[0] != "B" {
		t.Errorf("colours %v, want only black", cols)
	}
	if !containsString(effectiveSubtypes(t, g, dead), "Nightmare") || !containsString(effectiveSubtypes(t, g, dead), "Knight") {
		t.Errorf("subtypes %v, want Knight and Nightmare", effectiveSubtypes(t, g, dead))
	}
	if p, tough := effectivePTOf(t, g, dead); p != 3 || tough != 3 {
		t.Errorf("the Nightmare is %d/%d, want 3/3 with Chainer's anthem", p, tough)
	}
	_ = g.DestroyPermanentForEffect(chainer)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dead) || !g.Exile.Contains(dead) {
		t.Error("the Nightmare was not exiled when Chainer left the battlefield")
	}
}

// --- Clot Sliver ---

func TestGapB99ClotSliverLetsEverySliverRegenerateItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Clot Sliver", "Creature — Sliver", gapClotSliverOracle, false)
	theirs := tspB1Sliver(g, opp.ID, "Their Sliver", 1, 1)
	bear := seedBear(g, opp.ID)
	if idx, _ := grantedActivatedIndex(t, g, bear); idx >= 0 {
		t.Fatal("a non-Sliver has the granted ability")
	}
	idx, ref := grantedActivatedIndex(t, g, theirs)
	if idx < 0 {
		t.Fatal("an opponent's Sliver has no granted ability")
	}
	if err := g.ActivateCatalogAbility(opp.ID, theirs, idx, game.ActivateAbilityParams{Ref: ref, Strict: true}); err == nil {
		t.Fatal("the ability was activated with no mana")
	}
	floatMana(t, g, opp, "{C}{C}")
	if err := g.ActivateCatalogAbility(opp.ID, theirs, idx, game.ActivateAbilityParams{Ref: ref, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	_ = g.DestroyPermanentForEffect(theirs)
	if !g.Battlefield.Contains(theirs) {
		t.Error("the Sliver was destroyed despite its regeneration shield")
	}
}
