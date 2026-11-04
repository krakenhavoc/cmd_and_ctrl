package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// permission_per_face_view_test.go — #4696's view half. Two permissions
// open different faces of one adventurer card in a graveyard ("creature
// spells" and "instant and sorcery spells"): both faces are stamped
// castable (CR 601.3, 715.3a), whichever permission the lookup finds
// first, and the face the pile shows carries the card-level stamp.

func TestViewStampsBothFacesUnderTwoPermissionsThatOpenOneEach(t *testing.T) {
	const creatureOracle, spellsOracle = "test-view-4696-creatures", "test-view-4696-spells"
	perms := map[string]game.CastPermission{
		creatureOracle: {Zone: game.ZoneGraveyard, Scope: game.ScopeStanding, Duration: game.WhileInZoneDuration(),
			Filter: game.PermissionFilter{CreatureOnly: true}},
		spellsOracle: {Zone: game.ZoneGraveyard, Scope: game.ScopeStanding, Duration: game.WhileInZoneDuration(),
			Filter: game.PermissionFilter{InstantOrSorceryOnly: true}},
	}
	prev := game.CatalogCastPermissions
	game.CatalogCastPermissions = func(id string) []game.CastPermission {
		if p, ok := perms[id]; ok {
			return []game.CastPermission{p}
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogCastPermissions = prev })

	for _, order := range [][]string{{creatureOracle, spellsOracle}, {spellsOracle, creatureOracle}} {
		g := buildActiveGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		mainPhaseForView(t, g)
		for _, oracle := range order {
			source := game.NewCard("Test Source "+oracle, me.ID)
			source.TypeLine = "Creature — Human Wizard"
			source.OracleID = oracle
			source.Controller = me.ID
			g.Battlefield.PushTop(source)
		}
		adventure := twoFaced(me.ID, game.LayoutAdventure,
			game.Face{Name: "Test Knight", TypeLine: "Creature — Human Knight", ManaCost: "{0}", Colors: []string{"B"},
				Power: 1, Toughness: 1},
			game.Face{Name: "Test Insight", TypeLine: "Instant — Adventure", ManaCost: "{0}", Colors: []string{"B"}},
		)
		adventure.KnownBy = map[uuid.UUID]bool{me.ID: true}
		me.Graveyard.PushTop(adventure)

		v := ViewOfGameFor(g, me.ID.String())
		got := cardInZone(v.Seats[g.Turn.ActiveSeat].Graveyard, adventure.InstanceID)
		if got == nil || len(got.Faces) != 2 {
			t.Fatalf("order %v: the graveyard card is missing or has no faces: %+v", order, got)
		}
		if !got.Faces[0].CastableHere {
			t.Errorf("order %v: faces[0], the creature, is not stamped castable", order)
		}
		if !got.Faces[1].CastableHere {
			t.Errorf("order %v: faces[1], the Adventure, is not stamped castable", order)
		}
		if !got.CastableHere {
			t.Errorf("order %v: the face the pile shows (the creature) is not stamped castable", order)
		}
	}
}
