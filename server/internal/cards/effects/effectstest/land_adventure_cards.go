package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// land_adventure_cards_test.go — the four Town lands with an Adventure
// half (#2176). The engine lifecycle is pinned in
// game/adventure_land_test.go; these prove each card's two halves.

// landAdventureCard builds a printed land // adventure card the way the
// deck importer does.
func LandAdventureCard(owner uuid.UUID, oracle, landName, adventureName, cost string, colors []string) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: landName, TypeLine: "Land — Town"},
			{Name: adventureName, TypeLine: "Sorcery — Adventure", ManaCost: cost, Colors: colors},
		},
	}
	c.SetFace(0)
	return c
}

// seatLandAdventure puts the card in the active seat's hand at main
// phase 1 and returns the game, seat and card ID.
func SeatLandAdventure(t *testing.T, oracle, landName, adventureName, cost string, colors []string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := LandAdventureCard(me.ID, oracle, landName, adventureName, cost, colors)
	me.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g, me, c.InstanceID
}

// assertLandPlayableFromExile checks CR 715.3d's grant is a play
// permission, then plays the land and checks it entered tapped.
func AssertLandPlayableFromExile(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) {
	t.Helper()
	if !g.Exile.Contains(id) {
		t.Fatal("CR 715.3d: the resolved Adventure is not in exile")
	}
	perm := g.CastPermissionOnCardByIDForEffect(id)
	if !perm.Granted() || perm.CastOnly {
		t.Fatalf("grant = %+v, want a play (not cast-only) permission", perm)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play the land from exile: %v", err)
	}
	var landed *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			landed = &g.Battlefield.Cards[i]
		}
	}
	if landed == nil {
		t.Fatal("the land did not reach the battlefield")
	}
	if !landed.IsLand() || !landed.Tapped {
		t.Errorf("land = %+v, want a land that entered tapped", landed)
	}
}
