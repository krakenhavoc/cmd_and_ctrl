package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// free_offer_combined_test.go — #4697: Ziatora's Envoy's "you may play a
// land ... or cast a spell with mana value N or less ... If you don't, put
// that card into your hand" over a card that gives more than one answer is
// one resolution_cast under MayCastKeywordPlayOrFreeOrHand: a land play of
// each face that can be played now (CR 305.2b, 712.12), the capped cast of
// a spell face (CR 601.3e, 712.11c, 715.3a), or the decline to the hand.
// The offer is made inside an open resolution, as the trigger makes it; the
// card-level tests are cards/effects/free_offer_combined_4697_test.go.

// combinedOfferOver puts c on top of the active player's library and makes
// the Envoy's offer, capped at 3, inside an open resolution.
func combinedOfferOver(t *testing.T, c Card) (*Game, *Player, uuid.UUID, *PendingChoice) {
	t.Helper()
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	c.Owner, c.Controller = me.ID, me.ID
	g.WithWriteLock(func() {
		me.Library.PushTop(c)
		g.resolutionOpen = true
		if err := g.OfferFreePlayFromTopOrHandForEffect(FreePlayFromTop{
			Player: me.ID, Source: uuid.New(), SourceName: "Test Envoy", MaxManaValue: 3, Land: true,
		}); err != nil {
			t.Fatalf("OfferFreePlayFromTopOrHandForEffect: %v", err)
		}
	})
	var offer *PendingChoice
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == PendingChoiceResolutionCast && pc.Chooser == me.ID {
			offer = pc
		}
	}
	if offer == nil || offer.MayCastKeyword != MayCastKeywordPlayOrFreeOrHand ||
		len(offer.DiscardOptions) != 1 || offer.DiscardOptions[0] != c.InstanceID {
		t.Fatalf("offer %+v (in hand %v), want the combined land play or cast", offer, me.Hand.Contains(c.InstanceID))
	}
	if mayCastChoiceFor(g, me.ID) != nil {
		t.Error("a second, land-only prompt was asked beside the combined one")
	}
	return g, me, c.InstanceID, offer
}

// cheapSpellLandMDFC is a modal DFC with a mana value 2 sorcery front and a
// land back.
func cheapSpellLandMDFC(owner uuid.UUID) Card {
	return mdfcFixture(owner,
		Face{Name: "Fixture Bloom", TypeLine: "Sorcery", ManaCost: "{1}{G}", Colors: []string{"G"}},
		Face{Name: "Fixture Grove", TypeLine: "Land"},
	)
}

// pathwayFixture is a pathway: two land faces.
func pathwayFixture(owner uuid.UUID) Card {
	return mdfcFixture(owner,
		Face{Name: "Fixture Dawn Pathway", TypeLine: "Land"},
		Face{Name: "Fixture Dusk Pathway", TypeLine: "Land"},
	)
}

// combinedAnswered checks the offer is gone with its permission.
func combinedAnswered(t *testing.T, g *Game, me *Player, id uuid.UUID) {
	t.Helper()
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == PendingChoiceResolutionCast && pc.Chooser == me.ID {
			t.Error("the offer is still open")
		}
	}
	if freeCastGrantOn(g, id) != nil {
		t.Error("the free cast outlived the answer")
	}
}

