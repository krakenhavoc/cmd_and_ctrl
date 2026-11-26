package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 274 (Conspiracy: Take the Crown): Pharika's Mender,
// Dread Statuary, Dragonlair Spider.

const (
	oldB274Mender   = "00771a33-7443-43f9-a61e-1737b62b1f81"
	oldB274Statuary = "a9789ce4-69cf-435c-b99a-78a21609830c"
	oldB274Spider   = "f17d3aa4-6e88-4973-b2a4-942098a1c463"
)

func TestOldB274CardsAreFullWithNoCaveats(t *testing.T) {
	for _, oracle := range []string{oldB274Mender, oldB274Statuary, oldB274Spider} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v, want full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB274MenderReturnsACreatureOrEnchantmentCardNotALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.GraveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}{G}", "")
	aura := effectstest.GraveyardCardOf(g, me, "Pacifism", "Enchantment — Aura", "{1}{W}", "")
	land := effectstest.GraveyardCardOf(g, me, "Forest", "Basic Land — Forest", "", "")
	effectstest.CastHoldCreature(t, g, "Pharika's Mender", "Creature — Gorgon", oldB274Mender, 4, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || !effectstest.HasID(p.PickTargetCards, bear) || !effectstest.HasID(p.PickTargetCards, aura) {
		t.Fatal("the creature and enchantment cards must both be offered")
	}
	if effectstest.HasID(p.PickTargetCards, land) {
		t.Error("a land card is offered")
	}
	effectstest.PickCard(t, g, me.ID, aura)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) || me.Hand.Contains(bear) {
		t.Error("only the chosen enchantment card should be in hand")
	}
}

func TestOldB274MenderMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.GraveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.CastHoldCreature(t, g, "Pharika's Mender", "Creature — Gorgon", oldB274Mender, 4, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(bear) {
		t.Error("declined, but the card returned")
	}
}

func TestOldB274StatuaryAnimatesAsA4by2GolemForFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Dread Statuary", "Land", oldB274Statuary, false)
	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("animated for {3}")
	}
	effectstest.FillPool(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.EffectiveCardForTest(t, g, land)
	if !c.IsCreature() || !c.IsLand() || !c.IsArtifact() || !c.HasSubtype("Golem") {
		t.Errorf("creature %v land %v artifact %v Golem %v", c.IsCreature(), c.IsLand(), c.IsArtifact(), c.HasSubtype("Golem"))
	}
	if p, tough := effectstest.EffectivePower(t, g, land), effectstest.EffectiveToughness(t, g, land); p != 4 || tough != 2 {
		t.Errorf("Statuary is %d/%d, want 4/2", p, tough)
	}
}

func TestOldB274SpiderMakesAnInsectOnlyForAnOpponentsSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Dragonlair Spider", "Creature — Spider", oldB274Spider, false)
	count := func() int { n, _ := effectstest.CountTokensNamed(g, me.ID, "Insect"); return n }
	n0 := count()
	g.EmitEvent(game.Event{Kind: game.EventCast, Actor: me.ID, CardID: uuid.New()})
	effectstest.PassPriorityAroundTable(t, g)
	if n := count(); n != n0 {
		t.Errorf("your own spell made %d Insects", n-n0)
	}
	g.EmitEvent(game.Event{Kind: game.EventCast, Actor: opp.ID, CardID: uuid.New()})
	effectstest.PassPriorityAroundTable(t, g)
	if n := count(); n != n0+1 {
		t.Errorf("an opponent's spell made %d Insects, want 1", n-n0)
	}
}
