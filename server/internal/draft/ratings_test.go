package draft

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draftratings"
)

// fixtureBook is a 17Lands book of the bot fixture set (c2d) that rates
// the named cards from worst to best.
func fixtureBook(t testing.TB, names ...string) *draftratings.Book {
	t.Helper()
	var rows []string
	for i, n := range names {
		rows = append(rows, fmt.Sprintf(`{"n":%q,"w":%.4f,"g":900,"s":4.5,"p":6.25}`, n, 0.45+0.03*float64(i)))
	}
	data := fmt.Sprintf(`{"set":%q,"source":"17lands","format":"PremierDraft","start":"2026-01-01","fetched":"2026-11-11","min_games":500,"cards":[%s]}`,
		botFixtureSet, strings.Join(rows, ","))
	b, err := draftratings.LoadFS(fstest.MapFS{"data/" + botFixtureSet + ".json": {Data: []byte(data)}}, "data")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// Real cards of the fixture set: the first three non-basic, nonland
// cards by name.
func fixtureCardNames(t testing.TB, fx *botFixture, n int) []string {
	t.Helper()
	var out []string
	for _, c := range fx.all {
		if c.Basic || strings.Contains(c.TypeLine, "Land") || strings.HasPrefix(c.Name, botFixtureVanilla) {
			continue
		}
		out = append(out, c.Name)
		if len(out) == n {
			return out
		}
	}
	t.Fatalf("the fixture has fewer than %d real nonland cards", n)
	return nil
}

func TestRatingFromZIsMonotoneAndStaysOnTheScale(t *testing.T) {
	if got := RatingFromZ(0); got != RatingFiller {
		t.Errorf("a card at its set's mean rates %.2f, want RatingFiller %.2f", got, RatingFiller)
	}
	prev := -1.0
	for z := -6.0; z <= 6.0; z += 0.25 {
		v := RatingFromZ(z)
		if v < prev {
			t.Errorf("z %.2f rates %.2f, below the %.2f of the z under it", z, v, prev)
		}
		if v < RatingFloor || v > RatingCeiling {
			t.Errorf("z %.2f rates %.2f, off the scale", z, v)
		}
		prev = v
	}
	if RatingFromZ(-100) != RatingFloor || RatingFromZ(100) != RatingCeiling {
		t.Error("the extremes are not clamped to the floor and the ceiling")
	}
	// A +2.5 sigma card is a bomb by rating.go's own words (above 5).
	if got := RatingFromZ(2.5); got <= 5 {
		t.Errorf("z 2.5 rates %.2f, want above 5", got)
	}
}

func TestBookOverridesReplaceTheComputedRatingOfACoveredCardOnly(t *testing.T) {
	fx := loadBotFixture(t)
	names := fixtureCardNames(t, fx, 4)
	book := fixtureBook(t, names[:3]...) // the fourth has no rating
	with := Rater{Facts: fx.rater.Facts, Overrides: BookOverrides(book)}

	best := fx.card(t, names[2], "a")
	r, ok := book.For(best.Set, best.OracleID, best.Name)
	if !ok {
		t.Fatal("the book does not rate the best card")
	}
	if got, want := with.Rating(best), RatingFromZ(r.Z); got != want {
		t.Errorf("%s rates %.3f with the book, want RatingFromZ(%.2f) = %.3f", best.Name, got, r.Z, want)
	}
	worst, mid := fx.card(t, names[0], "b"), fx.card(t, names[1], "c")
	if !(with.Rating(worst) < with.Rating(mid) && with.Rating(mid) < with.Rating(best)) {
		t.Error("the mapping does not keep the order of the win rates")
	}
	other := fx.card(t, names[3], "d")
	if got, want := with.Rating(other), fx.rater.Rating(other); got != want {
		t.Errorf("a card the book lacks rates %.3f, want its computed %.3f", got, want)
	}
	if got := with.Rating(fx.card(t, "Plains", "e")); got != BasicLandRating {
		t.Errorf("a basic land rates %.2f with the book", got)
	}
	if _, ok := BookOverrides(nil).Rating("x", best); ok {
		t.Error("a nil book overrides a card")
	}
}

// The bots draft by the 17Lands numbers once they are in use, the
// draft stays deterministic, and a book that rates nothing the draft
// holds changes nothing.
func TestBotsDraftBy17LandsRatingsAndStayDeterministic(t *testing.T) {
	fx := loadBotFixture(t)
	t.Cleanup(func() { UseCards(nil); UseRatings(nil) })
	UseCards(fx.idx)
	run := func() [][]Card { return botDraft(t, fx, PickOne, 8, 33, nil) }

	UseRatings(nil)
	plain := run()
	// A card the plain bots leave late is the book's best: the bots
	// should now take it far earlier, in different pools.
	names := fixtureCardNames(t, fx, 30)
	var rows []string
	for i, n := range names {
		w := 0.40
		if i%5 == 0 {
			w = 0.62
		}
		rows = append(rows, fmt.Sprintf(`{"n":%q,"w":%.4f,"g":900,"s":4,"p":5}`, n, w+0.001*float64(i)))
	}
	data := fmt.Sprintf(`{"set":%q,"source":"17lands","format":"PremierDraft","start":"2026-01-01","fetched":"2026-11-11","min_games":500,"cards":[%s]}`,
		botFixtureSet, strings.Join(rows, ","))
	book, err := draftratings.LoadFS(fstest.MapFS{"data/" + botFixtureSet + ".json": {Data: []byte(data)}}, "data")
	if err != nil {
		t.Fatal(err)
	}
	UseRatings(book)
	a, b := run(), run()
	if !reflect.DeepEqual(a, b) {
		t.Error("the same seed drafted two ways with ratings in use")
	}
	if reflect.DeepEqual(a, plain) {
		t.Error("17Lands ratings made no difference to the bots' picks")
	}
	// Every card of a full table is drafted, so the bots' taste shows in
	// when a card is taken: the book's favourites go earlier.
	favoured := map[string]bool{}
	for i, n := range names {
		if i%5 == 0 {
			favoured[n] = true
		}
	}
	meanPick := func(pools [][]Card) float64 {
		sum, n := 0, 0
		for _, p := range pools {
			for i, c := range p {
				if favoured[c.Name] {
					sum += i % 14 // the pick within its round
					n++
				}
			}
		}
		return float64(sum) / float64(n)
	}
	if with, without := meanPick(a), meanPick(plain); with >= without {
		t.Errorf("the book's favourites were taken at pick %.2f with it and %.2f without", with, without)
	}
	UseRatings(nil)
	if !reflect.DeepEqual(run(), plain) {
		t.Error("taking the ratings away did not restore the draft")
	}
}

func TestAViewCarriesEachCardsRatingAndItsSource(t *testing.T) {
	fx := loadBotFixture(t)
	t.Cleanup(func() { UseCards(nil); UseRatings(nil) })
	UseCards(fx.idx)
	names := fixtureCardNames(t, fx, 3)
	UseRatings(fixtureBook(t, names...))

	d, err := New(Config{Name: "d", Set: botFixtureSet, Mode: PickOne, Seats: 2, Seed: 5}, t0)
	if err != nil {
		t.Fatal(err)
	}
	host, _ := d.AddSeat("host", uuid.Nil)
	if err := d.Start(host.ID, fx.src, t0); err != nil {
		t.Fatal(err)
	}
	// The draft's own cards carry no rating: it is filled into a copy.
	for _, c := range d.HeldPack(host.ID) {
		if c.Rating != nil {
			t.Fatalf("%s: the draft's own card carries a rating", c.Name)
		}
	}
	v, ok := d.ViewFor(host.ID, t0, nil)
	if !ok || len(v.Pack) == 0 {
		t.Fatal("no pack in the view")
	}
	seenLands, seenHouse := false, false
	for _, c := range v.Pack {
		switch {
		case c.Basic:
			if c.Rating != nil {
				t.Errorf("basic land %s has a rating", c.Name)
			}
		case c.Rating == nil:
			t.Errorf("%s has no rating", c.Name)
		case c.Rating.Source == SourceLands:
			seenLands = true
			if c.Rating.Tier == "" || c.Rating.Games != 900 || c.Rating.Set != botFixtureSet || c.Rating.Value < 40 || c.Rating.Value > 70 {
				t.Errorf("%s: 17Lands rating %+v", c.Name, c.Rating)
			}
		case c.Rating.Source == SourceHouse:
			seenHouse = true
			if c.Rating.Tier != "" || c.Rating.Games != 0 || c.Rating.Value < 0 || c.Rating.Value > 9 {
				t.Errorf("%s: in-house rating %+v", c.Name, c.Rating)
			}
		default:
			t.Errorf("%s: source %q", c.Name, c.Rating.Source)
		}
	}
	if !seenHouse {
		t.Error("no card of the pack fell back to the in-house rating")
	}
	_ = seenLands // the three book cards may not be in this pack; the picks below do the rest

	// Picks and the pool carry it too, once taken; take one of each source
	// if the pack has it.
	take := []string{}
	for _, c := range v.Pack {
		if c.Rating != nil {
			take = append(take, c.ID)
			break
		}
	}
	if err := d.Pick(host.ID, take, t0); err != nil {
		t.Fatalf("Pick: %v", err)
	}
	v, _ = d.ViewFor(host.ID, t0, nil)
	if len(v.Picks) != 1 || v.Picks[0].Rating == nil {
		t.Errorf("the picks carry no rating: %+v", v.Picks)
	}
	raw, _ := json.Marshal(v.Picks[0])
	if !strings.Contains(string(raw), `"rating":{"source":`) {
		t.Errorf("the wire has no rating object: %s", raw)
	}
}

// Without UseRatings every rating is the in-house one, so a draft with
// no 17Lands file reads as before, plus the rating.
func TestWithNoBookEveryRatingIsTheHouseOne(t *testing.T) {
	fx := loadBotFixture(t)
	UseRatings(nil)
	t.Cleanup(func() { UseCards(nil) })
	UseCards(fx.idx)
	for _, name := range fixtureCardNames(t, fx, 5) {
		c := fx.card(t, name, "x")
		got := fx.rater.View(nil, c)
		if got == nil || got.Source != SourceHouse {
			t.Fatalf("%s: %+v", name, got)
		}
		if want := fx.rater.Rating(c); got.Value < want-0.05 || got.Value > want+0.05 {
			t.Errorf("%s: shown %.1f, rated %.2f", name, got.Value, want)
		}
	}
}
