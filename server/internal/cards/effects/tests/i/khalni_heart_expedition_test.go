package testsi

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const khalniOracle = "be6155de-c5b2-415c-ad83-142f9926462a"

func TestKhalniHeartExpeditionIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(khalniOracle)
	if !ok || spec.Name != "Khalni Heart Expedition" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Khalni Heart Expedition spec = %+v, registered %v", spec, ok)
	}
}

func TestKhalniHeartExpeditionLandfallIsOptionalAndCountsQuest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	k := effectstest.B12Push(g, me.ID, "Khalni Heart Expedition", "Enchantment", khalniOracle, 0, 0)

	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.DrcSettle(t, g, me.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.LookupBattlefield(g, k); c.Counters["quest"] != 1 {
		t.Fatalf("quest counters after one land and a yes = %d, want 1", c.Counters["quest"])
	}

	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.DrcSettle(t, g, me.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.LookupBattlefield(g, k); c.Counters["quest"] != 1 {
		t.Fatalf("quest counters after declining = %d, want 1", c.Counters["quest"])
	}
}

func TestKhalniHeartExpeditionPaysThreeCountersAndFetchesTwoTappedBasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	k := effectstest.B12Push(g, me.ID, "Khalni Heart Expedition", "Enchantment", khalniOracle, 0, 0)
	ids := effectstest.SeedBasicsInLibrary(me, 3)
	effectstest.SetCountersOn(g, k, map[string]int{"quest": 2})

	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, k, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with two quest counters")
	}
	if _, ok := effectstest.LookupBattlefield(g, k); !ok {
		t.Fatal("a refused activation sacrificed Khalni")
	}

	effectstest.SetCountersOn(g, k, map[string]int{"quest": 4})
	if err := g.ActivateCatalogAbility(me.ID, k, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with four counters: %v", err)
	}
	if _, ok := effectstest.LookupBattlefield(g, k); ok {
		t.Fatal("Khalni is still on the battlefield after paying its sacrifice")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	n := 0
	for _, id := range ids {
		if c, ok := effectstest.BattlefieldCard(g, id); ok {
			n++
			if !c.Tapped {
				t.Error("a fetched basic entered untapped")
			}
		}
	}
	if n != 2 {
		t.Errorf("%d basics fetched, want exactly two", n)
	}
}
