package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// land_mana_color_test.go pins Hall of Gemstone's replacement (#1997):
// what a land tapped for mana makes under it, and how the auto-tap
// planner prices it. The card is pinned in cards/effects.

func landManaColorGame(t *testing.T) (*Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	me := g.Seats[0].ID
	land := uuid.New()
	rock := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{InstanceID: land, Name: "Land", TypeLine: "Land", Owner: me, Controller: me})
		g.Battlefield.PushTop(Card{InstanceID: rock, Name: "Rock", TypeLine: "Artifact", Owner: me, Controller: me})
		if !g.LandsTappedProduceColorThisTurnForEffect(uuid.Nil, me, "R", "Hall of Gemstone") {
			t.Fatal("the replacement was not registered")
		}
	})
	return g, land, rock
}

// Every coloured mana of a land's tap becomes the chosen colour, one for
// one; colourless stays (CR 106.1a, 106.1b), and so does the chosen
// colour. A production that is not a tap, or not a land's, is untouched.
func TestLandManaIsColorRewritesOnlyColouredManaFromALandsTap(t *testing.T) {
	g, land, rock := landManaColorGame(t)
	me := g.Seats[0].ID
	for _, tc := range []struct {
		name    string
		source  uuid.UUID
		fromTap bool
		in      []string
		want    []string
	}{
		{"a Forest", land, true, []string{"G"}, []string{"R"}},
		{"a Karoo's two colours", land, true, []string{"W", "U"}, []string{"R", "R"}},
		{"colourless stays", land, true, []string{"C", "G"}, []string{"C", "R"}},
		{"already red", land, true, []string{"R"}, []string{"R"}},
		{"not tapped for mana", land, false, []string{"G"}, []string{"G"}},
		{"not a land", rock, true, []string{"G"}, []string{"G"}},
	} {
		var got []string
		g.ReadSnapshot(func() { got = g.producedManaPreviewLocked(me, tc.source, tc.in, tc.fromTap) })
		if !slices.Equal(got, tc.want) {
			t.Errorf("%s: %v became %v, want %v", tc.name, tc.in, got, tc.want)
		}
	}
}

// The planner books what the land will make: a {G} slot is an {R} slot,
// and a {W}/{U} pick is one {R} option.
func TestLandManaIsColorIsPricedForThePlanner(t *testing.T) {
	g, land, _ := landManaColorGame(t)
	me := g.Seats[0].ID
	var single, pick []ProducedManaEntry
	g.ReadSnapshot(func() {
		single = g.priceProducedSlotsLocked(me, land, []ProducedManaEntry{{Options: []string{"G"}}}, true)
		pick = g.priceProducedSlotsLocked(me, land, []ProducedManaEntry{{Options: []string{"W", "U"}}}, true)
	})
	if len(single) != 1 || !slices.Equal(single[0].Options, []string{"R"}) {
		t.Errorf("a {G} slot is priced %+v, want one {R} slot", single)
	}
	if len(pick) != 1 || !slices.Equal(pick[0].Options, []string{"R"}) {
		t.Errorf("a {W}/{U} pick is priced %+v, want one {R} slot", pick)
	}
}

// A colour that is not one of the five registers nothing.
func TestLandManaIsColorRefusesAColourThatIsNotOne(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		for _, bad := range []string{"", "C", "purple"} {
			if g.LandsTappedProduceColorThisTurnForEffect(uuid.Nil, g.Seats[0].ID, bad, "Hall") {
				t.Errorf("registered a replacement for colour %q", bad)
			}
		}
	})
}
