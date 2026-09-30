package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spectator_public_view_test.go — #1588, ADR 0069's 2026-09-30
// amendment. A spectator learns only what every seat publicly knows.
// The admin keeps the omniscient debug view, which used to be the same
// view ("" was both).

func spectatorFixture(t *testing.T) (g *game.Game, me *game.Player, ids map[string]uuid.UUID) {
	t.Helper()
	g = buildActiveGame(t)
	me = g.Seats[0]
	ids = map[string]uuid.UUID{}
	all := []uuid.UUID{g.Seats[0].ID, g.Seats[1].ID}

	mk := func(label, name string, kind game.FaceDownKind) game.Card {
		c := game.NewCard(name, me.ID)
		c.TypeLine = "Creature — Secret"
		c.Power, c.Toughness = 3, 3
		if kind != game.FaceDownNone {
			c.SetFaceDown(kind)
		}
		ids[label] = c.InstanceID
		return c
	}

	foretold := mk("foretold", "Secret Foretold", game.FaceDownForetold)
	foretold.KnownBy = map[uuid.UUID]bool{me.ID: true}
	hideaway := mk("hideaway", "Secret Hideaway", game.FaceDownHidden)
	hideaway.KnownBy = map[uuid.UUID]bool{me.ID: true}
	morph := mk("morph", "Secret Morph", game.FaceDownMorphed)
	morph.KnownBy = map[uuid.UUID]bool{me.ID: true}
	necro := mk("necro", "Secret Necro", game.FaceDownExiled)
	hand := mk("hand", "Secret Hand Card", game.FaceDownNone)
	hand.KnownBy = map[uuid.UUID]bool{me.ID: true}
	public := mk("public", "Public Bear", game.FaceDownNone)
	public.AddKnowersAll(all)

	g.WithWriteLock(func() {
		g.Exile.PushTop(foretold)
		g.Exile.PushTop(hideaway)
		g.Exile.PushTop(necro)
		g.Battlefield.PushTop(morph)
		g.Battlefield.PushTop(public)
		me.Hand.PushTop(hand)
	})
	return g, me, ids
}

func findInView(v GameView, id uuid.UUID) (CardView, bool) {
	zones := []ZoneView{v.Exile, v.Battlefield, v.Stack}
	for _, s := range v.Seats {
		zones = append(zones, s.Hand, s.Library, s.Graveyard, s.Command)
	}
	for _, z := range zones {
		for _, c := range z.Cards {
			if c.InstanceID == id.String() {
				return c, true
			}
		}
	}
	return CardView{}, false
}

func TestSpectatorSeesPublicInformationOnly(t *testing.T) {
	g, me, ids := spectatorFixture(t)
	raw := ViewOfGame(g)

	spectator := FilterViewFor(raw, SpectatorViewerID)
	for _, label := range []string{"foretold", "hideaway", "morph", "necro"} {
		c, ok := findInView(spectator, ids[label])
		if !ok {
			t.Fatalf("%s: missing from the spectator view", label)
		}
		if !c.FaceDown {
			t.Errorf("%s: face_down must stay public", label)
		}
		if c.FaceVisible {
			t.Errorf("%s: a spectator may look at the face", label)
		}
		if label == "morph" {
			// A face-down permanent ships its public 2/2 body (CR 708.2),
			// which a non-knower is meant to see; the card under it is
			// what must be gone.
			if c.KnownByYou || c.Name != "" || c.ScryfallID != "" || c.ManaCost != "" {
				t.Errorf("spectator can identify the morph: %+v", c)
			}
			continue
		}
		assertRedacted(t, "spectator / "+label, c)
	}
	// A hand is not public: not the card, not a handle on it.
	if c, ok := findInView(spectator, ids["hand"]); ok {
		t.Errorf("spectator was handed a hand card: %+v", c)
	}
	if got := spectator.Seats[0].Hand.Count; got != raw.Seats[0].Hand.Count {
		t.Errorf("hand count = %d, want the public size %d", got, raw.Seats[0].Hand.Count)
	}
	// A card every seat knows is still shown in full.
	if c, ok := findInView(spectator, ids["public"]); !ok || c.Name != "Public Bear" || !c.KnownByYou {
		t.Errorf("a public card was redacted for the spectator: %+v", c)
	}

	// The controller still reads its own.
	mine := FilterViewFor(raw, me.ID.String())
	for _, label := range []string{"foretold", "hideaway", "morph"} {
		if c, _ := findInView(mine, ids[label]); !c.KnownByYou || c.Name == "" && label != "morph" {
			t.Errorf("%s: the owner lost their own card: %+v", label, c)
		}
	}
}

func TestAdminKeepsTheOmniscientView(t *testing.T) {
	g, _, ids := spectatorFixture(t)
	admin := FilterViewFor(ViewOfGame(g), "")
	for _, label := range []string{"foretold", "hideaway", "necro"} {
		c, ok := findInView(admin, ids[label])
		if !ok || !c.KnownByYou || c.Name == "" {
			t.Errorf("%s: admin lost the debug view: %+v", label, c)
		}
	}
	if c, ok := findInView(admin, ids["morph"]); !ok || !c.KnownByYou || !c.FaceVisible {
		t.Errorf("morph: admin cannot look at the face: %+v", c)
	}
	// Hands stay hidden wholesale even for the admin (unchanged).
	if c, ok := findInView(admin, ids["hand"]); ok {
		t.Errorf("admin was handed a hand card: %+v", c)
	}
}

// A spectator is never a seat: the sentinel cannot be a player UUID,
// and it gets no seat's move list or private prompt.
func TestSpectatorViewerIDIsNotASeat(t *testing.T) {
	if _, err := uuid.Parse(SpectatorViewerID); err == nil {
		t.Fatalf("SpectatorViewerID %q parses as a UUID; a seat could collide with it", SpectatorViewerID)
	}
	g := buildActiveGame(t)
	if moves := FilterViewFor(ViewOfGame(g), SpectatorViewerID).LegalMoves; len(moves) != 0 {
		t.Errorf("spectator got %d legal moves", len(moves))
	}
}
