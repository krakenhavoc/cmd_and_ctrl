package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const scavengedBrawlerOracle = "d895863e-a3c5-4e1f-bd28-e47a8717bbd0"

func TestScavengedBrawlerIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(scavengedBrawlerOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registered %v, spec %+v", ok, spec)
	}
	for _, k := range []string{"flying", "vigilance", "trample", "lifelink"} {
		if !hasKeyword(spec.PrintedKeywords, k) {
			t.Errorf("missing printed keyword %s", k)
		}
	}
}

// From the graveyard: exile it, put four +1/+1 counters and the four
// keyword counters on the target creature.
func TestScavengedBrawlerGivesItsCountersFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := pushBattlefieldCreature(g, me.ID)
	id, _ := activateFromGraveyard(t, g, "Scavenged Brawler", "Artifact Creature — Construct",
		scavengedBrawlerOracle, 4, 4, "{C}{C}{C}{C}{C}", game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
	if !g.Exile.Contains(id) {
		t.Fatal("the cost did not exile the Brawler")
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, target, game.CounterPlusOne); got != 4 {
		t.Errorf("%d +1/+1 counters, want 4", got)
	}
	for _, k := range []string{game.CounterFlying, game.CounterVigilance, game.CounterTrample, game.CounterLifelink} {
		if got := countersOn(g, target, k); got != 1 {
			t.Errorf("%d %s counters, want 1", got, k)
		}
		if !hasEffectiveKeyword(t, g, target, k) {
			t.Errorf("the target lacks %s", k)
		}
	}
	if p, tt := ptOf(t, g, target); p != 6 || tt != 6 {
		t.Errorf("the 2/2 target is %d/%d, want 6/6", p, tt)
	}
}

// Sorcery speed only.
func TestScavengedBrawlerIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := pushBattlefieldCreature(g, me.ID)
	id := pushCatalogGraveyardCard(me, "Scavenged Brawler", "Artifact Creature — Construct", scavengedBrawlerOracle, 4, 4)
	advanceTo(t, g, game.StepBeginCombat)
	fillPoolColored(me, "C", 5)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err == nil {
		t.Error("activated in combat")
	}
}
