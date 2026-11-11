package draft

import (
	"math"
	"sync/atomic"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draftratings"
)

// ratings.go is where outside card ratings meet the draft (ADR 0222 §7,
// amendment 2026-11-11, #5703): the number the screen shows beside a
// card, and the number the bots draft by. Both come from 17Lands'
// games-in-hand win rate (package draftratings) where it has the card,
// and from the in-house rating (rating.go) where it does not.

// The 17Lands number as a bot rating. A card's win rate is read as how
// far it sits from its set's mean in standard deviations (Rating.Z); a
// card at the mean rates RatingFiller, and each standard deviation above
// or below moves it RatingPerStdDev points, kept between RatingFloor and
// RatingCeiling. So the mapping is monotone in the win rate, never
// leaves the scale rating.go defines (a bomb is above 5, filler is
// about 2.7), and reads the same in a strong set and a weak one. With a
// set's spread of about 3.5 points of win rate, the best card of a set
// (about +2.5) rates about 5.5 and the worst (about -2.5) about 0.
const (
	// RatingPerStdDev is the points one standard deviation of win rate
	// is worth.
	RatingPerStdDev = 1.1
	// RatingCeiling is the most a card rates from outside data.
	RatingCeiling = 7.0
)

// RatingFromZ is the bot rating of a card whose 17Lands win rate is z
// standard deviations from its set's mean.
func RatingFromZ(z float64) float64 {
	return math.Min(math.Max(RatingFiller+RatingPerStdDev*z, RatingFloor), RatingCeiling)
}

// The sources of a CardRating.
const (
	// SourceLands is 17Lands: Value is the games-in-hand win rate.
	SourceLands = "17lands"
	// SourceHouse is this server's own rating (Rate): Value is points
	// on the bots' scale.
	SourceHouse = "c2"
)

// CardRating is a rating as the client shows it beside a card: the
// `rating` object of a card in a draft view (docs/protocol.md).
type CardRating struct {
	// Source is SourceLands or SourceHouse.
	Source string `json:"source"`
	// Value is, for 17Lands, the games-in-hand win rate in percent, one
	// decimal (58.3), and for the in-house rating its points on the
	// bots' scale, one decimal (3.4). The two are not comparable.
	Value float64 `json:"value"`
	// Tier is a letter grade from the card's set's own spread of win
	// rates ("B+"). 17Lands only.
	Tier string `json:"tier,omitempty"`
	// Games is the games behind a 17Lands win rate, AvgPick the pick
	// number the card is taken at on average and AvgSeen the pick
	// number it is last seen at.
	Games   int     `json:"games,omitempty"`
	AvgPick float64 `json:"avg_pick,omitempty"`
	AvgSeen float64 `json:"avg_seen,omitempty"`
	// Set is the set whose 17Lands numbers these are: the card's own,
	// or for a cube card or a reprint the most recent set that rated
	// it. Format is the 17Lands format and Fetched the "YYYY-MM-DD" the
	// numbers were fetched.
	Set     string `json:"set,omitempty"`
	Format  string `json:"format,omitempty"`
	Fetched string `json:"fetched,omitempty"`
}

// ratingBook is the 17Lands ratings the draft uses, set once at start
// (UseRatings).
var ratingBook atomic.Pointer[draftratings.Book]

// UseRatings tells the draft where outside ratings come from: the bots
// draft by them (through RatingOverrides) and a view carries them. The
// server calls it once at start; nil takes them away. Before it is
// called, and in a test that never calls it, every rating is the
// in-house one.
func UseRatings(b *draftratings.Book) { ratingBook.Store(b) }

// bookOverrides is the RatingOverrides over a Book: the 17Lands win
// rate of a card, as a bot rating (RatingFromZ).
type bookOverrides struct{ book *draftratings.Book }

// Rating answers with the card's 17Lands rating when its set, or the
// most recent set that has it, has one.
func (o bookOverrides) Rating(set string, c Card) (float64, bool) {
	r, ok := o.book.For(set, c.OracleID, c.Name)
	if !ok {
		return 0, false
	}
	return RatingFromZ(r.Z), true
}

// BookOverrides is the per-set ratings seam filled from a Book of
// 17Lands ratings. A nil Book overrides nothing.
func BookOverrides(b *draftratings.Book) RatingOverrides {
	if b == nil {
		return NoOverrides{}
	}
	return bookOverrides{book: b}
}

// View is the rating a screen shows for a card: 17Lands' where the book
// has the card, the rater's own otherwise. A basic land has none: it is
// free in deckbuilding and never worth a pick.
func (r Rater) View(book *draftratings.Book, c Card) *CardRating {
	if c.Basic {
		return nil
	}
	if lr, ok := book.For(c.Set, c.OracleID, c.Name); ok {
		return &CardRating{
			Source: SourceLands, Value: lr.Percent(), Tier: lr.Tier, Games: lr.Games,
			AvgPick: lr.AvgPick, AvgSeen: lr.AvgSeen,
			Set: lr.Set, Format: lr.Format, Fetched: lr.Fetched,
		}
	}
	return &CardRating{Source: SourceHouse, Value: math.Round(Rate(r.FactsOf(c))*10) / 10}
}

// withRatings is cs with each card's rating filled in, as a copy: the draft's
// own cards carry none.
func withRatings(cs []Card) []Card {
	if cs == nil {
		return nil
	}
	book, r := ratingBook.Load(), defaultRater()
	out := make([]Card, len(cs))
	for i, c := range cs {
		c.Rating = r.View(book, c)
		out[i] = c
	}
	return out
}
