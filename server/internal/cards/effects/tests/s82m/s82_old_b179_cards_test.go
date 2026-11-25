package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 179 (Ravnica Allegiance: addendum and spectacle, #5521). Every
// test plays the real catalog entry through the engine.

const (
	oldB179Insight   = "01f675c3-d1c4-45c8-8a2e-75f535db3025"
	oldB179Perceive  = "275d0b90-a8d6-4388-b7db-34615287f532"
	oldB179Powers    = "9127ac59-4c7b-4652-b1e5-d82af47f2717"
	oldB179Mark      = "ac34bc09-ba8a-4bfb-bc45-e9c3e4d7bdc0"
	oldB179Admonish  = "ccfb612a-762d-41a7-bc3d-268c6abf8e85"
	oldB179Demon     = "2e7a03ac-44bb-45f1-b209-85fc3b6e90d3"
	oldB179Revels    = "4354eb0a-28c5-4f3f-8e78-fd7eec230c7b"
	oldB179Hackrobat = "65ce903f-8c3e-4283-9b6e-ee9b3176f059"
	oldB179Reveler   = "bad5b8d4-9089-4ebd-8eaa-07552737c527"
	oldB179Acrobat   = "e9b403db-8981-440f-a451-aa77ae3cf006"
)

func TestOldB179EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB179Insight: "Sphinx's Insight", oldB179Perceive: "Precognitive Perception",
		oldB179Powers: "Emergency Powers", oldB179Mark: "Sentinel's Mark",
		oldB179Admonish: "Arrester's Admonition", oldB179Demon: "Rafter Demon",
		oldB179Revels: "Dead Revels", oldB179Hackrobat: "Hackrobat",
		oldB179Reveler: "Rix Maadi Reveler", oldB179Acrobat: "Spikewheel Acrobat",
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

// oldB179CastIn casts a card for the active seat at the given step.
func oldB179CastIn(t *testing.T, g *game.Game, step game.Step, name, typeLine, oracle string, targets []game.TargetRef) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, step)
	id := effectstest.HandCardFull(me, name, typeLine, "", oracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB179SphinxsInsightGainsLifeOnlyInYourMainPhase(t *testing.T) {
	for _, tc := range []struct {
		step game.Step
		life int
	}{{game.StepPrecombatMain, 2}, {game.StepBeginCombat, 0}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		life, hand := me.Life, me.Hand.Size()
		oldB179CastIn(t, g, tc.step, "Sphinx's Insight", "Instant", oldB179Insight, nil)
		if me.Life != life+tc.life {
			t.Errorf("%v: life %d, want %d", tc.step, me.Life, life+tc.life)
		}
		if got := me.Hand.Size(); got != hand+2 {
			t.Errorf("%v: hand %d, want %d", tc.step, got, hand+2)
		}
	}
}

func TestOldB179PrecognitiveBeholdsScryThenDrawsInMainOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	oldB179CastIn(t, g, game.StepPrecombatMain, "Precognitive Perception", "Instant", oldB179Perceive, nil)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("hand %d before the scry is answered, want %d", got, hand)
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d after the scry, want %d", got, hand+3)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	hand = me.Hand.Size()
	oldB179CastIn(t, g, game.StepBeginCombat, "Precognitive Perception", "Instant", oldB179Perceive, nil)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Error("a scry was offered outside the main phase")
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d, want %d", got, hand+3)
	}
}

func TestOldB179EmergencyPowersShufflesDrawsExilesAndPutsInMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dead := effectstest.B17GraveyardCard(opp, "Dead", "Creature — Bear", "{1}{G}")
	effectstest.HandCardFull(me, "Seven Drop", "Creature — Beast", "{7}", "", nil)
	effectstest.HandCardFull(me, "Eight Drop", "Creature — Beast", "{8}", "", nil)
	id := effectstest.HandCardFull(me, "Emergency Powers", "Instant", "", oldB179Powers, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Hand.Size() < 7 {
			t.Errorf("%s holds %d cards, want at least 7", p.Name, p.Hand.Size())
		}
	}
	if opp.Graveyard.Contains(dead) {
		t.Error("the graveyard was not shuffled away")
	}
	if me.Graveyard.Contains(id) {
		t.Error("Emergency Powers went to the graveyard instead of exile")
	}
	if effectstest.FindCardAnywhere(t, g, id).Name == "" {
		t.Error("Emergency Powers is nowhere")
	}
	if effectstest.LatestPutFromHandChoice(g, me.ID) == nil {
		t.Fatal("the addendum offered no put from hand")
	}
}

func TestOldB179EmergencyPowersOutsideMainOffersNoPut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB179CastIn(t, g, game.StepBeginCombat, "Emergency Powers", "Instant", oldB179Powers, nil)
	if me.Hand.Size() < 7 {
		t.Errorf("hand %d, want at least 7", me.Hand.Size())
	}
	if effectstest.LatestPutFromHandChoice(g, me.ID) != nil {
		t.Error("the addendum fired outside the main phase")
	}
}

