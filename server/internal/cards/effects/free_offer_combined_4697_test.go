package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// free_offer_combined_4697_test.go — #4697, through Ziatora's Envoy's real
// trigger: "you may play a land from the top of your library or cast a
// spell with mana value less than or equal to the damage dealt from the top
// of your library without paying its mana cost. If you don't, put that card
// into your hand." A card that gives the player more than one of those
// answers is one offer, the resolution_cast under play_or_free_or_hand,
// that takes the land play of each land face that can be played now
// (CR 305.2b, 712.12) or the capped cast of a spell face (CR 601.3e,
// 712.11c, 715.3a), and declining puts the card into the hand. Envoy's
// damage is 5.

// envoyTown is a Town whose mana value 2 Adventure fits the cap.
func envoyTown() game.Card {
	return landAdventureCard(uuid.Nil, "", "Fixture Town", "Fixture Overture", "{1}{U}", []string{"U"})
}

// envoySpellLandMDFC is a modal DFC whose front face is a mana value 2
// sorcery and whose back face is a land.
func envoySpellLandMDFC() game.Card {
	return mdfcCard(uuid.Nil, "", game.LayoutModalDFC,
		game.Face{Name: "Fixture Bloom", TypeLine: "Sorcery", ManaCost: "{1}{G}", Colors: []string{"G"}},
		game.Face{Name: "Fixture Grove", TypeLine: "Land"},
	)
}

// envoyPathway is a pathway: two land faces.
func envoyPathway() game.Card {
	return mdfcCard(uuid.Nil, "", game.LayoutModalDFC,
		game.Face{Name: "Fixture Dawn Pathway", TypeLine: "Land"},
		game.Face{Name: "Fixture Dusk Pathway", TypeLine: "Land"},
	)
}

// combinedOffer attacks and returns the combined offer over id, held
// inside the trigger's resolution.
func (tb *envoyTable) combinedOffer(t *testing.T, id uuid.UUID) *game.PendingChoice {
	t.Helper()
	tb.attack(t)
	c := keywordCastPromptFor(tb.g, tb.me.ID)
	if c == nil || c.MayCastKeyword != game.MayCastKeywordPlayOrFreeOrHand || keywordCastCard(c) != id {
		t.Fatalf("offer %+v (in hand %v), want the combined land play or cast", c, tb.me.Hand.Contains(id))
	}
	if !tb.g.ResolutionPaused() {
		t.Fatal("the offer is not held inside the trigger's resolution")
	}
	return c
}

// settled checks the offer is answered and its permission gone.
func (tb *envoyTable) settled(t *testing.T, id uuid.UUID) {
	t.Helper()
	if keywordCastPromptFor(tb.g, tb.me.ID) != nil || tb.g.ResolutionPaused() {
		t.Error("the offer is still open")
	}
	var perm *game.CastPermission
	tb.g.ReadSnapshot(func() { perm = tb.g.CastPermissionOnCardByIDForEffect(id) })
	if perm != nil {
		t.Error("the free cast outlived the answer")
	}
}

// playFace plays the card as face through the offer.
func (tb *envoyTable) playFace(t *testing.T, c *game.PendingChoice, id uuid.UUID, face int, name string) {
	t.Helper()
	if err := tb.g.CastForResolutionChoice(tb.me.ID, c.ID, id, game.CastSpellParams{FromZone: "library", Face: face}); err != nil {
		t.Fatalf("play face %d: %v", face, err)
	}
	land, ok := battlefieldCard(tb.g, id)
	if !ok || land.ActiveFace != face || !land.IsLand() || land.Name != name {
		t.Fatalf("on the battlefield: %+v, want %s face up", land, name)
	}
	if got := tb.g.LandsPlayedThisTurnFor(tb.me.ID); got != 1 {
		t.Errorf("lands played this turn = %d, want 1", got)
	}
	tb.settled(t, id)
}

// castFace casts face through the offer.
func (tb *envoyTable) castFace(t *testing.T, c *game.PendingChoice, id uuid.UUID, face int) game.Card {
	t.Helper()
	if err := tb.g.CastForResolutionChoice(tb.me.ID, c.ID, id, game.CastSpellParams{FromZone: "library", Face: face}); err != nil {
		t.Fatalf("cast face %d: %v", face, err)
	}
	spell, ok := stackCardOf(tb.g, id)
	if !ok || spell.ActiveFace != face || spell.IsLand() {
		t.Fatalf("on the stack: %+v, want face %d as a spell", spell, face)
	}
	if got := tb.g.LandsPlayedThisTurnFor(tb.me.ID); got != 0 {
		t.Errorf("lands played this turn = %d, want 0 after a cast", got)
	}
	tb.settled(t, id)
	return spell
}

