package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	return effectstest.LandAdventureCard(uuid.Nil, "", "Fixture Town", "Fixture Overture", "{1}{U}", []string{"U"})
}

// envoySpellLandMDFC is a modal DFC whose front face is a mana value 2
// sorcery and whose back face is a land.
func envoySpellLandMDFC() game.Card {
	return effectstest.MdfcCard(uuid.Nil, "", game.LayoutModalDFC,
		game.Face{Name: "Fixture Bloom", TypeLine: "Sorcery", ManaCost: "{1}{G}", Colors: []string{"G"}},
		game.Face{Name: "Fixture Grove", TypeLine: "Land"},
	)
}

// envoyPathway is a pathway: two land faces.
func envoyPathway() game.Card {
	return effectstest.MdfcCard(uuid.Nil, "", game.LayoutModalDFC,
		game.Face{Name: "Fixture Dawn Pathway", TypeLine: "Land"},
		game.Face{Name: "Fixture Dusk Pathway", TypeLine: "Land"},
	)
}

func TestZiatorasEnvoyCombinedTownPlaysItsLand(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	id := tb.PushTop(envoyTown())
	c := tb.CombinedOffer(t, id)
	tb.PlayFace(t, c, id, 0, "Fixture Town")
}

func TestZiatorasEnvoyCombinedTownCastsItsAdventure(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	id := tb.PushTop(envoyTown())
	c := tb.CombinedOffer(t, id)
	if spell := tb.CastFace(t, c, id, 1); !spell.IsSorcery() {
		t.Errorf("cast %+v, want the sorcery Adventure", spell)
	}
}

func TestZiatorasEnvoyCombinedTownDeclinedGoesToHand(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	id := tb.PushTop(envoyTown())
	tb.Decline(t, tb.CombinedOffer(t, id), id)
}

func TestZiatorasEnvoyCombinedModalDFCPlaysItsLandBack(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	id := tb.PushTop(envoySpellLandMDFC())
	c := tb.CombinedOffer(t, id)
	tb.PlayFace(t, c, id, 1, "Fixture Grove")
}

func TestZiatorasEnvoyCombinedModalDFCCastsItsSpellFront(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	id := tb.PushTop(envoySpellLandMDFC())
	c := tb.CombinedOffer(t, id)
	if spell := tb.CastFace(t, c, id, 0); !spell.IsSorcery() || spell.Name != "Fixture Bloom" {
		t.Errorf("cast %+v, want the sorcery front face", spell)
	}
}

func TestZiatorasEnvoyCombinedModalDFCDeclinedGoesToHand(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	id := tb.PushTop(envoySpellLandMDFC())
	tb.Decline(t, tb.CombinedOffer(t, id), id)
	if card, ok := envoyHandCard(tb.Me, id); !ok || card.ActiveFace != 0 {
		t.Errorf("in hand: %+v, want the card front face up", card)
	}
}

func TestZiatorasEnvoyCombinedPathwayPlaysItsFront(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	id := tb.PushTop(envoyPathway())
	c := tb.CombinedOffer(t, id)
	tb.PlayFace(t, c, id, 0, "Fixture Dawn Pathway")
}

func TestZiatorasEnvoyCombinedPathwayPlaysItsBack(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	id := tb.PushTop(envoyPathway())
	c := tb.CombinedOffer(t, id)
	tb.PlayFace(t, c, id, 1, "Fixture Dusk Pathway")
}

func TestZiatorasEnvoyCombinedPathwayDeclinedGoesToHand(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	id := tb.PushTop(envoyPathway())
	tb.Decline(t, tb.CombinedOffer(t, id), id)
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
			tb := effectstest.NewEnvoyTable(t)
			id := tb.PushTop(tc.card)
			c := tb.CombinedOffer(t, id)
			var plays []string
			casts, declines := 0, 0
			for _, mv := range legal.EnumerateFor(tb.G, tb.Me.ID) {
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
	tb := effectstest.NewEnvoyTable(t)
	id := tb.PushTop(envoyTown())
	c := tb.CombinedOffer(t, id)
	tb.G.WithWriteLock(func() { tb.G.LandsPlayedThisTurn = map[uuid.UUID]int{tb.Me.ID: 1} })
	if err := tb.G.CastForResolutionChoice(tb.Me.ID, c.ID, id, game.CastSpellParams{FromZone: "library", Face: 0}); err == nil {
		t.Fatal("a land was played with no land drop left")
	}
	if effectstest.KeywordCastPromptFor(tb.G, tb.Me.ID) == nil || !tb.Me.Library.Contains(id) {
		t.Fatal("the refused play closed the offer")
	}
	if err := tb.G.CastForResolutionChoice(tb.Me.ID, c.ID, id, game.CastSpellParams{FromZone: "library", Face: 1}); err != nil {
		t.Fatalf("cast the Adventure after the refused play: %v", err)
	}
	if spell, ok := effectstest.StackCardOf(tb.G, id); !ok || !spell.IsSorcery() {
		t.Fatalf("on the stack: %+v, want the Adventure", spell)
	}
	tb.Settled(t, id)
}

// The offer is a restore point: the restored table plays a pathway's
// back face.
func TestZiatorasEnvoyCombinedOfferSurvivesARestore(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	id := tb.PushTop(envoyPathway())
	tb.CombinedOffer(t, id)
	restored := effectstest.RestoreRoundTrip(t, tb.G, true)
	rme := restored.PlayerByIDForEffect(tb.Me.ID)
	rc := effectstest.KeywordCastPromptFor(restored, rme.ID)
	if rc == nil || rc.MayCastKeyword != game.MayCastKeywordPlayOrFreeOrHand || !restored.ResolutionPaused() {
		t.Fatalf("the restored table lost the offer: %+v", rc)
	}
	if err := restored.CastForResolutionChoice(rme.ID, rc.ID, id, game.CastSpellParams{FromZone: "library", Face: 1}); err != nil {
		t.Fatalf("the restored play: %v", err)
	}
	if land, ok := effectstest.BattlefieldCard(restored, id); !ok || land.ActiveFace != 1 || land.Name != "Fixture Dusk Pathway" {
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