func TestOldB179SentinelsMarkPumpsAndGrantsLifelinkOnlyFromAMainPhaseCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB179CastIn(t, g, game.StepPrecombatMain, "Sentinel's Mark", "Enchantment — Aura", oldB179Mark, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, to := effectstest.EffectivePTOf(t, g, bear); p != 3 || to != 4 {
		t.Errorf("the creature is %d/%d, want 3/4", p, to)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") || !effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("main-phase cast: want vigilance and lifelink")
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB179CastIn(t, g, game.StepBeginCombat, "Sentinel's Mark", "Enchantment — Aura", oldB179Mark, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("no vigilance")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("lifelink from a cast outside the main phase")
	}
}

func TestOldB179ArrestersAdmonitionBouncesAndDrawsInMainOnly(t *testing.T) {
	for _, tc := range []struct {
		step game.Step
		draw int
	}{{game.StepPrecombatMain, 1}, {game.StepBeginCombat, 0}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		hand := me.Hand.Size()
		oldB179CastIn(t, g, tc.step, "Arrester's Admonition", "Instant", oldB179Admonish, effectstest.CardTarget(bear))
		if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
			t.Errorf("%v: the creature was not returned to its owner's hand", tc.step)
		}
		if got := me.Hand.Size(); got != hand+tc.draw {
			t.Errorf("%v: hand %d, want %d", tc.step, got, hand+tc.draw)
		}
	}
}

// oldB179LoseLife makes an opponent lose life this turn so spectacle is on.
func oldB179LoseLife(t *testing.T, g *game.Game) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, -1); err != nil {
			t.Fatal(err)
		}
	})
}

func TestOldB179SpectacleNeedsAnOpponentToHaveLostLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(me, "Spikewheel Acrobat", "Creature — Human Rogue", "{3}{R}", oldB179Acrobat, []string{"R"})
	params := game.CastSpellParams{Strict: true, AlternativeCost: "spectacle"}
	effectstest.FloatForTest(g, me, "RR")
	if err := g.CastSpell(me.ID, id, params); err == nil {
		t.Fatal("spectacle was accepted with no life lost this turn")
	}
	oldB179LoseLife(t, g)
	effectstest.FloatForTest(g, me, "RR")
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("spectacle after an opponent lost life: %v", err)
	}
}

func TestOldB179RafterDemonMakesOpponentsDiscardOnlyWithSpectacle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	oldB179LoseLife(t, g)
	id := effectstest.HandCardFull(me, "Rafter Demon", "Creature — Demon", "{2}{B}{R}", oldB179Demon, []string{"B", "R"})
	effectstest.FloatForTest(g, me, "CCBRB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "spectacle"}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, opp := range g.Seats[1:] {
		if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
			t.Fatalf("%s was not asked to discard", opp.Name)
		}
		effectstest.DiscardFromHand(t, g, opp.ID)
	}

	g = effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastHoldCreature(t, g, "Rafter Demon", "Creature — Demon", oldB179Demon, 4, 2)
	for _, opp := range g.Seats[1:] {
		if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
			t.Errorf("%s discarded for a hard-cast Demon", opp.Name)
		}
	}
}

func TestOldB179DeadRevelsReturnsTwoCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B17GraveyardCard(me, "Bear", "Creature — Bear", "{1}{G}")
	b := effectstest.B17GraveyardCard(me, "Wolf", "Creature — Wolf", "{1}{G}")
	c := effectstest.B17GraveyardCard(me, "Bolt", "Instant", "{R}")
	effectstest.CastCatalogSpell(t, g, "Dead Revels", "Sorcery", oldB179Revels, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) || me.Hand.Contains(c) {
		t.Error("want both targeted creature cards in hand and the instant left behind")
	}
}

func TestOldB179HackrobatActivations(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Hackrobat", "Creature — Human Rogue", oldB179Hackrobat, 2, 3)
	effectstest.FloatForTest(g, me, "B")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("{B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, id, "deathtouch") {
		t.Error("no deathtouch")
	}
	effectstest.FloatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("{R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, to := effectstest.EffectivePTOf(t, g, id); p != 4 || to != 1 {
		t.Errorf("Hackrobat is %d/%d, want 4/1", p, to)
	}
}

func TestOldB179RixMaadiRevelerLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastHoldCreature(t, g, "Rix Maadi Reveler", "Creature — Human Shaman", oldB179Reveler, 2, 2)
	hand := me.Hand.Size()
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after discard then draw, want %d", got, hand)
	}
}

func TestOldB179RixMaadiRevelerSpectacleDiscardsTheHandAndDrawsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	oldB179LoseLife(t, g)
	id := effectstest.HandCardFull(me, "Rix Maadi Reveler", "Creature — Human Shaman", "{1}{R}", oldB179Reveler, []string{"R"})
	effectstest.FloatForTest(g, me, "CCBR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "spectacle"}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand %d, want exactly the three drawn cards", got)
	}
}
