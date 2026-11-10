package protocol

import (
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Thornvault Forager
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// forage_mana_view_test.go — #5397: a mana ability whose cost forages
// (Thornvault Forager's "{T}, Forage: Add two mana in any combination of
// colors") ships `cost_branches` on its mana_abilities row, the shape an
// activated row's either/or cost has: the graveyard branch with its
// cards and zone, the Food branch with its sacrifice options, each
// greyed when it can't be paid. A graveyard and a Food are public, so
// every viewer reads the same branches.

func TestForageManaAbilityRowCarriesCostBranches(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forager := game.NewCard("Thornvault Forager", me.ID)
	forager.TypeLine = "Creature — Squirrel Ranger"
	forager.OracleID = "7b18cfad-a1cf-46c2-bcdd-7762c13d50e5"
	forager.Power, forager.Toughness = 2, 2
	var grave []uuid.UUID
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(forager)
		me.Graveyard.Cards = nil
		for range 2 {
			c := game.NewCard("Old Card", me.ID)
			c.TypeLine = "Instant"
			me.Graveyard.PushTop(c)
			grave = append(grave, c.InstanceID)
		}
	})

	rows := controllerFrameCard(t, g, forager.InstanceID).ManaAbilities
	if len(rows) != 2 {
		t.Fatalf("mana_abilities = %+v, want the plain {G} and the forage row", rows)
	}
	if len(rows[0].CostBranches) != 0 {
		t.Errorf("the plain {T}: Add {G} row carries branches %+v", rows[0].CostBranches)
	}
	br := rows[1].CostBranches
	if len(br) != 2 {
		t.Fatalf("forage row cost_branches = %+v, want two", br)
	}
	if b := br[0]; b.Label != "Exile three cards from your graveyard" || b.Payable || b.ExileCostN != 3 ||
		b.ExileCostZone != string(game.ZoneGraveyard) || len(b.ExileCostOptions) != 2 {
		t.Errorf("graveyard branch %+v: want three-card exile, unpayable over two cards", b)
	}
	if b := br[1]; b.Label != "Sacrifice a Food" || b.Payable || b.SacrificeOptions == nil || len(b.SacrificeOptions.Cards) != 0 {
		t.Errorf("Food branch %+v: want unpayable with no Food", b)
	}
	if rows[1].SacrificeOptions != nil || rows[1].ExileCostN != 0 {
		t.Error("the forage row itself carries a branch's clause; only the branches should")
	}

	food := game.NewCard("Food", me.ID)
	food.TypeLine = "Token Artifact — Food"
	g.WithWriteLock(func() { g.Battlefield.PushTop(food) })
	for _, viewer := range []string{me.ID.String(), opp.ID.String()} {
		rows = frameCard(t, g, viewer, forager.InstanceID).ManaAbilities
		if b := rows[1].CostBranches[1]; !b.Payable || len(b.SacrificeOptions.Cards) != 1 || b.SacrificeOptions.Cards[0] != food.InstanceID.String() {
			t.Errorf("viewer %s: Food branch %+v, want payable with the Food", viewer, b)
		}
		if b := rows[1].CostBranches[0]; len(b.ExileCostOptions) != 2 {
			t.Errorf("viewer %s: graveyard branch %+v, want the public graveyard's two cards", viewer, b)
		}
	}
}