func TestCombinedFreePlayOfferAnswers(t *testing.T) {
	type answer struct {
		name string
		// face is the face sent; below zero is the decline.
		face int
		// land is the face expected on the battlefield, spell the face
		// expected on the stack; below zero for neither.
		land, spell int
	}
	for _, tc := range []struct {
		name    string
		card    func(uuid.UUID) Card
		answers []answer
	}{
		{"town", landAdventureFixture, []answer{
			{"play the land", 0, 0, -1},
			{"cast the Adventure", 1, -1, 1},
			{"decline", -1, -1, -1},
		}},
		{"spell front land back", cheapSpellLandMDFC, []answer{
			{"play the land back", 1, 1, -1},
			{"cast the spell front", 0, -1, 0},
			{"decline", -1, -1, -1},
		}},
		{"pathway", pathwayFixture, []answer{
			{"play the front", 0, 0, -1},
			{"play the back", 1, 1, -1},
			{"decline", -1, -1, -1},
		}},
	} {
		for _, a := range tc.answers {
			t.Run(tc.name+"/"+a.name, func(t *testing.T) {
				g, me, id, offer := combinedOfferOver(t, tc.card(uuid.Nil))
				if a.face < 0 {
					if err := g.ResolvePendingChoice(offer.ID, me.ID, nil); err != nil {
						t.Fatalf("decline: %v", err)
					}
				} else if err := g.CastForResolutionChoice(me.ID, offer.ID, id, CastSpellParams{FromZone: "library", Face: a.face}); err != nil {
					t.Fatalf("answer face %d: %v", a.face, err)
				}
				combinedAnswered(t, g, me, id)
				land := findCardForTest(g.Battlefield, id)
				spell := findCardForTest(g.Stack, id)
				played := g.LandsPlayedThisTurnFor(me.ID)
				switch {
				case a.land >= 0:
					if land == nil || land.ActiveFace != a.land || !land.IsLand() {
						t.Fatalf("on the battlefield: %+v, want face %d", land, a.land)
					}
					if played != 1 {
						t.Errorf("lands played this turn = %d, want 1", played)
					}
				case a.spell >= 0:
					if spell == nil || spell.ActiveFace != a.spell || spell.IsLand() {
						t.Fatalf("on the stack: %+v, want face %d as a spell", spell, a.spell)
					}
					if played != 0 {
						t.Errorf("lands played this turn = %d after a cast", played)
					}
				default:
					inHand := findCardForTest(me.Hand, id)
					if inHand == nil || inHand.ActiveFace != 0 || land != nil || spell != nil {
						t.Fatalf("declined: in hand %+v, battlefield %+v, stack %+v", inHand, land, spell)
					}
					if played != 0 {
						t.Errorf("lands played this turn = %d after a decline", played)
					}
				}
			})
		}
	}
}

// The land face is played only while the land drop is there, and a face
// the card does not have is no answer: both are refused with the offer
// left open.
func TestCombinedFreePlayOfferRefusesALandPlayWithNoDropLeft(t *testing.T) {
	g, me, id, offer := combinedOfferOver(t, pathwayFixture(uuid.Nil))
	g.WithWriteLock(func() { g.LandsPlayedThisTurn = map[uuid.UUID]int{me.ID: 1} })
	err := g.CastForResolutionChoice(me.ID, offer.ID, id, CastSpellParams{FromZone: "library", Face: 1})
	if !errors.Is(err, ErrLandDropUnavailable) {
		t.Fatalf("play with no land drop: %v, want ErrLandDropUnavailable", err)
	}
	if !me.Library.Contains(id) || freeCastGrantOn(g, id) == nil {
		t.Fatal("the refused play spent the offer")
	}
	if err := g.ResolvePendingChoice(offer.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if !me.Hand.Contains(id) {
		t.Fatal("the decline did not put the card into the hand")
	}
}

// A card with one answer is offered as before: a plain land, and a modal
// DFC whose spell face does not fit the cap, get the land offer's may_cast.
func TestCombinedFreePlayOfferIsOnlyForMoreThanOneAnswer(t *testing.T) {
	for _, c := range []Card{
		{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest"},
		spellFrontLandBackMDFC(uuid.Nil),
	} {
		g := newActiveGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		toMainPhase(t, g)
		c.Owner, c.Controller = me.ID, me.ID
		g.WithWriteLock(func() {
			me.Library.PushTop(c)
			g.resolutionOpen = true
			if err := g.OfferFreePlayFromTopOrHandForEffect(FreePlayFromTop{
				Player: me.ID, Source: uuid.New(), SourceName: "Test Envoy", MaxManaValue: 3, Land: true,
			}); err != nil {
				t.Fatalf("offer: %v", err)
			}
		})
		if m := mayCastChoiceFor(g, me.ID); m == nil || m.MayCastKeyword != MayCastKeywordPlayOrHand {
			t.Errorf("%s: offer %+v, want the land play alone", c.Name, m)
		}
	}
}
