package testss82v

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const scavengedBrawlerOracle = "d895863e-a3c5-4e1f-bd28-e47a8717bbd0"

func TestScavengedBrawlerIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(scavengedBrawlerOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registered %v, spec %+v", ok, spec)
	}
	for _, k := range []string{"flying", "vigilance", "trample", "lifelink"} {
		if !effectstest.HasKeyword(spec.PrintedKeywords, k) {
			t.Errorf("missing printed keyword %s", k)
		}
	}
}

// From the graveyard: exile it, put four +1/+1 counters and the four
// keyword counters on the target creature.
func TestScavengedBrawlerGivesItsCountersFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := effectstest.PushBattlefieldCreature(g, me.ID)
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Scavenged Brawler", "Artifact Creature — Construct",
		scavengedBrawlerOracle, 4, 4, "{C}{C}{C}{C}{C}", game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
	if !g.Exile.Contains(id) {
		t.Fatal("the cost did not exile the Brawler")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, target, game.CounterPlusOne); got != 4 {
		t.Errorf("%d +1/+1 counters, want 4", got)
	}
	for _, k := range []string{game.CounterFlying, game.CounterVigilance, game.CounterTrample, game.CounterLifelink} {
		if got := effectstest.CountersOn(g, target, k); got != 1 {
			t.Errorf("%d %s counters, want 1", got, k)
		}
		if !effectstest.HasEffectiveKeyword(t, g, target, k) {
			t.Errorf("the target lacks %s", k)
		}
	}
	if p, tt := effectstest.PtOf(t, g, target); p != 6 || tt != 6 {
		t.Errorf("the 2/2 target is %d/%d, want 6/6", p, tt)
	}
}

// Sorcery speed only.
func TestScavengedBrawlerIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := effectstest.PushBattlefieldCreature(g, me.ID)
	id := effectstest.PushCatalogGraveyardCard(me, "Scavenged Brawler", "Artifact Creature — Construct", scavengedBrawlerOracle, 4, 4)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPoolColored(me, "C", 5)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err == nil {
		t.Error("activated in combat")
	}
}
