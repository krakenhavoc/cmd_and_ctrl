package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// permission_spell_face_view_test.go — #4509's view half. A permission
// phrased about the spell, or the land played, opens only the faces it
// admits (CR 601.3e, 712.11c, 715.3a), and the view stamps the cast
// surface on exactly those: the face the client's picker and pile button
// read (`faces[i].castable_here`), not the card, which shows its front.

// mainPhaseForView advances a fresh game to its first main phase.
func mainPhaseForView(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// twoFaced is a multi-face card of `layout` with the faces given, face 0
// materialised, as the deck importer builds one.
func twoFaced(owner uuid.UUID, layout string, front, back game.Face) game.Card {
	c := game.Card{InstanceID: uuid.New(), Owner: owner, Controller: owner, Layout: layout,
		Faces: []game.Face{front, back}}
	c.SetFace(0)
	return c
}

// Oracle of Mul Daya's shape, "play lands from the top of your library",
// over a modal DFC on top whose back face is a land: the land face is the
// one stamped playable.
func TestViewStampsAModalDFCsLandFaceUnderAPlayLandsPermission(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mainPhaseForView(t, g)
	const oracle = "test-view-mul-daya"
	withStandingPermission(t, oracle, game.CastPermission{
		Zone: game.ZoneLibrary, Scope: game.ScopeStanding, Duration: game.WhileInZoneDuration(),
		TopOfLibraryOnly: true, Filter: game.PermissionFilter{LandsOnly: true},
	})
	withTopVisibility(t, oracle, game.LibraryTopRevealed)
	source := game.NewCard("Test Oracle", me.ID)
	source.TypeLine = "Creature — Elf Shaman"
	source.OracleID = oracle
	source.Controller = me.ID
	g.Battlefield.PushTop(source)

	mdfc := twoFaced(me.ID, game.LayoutModalDFC,
		game.Face{Name: "Test Call", TypeLine: "Sorcery", ManaCost: "{0}", Colors: []string{"W"}},
		game.Face{Name: "Test Skyclave", TypeLine: "Land"},
	)
	me.Library.PushTop(mdfc)

	v := ViewOfGameFor(g, me.ID.String())
	top := cardInZone(v.Seats[g.Turn.ActiveSeat].Library, mdfc.InstanceID)
	if top == nil || len(top.Faces) != 2 {
		t.Fatalf("the top card is missing or has no faces: %+v", top)
	}
	if top.CastableHere {
		t.Error("the sorcery front is stamped castable under a play-lands permission")
	}
	if top.Faces[0].CastableHere {
		t.Error("faces[0], the sorcery, is stamped castable")
	}
	if !top.Faces[1].CastableHere {
		t.Error("faces[1], the land, is not stamped playable")
	}
}

// "Cast an instant or sorcery spell from your graveyard" over an
// adventurer card: the Adventure is the face stamped, and the creature
// half carries no cast surface.
func TestViewStampsOnlyTheAdventureUnderAnInstantOrSorceryPermission(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mainPhaseForView(t, g)
	const oracle = "test-view-kess"
	withStandingPermission(t, oracle, game.CastPermission{
		Zone: game.ZoneGraveyard, Scope: game.ScopeStanding, Duration: game.WhileInZoneDuration(),
		Filter: game.PermissionFilter{InstantOrSorceryOnly: true},
	})
	source := game.NewCard("Test Mage", me.ID)
	source.TypeLine = "Creature — Human Wizard"
	source.OracleID = oracle
	source.Controller = me.ID
	g.Battlefield.PushTop(source)

	seen := map[uuid.UUID]bool{me.ID: true, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID: true}
	adventure := twoFaced(me.ID, game.LayoutAdventure,
		game.Face{Name: "Test Knight", TypeLine: "Creature — Human Knight", ManaCost: "{0}", Colors: []string{"B"},
			Power: 1, Toughness: 1},
		game.Face{Name: "Test Insight", TypeLine: "Instant — Adventure", ManaCost: "{0}", Colors: []string{"B"}},
	)
	adventure.KnownBy = seen
	me.Graveyard.PushTop(adventure)

	v := ViewOfGameFor(g, me.ID.String())
	got := cardInZone(v.Seats[g.Turn.ActiveSeat].Graveyard, adventure.InstanceID)
	if got == nil || len(got.Faces) != 2 {
		t.Fatalf("the graveyard card is missing or has no faces: %+v", got)
	}
	if got.CastableHere || got.Faces[0].CastableHere {
		t.Error("the creature half is stamped castable under an instant-or-sorcery permission")
	}
	if !got.Faces[1].CastableHere {
		t.Error("faces[1], the Adventure, is not stamped castable")
	}
}