// decline answers the offer "neither": the card goes into the hand.
func (tb *envoyTable) decline(t *testing.T, c *game.PendingChoice, id uuid.UUID) {
	t.Helper()
	if err := tb.g.ResolvePendingChoice(c.ID, tb.me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if !tb.me.Hand.Contains(id) || tb.me.Library.Contains(id) {
		t.Fatal("the declined card is not in the hand")
	}
	if c, _ := battlefieldCard(tb.g, id); c.InstanceID == id {
		t.Error("a declined card reached the battlefield")
	}
	if got := tb.g.LandsPlayedThisTurnFor(tb.me.ID); got != 0 {
		t.Errorf("lands played this turn = %d, want 0 after a decline", got)
	}
	tb.settled(t, id)
}

func TestZiatorasEnvoyCombinedTownPlaysItsLand(t *testing.T) {
	tb := newEnvoyTable(t)
	id := tb.pushTop(envoyTown())
	c := tb.combinedOffer(t, id)
	tb.playFace(t, c, id, 0, "Fixture Town")
}

func TestZiatorasEnvoyCombinedTownCastsItsAdventure(t *testing.T) {
	tb := newEnvoyTable(t)
	id := tb.pushTop(envoyTown())
	c := tb.combinedOffer(t, id)
	if spell := tb.castFace(t, c, id, 1); !spell.IsSorcery() {
		t.Errorf("cast %+v, want the sorcery Adventure", spell)
	}
}

func TestZiatorasEnvoyCombinedTownDeclinedGoesToHand(t *testing.T) {
	tb := newEnvoyTable(t)
	id := tb.pushTop(envoyTown())
	tb.decline(t, tb.combinedOffer(t, id), id)
}

func TestZiatorasEnvoyCombinedModalDFCPlaysItsLandBack(t *testing.T) {
	tb := newEnvoyTable(t)
	id := tb.pushTop(envoySpellLandMDFC())
	c := tb.combinedOffer(t, id)
	tb.playFace(t, c, id, 1, "Fixture Grove")
}

func TestZiatorasEnvoyCombinedModalDFCCastsItsSpellFront(t *testing.T) {
	tb := newEnvoyTable(t)
	id := tb.pushTop(envoySpellLandMDFC())
	c := tb.combinedOffer(t, id)
	if spell := tb.castFace(t, c, id, 0); !spell.IsSorcery() || spell.Name != "Fixture Bloom" {
		t.Errorf("cast %+v, want the sorcery front face", spell)
	}
}

func TestZiatorasEnvoyCombinedModalDFCDeclinedGoesToHand(t *testing.T) {
	tb := newEnvoyTable(t)
	id := tb.pushTop(envoySpellLandMDFC())
	tb.decline(t, tb.combinedOffer(t, id), id)
	if card, ok := envoyHandCard(tb.me, id); !ok || card.ActiveFace != 0 {
		t.Errorf("in hand: %+v, want the card front face up", card)
	}
}

func TestZiatorasEnvoyCombinedPathwayPlaysItsFront(t *testing.T) {
	tb := newEnvoyTable(t)
	id := tb.pushTop(envoyPathway())
	c := tb.combinedOffer(t, id)
	tb.playFace(t, c, id, 0, "Fixture Dawn Pathway")
}

func TestZiatorasEnvoyCombinedPathwayPlaysItsBack(t *testing.T) {
	tb := newEnvoyTable(t)
	id := tb.pushTop(envoyPathway())
	c := tb.combinedOffer(t, id)
	tb.playFace(t, c, id, 1, "Fixture Dusk Pathway")
}

func TestZiatorasEnvoyCombinedPathwayDeclinedGoesToHand(t *testing.T) {
	tb := newEnvoyTable(t)
	id := tb.pushTop(envoyPathway())
	tb.decline(t, tb.combinedOffer(t, id), id)
}

// The enumerator offers each answer, each naming the prompt: a pathway's
// two land plays and no cast; a Town's land play and its Adventure's cast.
func TestZiatorasEnvoyCombinedOfferEnumeratesEachAnswer(t *testing.T) {
	for _, tc := range []struct {
		name  string
		card  game.Card
		plays []string
		casts int
	}{
		{"pathway", envoyPathway(), []string{"Play Fixture Dawn Pathway", "Play Fixture Dusk Pathway"}, 0},
		{"town", envoyTown(), []string{"Play Fixture Town"}, 1},
		{"mdfc", envoySpellLandMDFC(), []string{"Play Fixture Grove"}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tb := newEnvoyTable(t)
			id := tb.pushTop(tc.card)
			c := tb.combinedOffer(t, id)
			var plays []string
			casts, declines := 0, 0
			for _, mv := range legal.EnumerateFor(tb.g, tb.me.ID) {
				if mv.ResolutionCast != c.ID.String() {
					t.Errorf("move %q does not answer the offer", mv.Label)
					continue
				}
				switch {
				case mv.Kind == legal.KindLand && mv.Source == id:
					plays = append(plays, mv.Label)
				case mv.Kind == legal.KindCast && mv.Source == id:
					casts++
				case mv.Kind == legal.KindChoice && mv.AlwaysLegal:
					declines++
				}
			}
			want := make([]string, len(tc.plays))
			for i, p := range tc.plays {
				want[i] = "Ziatora's Envoy: " + p
			}
			if len(plays) != len(want) {
				t.Fatalf("land plays %q, want %q", plays, want)
			}
			for i := range want {
				if plays[i] != want[i] {
					t.Errorf("land play %d = %q, want %q", i, plays[i], want[i])
				}
			}
			if (casts > 0) != (tc.casts > 0) || declines != 1 {
				t.Errorf("offered %d casts and %d declines, want casts: %v and one decline", casts, declines, tc.casts > 0)
			}
		})
	}
}

