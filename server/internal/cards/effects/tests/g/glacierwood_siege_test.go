package testsg

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const glacierwoodSiegeOracle = "295a831c-490e-42ba-afc1-dab3524a4f0c"

// TestGlacierwoodSiegeTemurMillsOnInstantOrSorceryCastButOpensNoGraveyardPlay
// — the Temur trigger fires on an instant, targets a player, mills
// four, and the Sultai permission is not granted while Temur is
// chosen.
func TestGlacierwoodSiegeTemurMillsOnInstantOrSorceryCastButOpensNoGraveyardPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.CastSiege(t, g, "Glacierwood Siege", glacierwoodSiegeOracle, "Temur")

	land := effectstest.PushGraveyardCardTyped(me, "Dead Forest", "Basic Land — Forest")
	if perm := effectstest.GrantedPermissionOn(g, me.ID, land, game.ZoneGraveyard); perm.Granted() {
		t.Fatalf("a Temur Glacierwood Siege opened a graveyard land play: %+v", perm)
	}

	before := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if got := before - opp.Library.Size(); got != 4 {
		t.Errorf("mill = %d, want 4", got)
	}
}

// TestGlacierwoodSiegeSultaiPlaysALandFromYourGraveyardButDoesNotMill
// — the graveyard land play works, the land drop is spent, and
// casting an instant mills nothing while Sultai is chosen.
func TestGlacierwoodSiegeSultaiPlaysALandFromYourGraveyardButDoesNotMill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.CastSiege(t, g, "Glacierwood Siege", glacierwoodSiegeOracle, "Sultai")

	land := effectstest.PushGraveyardCardTyped(me, "Dead Forest", "Basic Land — Forest")
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err != nil {
		t.Fatalf("play the land from the graveyard: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land never reached the battlefield")
	}
	if got := g.LandsPlayedThisTurnFor(me.ID); got != 1 {
		t.Errorf("land plays this turn = %d, want 1", got)
	}

	before := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 0 {
		t.Errorf("a Sultai Glacierwood Siege milled %d cards on an instant cast, want 0", got)
	}
}

// TestGlacierwoodSiegeNonCreatureSpellRuleOnlyMatchesInstantsAndSorceries
// pins the "instant OR sorcery" filter: casting a creature spell does
// not mill.
func TestGlacierwoodSiegeNonCreatureSpellRuleOnlyMatchesInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.CastSiege(t, g, "Glacierwood Siege", glacierwoodSiegeOracle, "Temur")

	before := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 0 {
		t.Errorf("a creature spell milled %d cards, want 0", got)
	}
}
