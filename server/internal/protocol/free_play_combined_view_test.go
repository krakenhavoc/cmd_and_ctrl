package protocol

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// free_play_combined_view_test.go — #4697. Ziatora's Envoy's offer over a
// card that gives more than one answer (here a pathway, two land faces) is
// the resolution_cast under play_or_free_or_hand. Its chooser sees the card,
// the keyword and, in legal_moves, a land play of each face, each a
// cast_spell naming the prompt and the face; nobody else sees the card or
// is offered a move. The engine raises the prompt only inside an open
// resolution, so the permission is stamped and the prompt queued by hand,
// as free_play_from_top_view_test.go does.
func TestACombinedTopCardOfferListsALandPlayPerFace(t *testing.T) {
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	withTopVisibility(t, "unused", game.LibraryTopHidden)
	pathway := game.Card{
		InstanceID: uuid.New(), Layout: game.LayoutModalDFC, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Fixture Dawn Pathway", TypeLine: "Land"},
			{Name: "Fixture Dusk Pathway", TypeLine: "Land"},
		},
	}
	pathway.SetFace(0)
	me.Library.PushTop(pathway)
	g.WithWriteLock(func() {
		g.LookAtTopOfLibraryForEffect(me.ID, 1)
		limit := 3
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player: me.ID, Zone: game.ZoneLibrary, Duration: g.UntilEndOfTurnDuration(),
			Cost: "{0}", Timing: game.TimingFlash, CastOnly: true, MaxSpellManaValue: &limit,
			LapseOnPass: game.LapseToHand, SourceName: "Test Envoy",
			Label: "Test Envoy — cast it without paying its mana cost",
		}, []game.Card{pathway})
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind: game.PendingChoiceResolutionCast, Chooser: me.ID, Count: 1,
			Reason: "Test Envoy", DiscardOptions: []uuid.UUID{pathway.InstanceID},
			MayCastKeyword: game.MayCastKeywordPlayOrFreeOrHand,
			AcceptLabel:    "Play or cast it", DeclineLabel: "Put it into your hand",
		})
	})
	full := ViewOfGame(g)
	own := FilterViewFor(full, me.ID.String())
	c := choiceFor(t, own, string(game.PendingChoiceResolutionCast))
	if c.MayCastKeyword != "play_or_free_or_hand" || c.DeclineLabel != "Put it into your hand" ||
		len(c.Options) != 1 || c.Options[0].Name != "Fixture Dawn Pathway" {
		t.Fatalf("the chooser's offer: %+v", c)
	}
	faces := map[int]string{}
	declines := 0
	for _, m := range own.LegalMoves {
		if m.ResolutionCast != c.ID {
			t.Errorf("move %q does not answer the offer", m.Label)
			continue
		}
		if m.Type == legal.TypeResolveChoice {
			declines++
			continue
		}
		var p struct {
			InstanceID     string `json:"instance_id"`
			FromZone       string `json:"from_zone"`
			Face           int    `json:"face"`
			ResolutionCast string `json:"resolution_cast"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("params of %q: %v", m.Label, err)
		}
		if m.Type != legal.TypeCastSpell || m.Kind != legal.KindLand || p.InstanceID != pathway.InstanceID.String() ||
			p.FromZone != "library" || p.ResolutionCast != c.ID {
			t.Errorf("move %q: %+v (%s %s)", m.Label, p, m.Type, m.Kind)
		}
		faces[p.Face] = m.Label
	}
	if len(faces) != 2 || faces[0] != "Test Envoy: Play Fixture Dawn Pathway" ||
		faces[1] != "Test Envoy: Play Fixture Dusk Pathway" || declines != 1 {
		t.Errorf("land plays %v and %d declines, want one per face and the decline", faces, declines)
	}
	other := FilterViewFor(full, them.ID.String())
	oc := choiceFor(t, other, string(game.PendingChoiceResolutionCast))
	if len(oc.Options) != 0 || len(oc.Eligible) != 0 {
		t.Errorf("an opponent's copy leaks the looked-at card: %+v", oc)
	}
	for _, m := range other.LegalMoves {
		if m.ResolutionCast != "" {
			t.Errorf("an opponent is offered %q", m.Label)
		}
	}
}