// The land drop spent after the offer: the play is refused and the offer
// stays open for the cast or the decline.
func TestZiatorasEnvoyCombinedLandPlayWithNoDropLeftIsRefused(t *testing.T) {
	tb := newEnvoyTable(t)
	id := tb.pushTop(envoyTown())
	c := tb.combinedOffer(t, id)
	tb.g.WithWriteLock(func() { tb.g.LandsPlayedThisTurn = map[uuid.UUID]int{tb.me.ID: 1} })
	if err := tb.g.CastForResolutionChoice(tb.me.ID, c.ID, id, game.CastSpellParams{FromZone: "library", Face: 0}); err == nil {
		t.Fatal("a land was played with no land drop left")
	}
	if keywordCastPromptFor(tb.g, tb.me.ID) == nil || !tb.me.Library.Contains(id) {
		t.Fatal("the refused play closed the offer")
	}
	if err := tb.g.CastForResolutionChoice(tb.me.ID, c.ID, id, game.CastSpellParams{FromZone: "library", Face: 1}); err != nil {
		t.Fatalf("cast the Adventure after the refused play: %v", err)
	}
	if spell, ok := stackCardOf(tb.g, id); !ok || !spell.IsSorcery() {
		t.Fatalf("on the stack: %+v, want the Adventure", spell)
	}
	tb.settled(t, id)
}

// The offer is a restore point: the restored table plays a pathway's
// back face.
func TestZiatorasEnvoyCombinedOfferSurvivesARestore(t *testing.T) {
	tb := newEnvoyTable(t)
	id := tb.pushTop(envoyPathway())
	tb.combinedOffer(t, id)
	restored := restoreRoundTrip(t, tb.g, true)
	rme := restored.PlayerByIDForEffect(tb.me.ID)
	rc := keywordCastPromptFor(restored, rme.ID)
	if rc == nil || rc.MayCastKeyword != game.MayCastKeywordPlayOrFreeOrHand || !restored.ResolutionPaused() {
		t.Fatalf("the restored table lost the offer: %+v", rc)
	}
	if err := restored.CastForResolutionChoice(rme.ID, rc.ID, id, game.CastSpellParams{FromZone: "library", Face: 1}); err != nil {
		t.Fatalf("the restored play: %v", err)
	}
	if land, ok := battlefieldCard(restored, id); !ok || land.ActiveFace != 1 || land.Name != "Fixture Dusk Pathway" {
		t.Fatalf("on the restored battlefield: %+v, want the back face", land)
	}
	if restored.ResolutionPaused() {
		t.Error("the restored resolution is still held")
	}
}

func envoyHandCard(p *game.Player, id uuid.UUID) (game.Card, bool) {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}
