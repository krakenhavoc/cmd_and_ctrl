package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const priestOfFellRitesOracle = "ec6ccd8f-cc75-4a50-8639-f4ec31a280aa"

func TestPriestOfFellRitesTapPayLifeSacrificeReanimates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	priest := pushCatalogPermanent(g, me.ID, "Priest of Fell Rites", "Creature — Human Warlock", priestOfFellRitesOracle, false)
	victim := seedGraveyardCard(t, g, "Fodder Bear", "Creature — Bear", "")
	lifeBefore := me.Life

	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(priest) {
		t.Errorf("the Priest should be sacrificed as part of the cost")
	}
	if me.Life != lifeBefore-3 {
		t.Errorf("life %d -> %d, want -3", lifeBefore, me.Life)
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(victim) {
		t.Errorf("the targeted creature should be reanimated")
	}
}

func TestPriestOfFellRitesUnearth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := seedGraveyardCard(t, g, "Priest of Fell Rites", "Creature — Human Warlock", priestOfFellRitesOracle)

	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Unearth: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(id) {
		t.Fatalf("the unearthed Priest should be on the battlefield")
	}
}
