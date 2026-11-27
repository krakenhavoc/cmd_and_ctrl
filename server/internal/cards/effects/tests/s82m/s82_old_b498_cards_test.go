package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b498_cards_test.go — Zendikar, batch 498 (#5583).

const (
	oldB498NobleVestigeOracle     = "63ec8167-2dc3-49a6-a7ed-6b1474bbda2c"
	oldB498LivingTsunamiOracle    = "7260657a-c1cd-4081-8ab5-46d604dccca8"
	oldB498SummonersBaneOracle    = "b9b0c209-e346-45ee-a272-0b53589c21df"
	oldB498SavageSilhouetteOracle = "bc243497-d95a-44e0-b7e3-c56b4ff7cd69"
	oldB498GrimDiscoveryOracle    = "be78d7ca-b904-452f-b9eb-34792e588986"
)

func TestOldB498EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB498NobleVestigeOracle, oldB498LivingTsunamiOracle, oldB498SummonersBaneOracle,
		oldB498SavageSilhouetteOracle, oldB498GrimDiscoveryOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestOldB498NobleVestigePreventsOneDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vestige := effectstest.B12Push(g, me.ID, "Noble Vestige", "Creature — Spirit", oldB498NobleVestigeOracle, 1, 2)
	if !effectstest.HasEffectiveKeyword(t, g, vestige, "flying") {
		t.Error("Noble Vestige has no flying")
	}
	// An illegal target (a creature) is refused.
	bear := effectstest.SeedBear(g, me.ID)
	if err := effectstest.ConB1Activate(t, g, me, vestige, 0, "", game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("Noble Vestige targeted a creature")
	}
	if err := effectstest.ConB1Activate(t, g, me, vestige, 0, "", game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	before := opp.Life
	effectstest.Batch01OpponentCasts(t, g, me, "Lightning Bolt", effectstest.LightningBoltOracle, "", []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if got := before - opp.Life; got != 2 {
		t.Errorf("the player lost %d life to a 3-damage Bolt, want 2 (one prevented)", got)
	}
	if err := effectstest.ConB1Activate(t, g, me, vestige, 0, "", game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err == nil {
		t.Error("the tapped Noble Vestige activated again")
	}
}

func oldB498Upkeep(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tsunami := effectstest.B12Push(g, me.ID, "Living Tsunami", "Creature — Elemental", oldB498LivingTsunamiOracle, 4, 4)
	return g, me, tsunami
}

func TestOldB498LivingTsunamiIsSacrificedWithNoLandToReturn(t *testing.T) {
	g, me, tsunami := oldB498Upkeep(t)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tsunami) {
		t.Fatal("Living Tsunami stayed with no land to return")
	}
	if !me.Graveyard.Contains(tsunami) {
		t.Error("Living Tsunami is not in the graveyard")
	}
}

func TestOldB498LivingTsunamiReturnsALandToStay(t *testing.T) {
	g, me, tsunami := oldB498Upkeep(t)
	island := effectstest.MgbLand(g, me.ID, "Island", "Basic Land — Island")
	other := effectstest.MgbLand(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to return a land")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{island, other}); err == nil {
		t.Fatal("two lands were accepted as the payment")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{island}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tsunami) {
		t.Error("Living Tsunami was sacrificed after a land was returned")
	}
	if !me.Hand.Contains(island) || !g.Battlefield.Contains(other) {
		t.Error("the Island is not in hand, or the other land moved")
	}
}

func TestOldB498LivingTsunamiDeclineSacrifices(t *testing.T) {
	g, me, tsunami := oldB498Upkeep(t)
	island := effectstest.MgbLand(g, me.ID, "Island", "Basic Land — Island")
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to return a land")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tsunami) || !g.Battlefield.Contains(island) {
		t.Error("declining should sacrifice the Tsunami and keep the land")
	}
}

func TestOldB498SummonersBaneCountersACreatureSpellAndMakesAnIllusion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	id := effectstest.HandCardFull(me, "Summoner's Bane", "Instant", "{2}{U}{U}", oldB498SummonersBaneOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err == nil {
		t.Fatal("a noncreature spell was a legal target")
	}
	effectstest.PassPriorityAroundTable(t, g)
	bear := effectstest.B13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature spell was not countered")
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && effects.IsToken(c) && c.Name == "Illusion" {
			n++
			if p, tough := effectstest.PtOf(t, g, c.InstanceID); p != 2 || tough != 2 {
				t.Errorf("the Illusion is %d/%d, want 2/2", p, tough)
			}
		}
	}
	if n != 1 {
		t.Errorf("%d Illusion tokens, want 1", n)
	}
}

func TestOldB498SavageSilhouettePumpsAndGrantsRegeneration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Savage Silhouette", "Enchantment — Aura", oldB498SavageSilhouetteOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Fatalf("the enchanted creature is %d/%d, want 4/4", p, tough)
	}
	effectstest.ApaMana(me, "G", "G")
	effectstest.ApaActivate(t, g, me, bear, 0, game.ActivateAbilityParams{})
	if n := effectstest.RegenShieldsOn(t, g, bear); n != 1 {
		t.Fatalf("shields = %d, want 1", n)
	}
	effectstest.DestroyIt(t, g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the shield did not replace the destruction")
	}
}

func TestOldB498GrimDiscoveryReturnsOneOrBoth(t *testing.T) {
	grave := func(p *game.Player, name, typeLine string) uuid.UUID {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
		return id
	}
	t.Run("both", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := effectstest.Top100ActiveSeat(g)
		creature := grave(me, "Bear", "Creature — Bear")
		land := grave(me, "Swamp", "Basic Land — Swamp")
		id := effectstest.HandCardFull(me, "Grim Discovery", "Sorcery", "{1}{B}", oldB498GrimDiscoveryOracle, nil)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}, Targets: effectstest.CardTargets(creature, land)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(creature) || !me.Hand.Contains(land) {
			t.Error("both cards should be back in hand")
		}
	})
	t.Run("only the land mode", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := effectstest.Top100ActiveSeat(g)
		creature := grave(me, "Bear", "Creature — Bear")
		land := grave(me, "Swamp", "Basic Land — Swamp")
		id := effectstest.HandCardFull(me, "Grim Discovery", "Sorcery", "{1}{B}", oldB498GrimDiscoveryOracle, nil)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: effectstest.CardTarget(land)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(land) || me.Hand.Contains(creature) {
			t.Error("only the land should return")
		}
	})
	t.Run("a land card can't be the creature target", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := effectstest.Top100ActiveSeat(g)
		land := grave(me, "Swamp", "Basic Land — Swamp")
		id := effectstest.HandCardFull(me, "Grim Discovery", "Sorcery", "{1}{B}", oldB498GrimDiscoveryOracle, nil)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: effectstest.CardTarget(land)}); err == nil {
			t.Error("a land card was a legal creature-card target")
		}
	})
}
