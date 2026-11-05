package lobby

// advanced_stats.go is ADR 0217 slice 5 (#4880): the advanced-stats
// read routes.
//
//   - GET /me/stats/advanced?makeup=&range=: the overview, over the
//     caller's counted seats in that makeup and range, plus what comes
//     from their saved deck lists (cards in the most decks, the average
//     mana value) and the colours of the decks they played.
//   - GET /me/stats/decks/{deck_id}?version=&makeup=&range=: one deck,
//     its facts from the saved list (composition, mana curve, average
//     mana value, colour share), its versions (§5) and the stats of the
//     selected one: "current" (since last edit, the default), "all", or
//     a version hash.
//
// Both are for a signed-in person only (a guest or the admin token is
// a 403, signed out a 401), and both refuse with a 403 while the
// caller's advanced-stats setting is off (slice 3's setting). They
// read only the caller's own seats; another seat is never named.
//
// Card roles (§7, slice 4) are not here yet: composition is card types
// only, with roles_available false.

import (
	"context"
	"errors"
	"hash/fnv"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decklibrary"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
)

// errAdvancedStatsOff is the answer while the caller's setting is off.
var errAdvancedStatsOff = httpError(http.StatusForbidden, "advanced stats are off for you; switch them on under Settings → Stats")

// errNoAdvancedStats is the answer on a server with no database.
var errNoAdvancedStats = httpError(http.StatusServiceUnavailable, "advanced stats need the game database, and this server has none")

// advancedCaller is the signed-in caller whose advanced stats are on.
func (c Config) advancedCaller(r *http.Request) (uuid.UUID, error) {
	p, err := signedInUser(r)
	if err != nil {
		return uuid.Nil, err
	}
	prefs, def, err := c.advancedStatsFor(r.Context(), p.UserID)
	if err != nil {
		c.logger().Error("reading a user's stats setting failed", "err", err)
		return uuid.Nil, httpError(http.StatusInternalServerError, "could not load your stats setting; try again")
	}
	if !prefs.Effective(def.Advanced) {
		return uuid.Nil, errAdvancedStatsOff
	}
	if c.AdvancedStats == nil {
		return uuid.Nil, errNoAdvancedStats
	}
	return p.UserID, nil
}

// advancedFilters reads makeup and range; a bad value is a 400.
func advancedFilters(r *http.Request) (gamestats.MakeupFilter, gamestats.Range, error) {
	q := r.URL.Query()
	m, ok := gamestats.ParseMakeup(q.Get("makeup"))
	if !ok {
		return "", "", httpError(http.StatusBadRequest, "makeup must be all, people, with_bots or bots")
	}
	rg, ok := gamestats.ParseRange(q.Get("range"))
	if !ok {
		return "", "", httpError(http.StatusBadRequest, "range must be all, 90d or 30d")
	}
	return m, rg, nil
}

func advancedStatsFailed(c Config, err error) error {
	c.logger().Error("reading a user's advanced statistics failed", "err", err)
	return httpError(http.StatusInternalServerError, "could not load your statistics; try again")
}

// deckRef names a deck: its saved id when it has one.
type deckRef struct {
	DeckID string `json:"deck_id,omitempty"`
	Name   string `json:"name"`
	Games  int    `json:"games"`
}

// advancedDeckRow is one row of the overview's Decks section.
type advancedDeckRow struct {
	DeckID string          `json:"deck_id,omitempty"`
	Name   string          `json:"name"`
	Record gamestats.Tally `json:"record"`
	// Last is the deck's last results, newest first, for the sparkline.
	Last []gamestats.Result `json:"last"`

	placeSum, placed int
}

// deckRowsMax is the results a deck row's sparkline carries.
const deckRowsMax = 20

// colourGames is one colour's share of the caller's games: the games
// whose deck's colour identity (its commanders') has the colour, out of
// the games whose commanders resolved. A five-colour deck counts for
// all five, so the shares do not add up to 1.
type colourGames struct {
	Colour string         `json:"colour"`
	Games  gamestats.Rate `json:"games"`
}

// libraryCard is a card in several of the caller's saved decks.
type libraryCard struct {
	Name       string `json:"name"`
	ScryfallID string `json:"scryfall_id,omitempty"`
	Decks      int    `json:"decks"`
}

// cardsInMostDecksMax caps the overview's cards in the most decks.
const cardsInMostDecksMax = 20

// advancedLibrary is what the overview reads from saved deck lists.
type advancedLibrary struct {
	// CardIndex is false on a server with no card index: nothing here
	// can be computed then.
	CardIndex bool `json:"card_index"`
	// Decks is the caller's saved decks.
	Decks int `json:"decks"`
	// CardsInMostDecks counts saved decks per card (basic lands left
	// out), most decks first; a card in one deck only is not listed.
	CardsInMostDecks []libraryCard `json:"cards_in_most_decks"`
	// AvgManaValue is the mean of each saved deck's average mana value,
	// weighted by the games it played in the filter; AvgManaValueGames
	// is those games.
	AvgManaValue      *float64 `json:"avg_mana_value"`
	AvgManaValueGames int      `json:"avg_mana_value_games"`
	// RolesAvailable is false until card roles ship (ADR 0217 slice 4).
	RolesAvailable bool `json:"roles_available"`
}

// advancedOverviewResponse is GET /me/stats/advanced's body.
type advancedOverviewResponse struct {
	Makeup gamestats.MakeupFilter `json:"makeup"`
	Range  gamestats.Range        `json:"range"`
	gamestats.Summary
	MostPlayedDeck *deckRef          `json:"most_played_deck"`
	Decks          []advancedDeckRow `json:"decks"`
	Colours        []colourGames     `json:"colours"`
	Library        advancedLibrary   `json:"library"`
}

// myAdvancedStats is GET /me/stats/advanced.
func myAdvancedStats(c Config, w http.ResponseWriter, r *http.Request) error {
	user, err := c.advancedCaller(r)
	if err != nil {
		return err
	}
	makeup, rng, err := advancedFilters(r)
	if err != nil {
		return err
	}
	since := rng.Since(c.now())
	ctx, cancel := context.WithTimeout(r.Context(), statsTimeout)
	defer cancel()
	seats, err := c.AdvancedStats.AdvancedSeats(ctx, user.String(), func(s *gamestats.SeatRow) bool {
		return s.EndedAt >= since && makeup.Matches(s)
	})
	if err != nil {
		return advancedStatsFailed(c, err)
	}
	decks, err := c.libraryStore().List(ctx, user)
	if err != nil {
		return advancedStatsFailed(c, err)
	}
	out := advancedOverviewResponse{Makeup: makeup, Range: rng, Summary: gamestats.Summarise(seats),
		Decks: advancedDeckRows(seats)}
	out.Commander.ByCommander = gamestats.ByCommander(seats)
	if len(out.Decks) > 0 {
		d := out.Decks[0]
		out.MostPlayedDeck = &deckRef{DeckID: d.DeckID, Name: d.Name, Games: d.Record.Games}
	}
	out.Colours = gameColours(c.Cards, seats)
	out.Library = c.advancedLibraryOf(decks, seats)
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, out)
}

// advancedDeckRows groups seats by deck as GET /me/stats does: the
// saved deck id, or the name for a seat with none. Most played first.
func advancedDeckRows(seats []gamestats.SeatRow) []advancedDeckRow {
	m := map[string]*advancedDeckRow{}
	var order []string
	for i := range seats {
		s := &seats[i]
		if s.DeckID == "" && s.DeckName == "" {
			continue
		}
		key := s.DeckID
		if key == "" {
			key = "name:" + s.DeckName
		}
		row := m[key]
		if row == nil {
			name := s.DeckName // seats run newest first: the latest name
			if name == "" {
				name = "Unnamed deck"
			}
			row = &advancedDeckRow{DeckID: s.DeckID, Name: name, Last: []gamestats.Result{}}
			m[key] = row
			order = append(order, key)
		}
		row.Record.Games++
		switch s.Result {
		case gamestats.ResultWin:
			row.Record.Wins++
		case gamestats.ResultLoss:
			row.Record.Losses++
		case gamestats.ResultDraw:
			row.Record.Draws++
		}
		if s.Place != nil {
			row.placeSum += *s.Place
			row.placed++
		}
		if len(row.Last) < deckRowsMax {
			row.Last = append(row.Last, s.Result)
		}
	}
	out := make([]advancedDeckRow, 0, len(order))
	for _, k := range order {
		row := *m[k]
		if row.placed > 0 {
			v := float64(row.placeSum) / float64(row.placed)
			row.Record.AvgPlace = &v
		}
		out = append(out, row)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Record.Games > out[j].Record.Games })
	return out
}

// gameColours is the overview's colour share: per colour, the games
// whose commanders' colour identity has it.
func gameColours(idx *cards.Index, seats []gamestats.SeatRow) []colourGames {
	counts := map[string]int{}
	resolved := 0
	for i := range seats {
		s := &seats[i]
		if idx == nil || len(s.Commanders) == 0 {
			continue
		}
		identity, ok := map[string]bool{}, true
		for _, name := range s.Commanders {
			card, found := idx.FindByName(name)
			if !found {
				ok = false
				break
			}
			for _, l := range card.ColorIdentity {
				identity[strings.ToUpper(l)] = true
			}
		}
		if !ok {
			continue
		}
		resolved++
		for l := range identity {
			counts[l]++
		}
	}
	out := make([]colourGames, 0, 5)
	for _, l := range []string{"W", "U", "B", "R", "G"} {
		out = append(out, colourGames{Colour: l, Games: gamestats.RateOf(counts[l], resolved)})
	}
	return out
}

// advancedLibraryOf reads the caller's saved deck lists.
func (c Config) advancedLibraryOf(decks []decklibrary.Deck, seats []gamestats.SeatRow) advancedLibrary {
	out := advancedLibrary{CardIndex: c.Cards != nil, Decks: len(decks), CardsInMostDecks: []libraryCard{}}
	if c.Cards == nil {
		return out
	}
	games := map[string]int{}
	for i := range seats {
		if seats[i].DeckID != "" {
			games[seats[i].DeckID]++
		}
	}
	byOracle := map[uuid.UUID]*libraryCard{}
	var mvSum float64
	for _, d := range decks {
		f, ok := deckFactsFor(c.Cards, d)
		if !ok {
			continue
		}
		if n := games[d.ID.String()]; n > 0 && f.facts.AvgManaValue != nil {
			mvSum += *f.facts.AvgManaValue * float64(n)
			out.AvgManaValueGames += n
		}
		for _, lc := range f.cards {
			a := byOracle[lc.oracle]
			if a == nil {
				a = &libraryCard{Name: lc.name, ScryfallID: lc.scryfallID}
				byOracle[lc.oracle] = a
			}
			a.Decks++
		}
	}
	if out.AvgManaValueGames > 0 {
		v := mvSum / float64(out.AvgManaValueGames)
		out.AvgManaValue = &v
	}
	for _, a := range byOracle {
		if a.Decks >= 2 {
			out.CardsInMostDecks = append(out.CardsInMostDecks, *a)
		}
	}
	sort.Slice(out.CardsInMostDecks, func(i, j int) bool {
		a, b := out.CardsInMostDecks[i], out.CardsInMostDecks[j]
		if a.Decks != b.Decks {
			return a.Decks > b.Decks
		}
		return a.Name < b.Name
	})
	if len(out.CardsInMostDecks) > cardsInMostDecksMax {
		out.CardsInMostDecks = out.CardsInMostDecks[:cardsInMostDecksMax]
	}
	return out
}

// libraryOracleCard is one distinct non-basic card of a saved deck.
type libraryOracleCard struct {
	oracle     uuid.UUID
	name       string
	scryfallID string
}

// savedDeckFacts is a saved deck's facts and its distinct non-basic
// cards, for cards in the most decks.
type savedDeckFacts struct {
	facts gamestats.DeckFacts
	cards []libraryOracleCard
}

// deckFactsKey is the cache key ADR 0217 §5 names, (deck_id,
// updated_at), plus a hash of the stored list, so two saves in one
// millisecond cannot share an entry, and the card index's load time,
// since a refreshed index can resolve a name it did not before.
type deckFactsKey struct {
	id      uuid.UUID
	updated int64
	source  uint64
	index   int64
}

// deckFactsCacheMax bounds the cache; past it, the cache starts over.
const deckFactsCacheMax = 4096

var deckFactsCache = struct {
	sync.Mutex
	m map[deckFactsKey]savedDeckFacts
}{m: map[deckFactsKey]savedDeckFacts{}}

// deckFactsFor parses and resolves a saved deck (the parse a seat
// takes) and computes its facts, cached in memory. ok is false when the
// stored list no longer parses or there is no card index. A name the
// index does not know is counted in Unresolved, not refused.
func deckFactsFor(idx *cards.Index, d decklibrary.Deck) (savedDeckFacts, bool) {
	if idx == nil {
		return savedDeckFacts{}, false
	}
	source := fnv.New64a()
	_, _ = source.Write([]byte(d.SourceText))
	key := deckFactsKey{id: d.ID, updated: d.UpdatedAt.UnixMilli(), source: source.Sum64(), index: idx.LoadedAt().UnixMilli()}
	deckFactsCache.Lock()
	f, hit := deckFactsCache.m[key]
	deckFactsCache.Unlock()
	if hit {
		return f, true
	}
	entries, err := libraryDeckEntries(d)
	if err != nil {
		return savedDeckFacts{}, false
	}
	list, unresolved := resolveForStats(idx, entries)
	f = savedDeckFacts{facts: gamestats.FactsOf(list, unresolved)}
	seen := map[uuid.UUID]bool{}
	for _, dc := range list {
		c := dc.Card
		if c.OracleID == uuid.Nil || seen[c.OracleID] || gamestats.IsBasicLand(c) {
			continue
		}
		seen[c.OracleID] = true
		f.cards = append(f.cards, libraryOracleCard{oracle: c.OracleID, name: c.Name, scryfallID: c.ID.String()})
	}
	deckFactsCache.Lock()
	if len(deckFactsCache.m) >= deckFactsCacheMax {
		deckFactsCache.m = map[deckFactsKey]savedDeckFacts{}
	}
	deckFactsCache.m[key] = f
	deckFactsCache.Unlock()
	return f, true
}

// resolveForStats resolves a deck list's main deck and commanders
// against the index, as deck.Resolve does, but leniently: an unknown
// name is counted, not an error. The sideboard, a companion,
// Attractions and sticker sheets are not deck cards (ADR 0217 §5).
func resolveForStats(idx *cards.Index, entries []deck.Entry) ([]gamestats.DeckCard, int) {
	var out []gamestats.DeckCard
	unresolved := 0
	for _, e := range entries {
		if e.Count <= 0 || e.IsSideboard || e.IsCompanion {
			continue
		}
		c, ok := idx.FindByName(e.Name)
		if !ok {
			unresolved += e.Count
			continue
		}
		if deck.PrintsAttraction(c) || deck.PrintsStickerSheet(c) {
			continue
		}
		if p, found := idx.FindPrintingOf(c, e.PrintingID, e.SetCode, e.CollectorNo); found {
			c = p
		}
		out = append(out, gamestats.DeckCard{Card: c, Count: e.Count, Commander: e.IsCommander})
	}
	return out, unresolved
}

// deckVersion is one version of a deck the caller played (ADR 0217 §5).
type deckVersion struct {
	Hash string `json:"hash"`
	// Current is the saved list's version.
	Current bool `json:"current"`
	// First and Last are the first and last game's end, Unix ms.
	First int64 `json:"first"`
	Last  int64 `json:"last"`
	Games int   `json:"games"`
}

// advancedDeckInfo is the deck a deck page is about.
type advancedDeckInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	UpdatedAt  int64    `json:"updated_at"` // Unix ms
	Commanders []string `json:"commanders"`
	// Facts is nil when the stored list no longer parses or the server
	// has no card index.
	Facts *gamestats.DeckFacts `json:"facts"`
}

// typeAverage is a card type's mean count over the caller's other
// saved decks, for composition's comparison tick.
type typeAverage struct {
	Type    string  `json:"type"`
	Average float64 `json:"average"`
}

// otherDecksAverage is composition over the caller's other decks.
type otherDecksAverage struct {
	Decks int           `json:"decks"`
	Types []typeAverage `json:"types"`
}

// Deck page versions: "current" (since last edit) and "all".
const (
	deckVersionCurrent = "current"
	deckVersionAll     = "all"
)

var deckHashRe = regexp.MustCompile(`^[0-9a-f]{16}$`)

// advancedDeckResponse is GET /me/stats/decks/{deck_id}'s body.
type advancedDeckResponse struct {
	Deck advancedDeckInfo `json:"deck"`
	// Version is the view shown: "current", "all" or a version hash.
	Version string `json:"version"`
	// NoGamesSinceEdit is true when "current" has no games at all, in
	// any makeup or range: the page says so and offers All time.
	NoGamesSinceEdit bool                   `json:"no_games_since_edit"`
	Versions         []deckVersion          `json:"versions"`
	Makeup           gamestats.MakeupFilter `json:"makeup"`
	Range            gamestats.Range        `json:"range"`
	gamestats.Summary
	DeadCards  []gamestats.DeadCard `json:"dead_cards"`
	OtherDecks *otherDecksAverage   `json:"other_decks"`
}

// myAdvancedDeckStats is GET /me/stats/decks/{deck_id}.
func myAdvancedDeckStats(c Config, w http.ResponseWriter, r *http.Request) error {
	user, err := c.advancedCaller(r)
	if err != nil {
		return err
	}
	makeup, rng, err := advancedFilters(r)
	if err != nil {
		return err
	}
	version := r.URL.Query().Get("version")
	switch {
	case version == "":
		version = deckVersionCurrent
	case version == deckVersionCurrent, version == deckVersionAll, deckHashRe.MatchString(version):
	default:
		return httpError(http.StatusBadRequest, "version must be current, all or a version hash")
	}
	ctx, cancel := context.WithTimeout(r.Context(), statsTimeout)
	defer cancel()

	// Only a saved deck of the caller's: anyone else's, or no deck, is
	// a 404, so an id reveals nothing. A deleted deck's seats lose their
	// deck id (migration 0008), so its games group by name on the
	// overview and have no deck page.
	saved, err := ownedLibraryDeckByID(ctx, c, user, r.PathValue("deck_id"))
	if err != nil {
		var se *httpStatusError
		if errors.As(err, &se) {
			return err
		}
		return advancedStatsFailed(c, err)
	}
	seats, err := c.AdvancedStats.AdvancedSeats(ctx, user.String(), func(s *gamestats.SeatRow) bool {
		return s.DeckID == saved.ID.String()
	})
	if err != nil {
		return advancedStatsFailed(c, err)
	}

	out := advancedDeckResponse{Makeup: makeup, Range: rng, Versions: []deckVersion{}, DeadCards: []gamestats.DeadCard{}}
	out.Deck = advancedDeckInfo{ID: saved.ID.String(), Name: saved.Name, UpdatedAt: saved.UpdatedAt.UnixMilli(),
		Commanders: nonNil(saved.Commanders)}
	current := ""
	if f, ok := deckFactsFor(c.Cards, saved); ok {
		facts := f.facts
		out.Deck.Facts = &facts
		current = facts.Hash
	}
	decks, err := c.libraryStore().List(ctx, user)
	if err != nil {
		return advancedStatsFailed(c, err)
	}
	out.OtherDecks = otherDecksAverageOf(c.Cards, decks, saved.ID)
	out.Versions = deckVersionsOf(seats, current)

	if version == deckVersionCurrent && current == "" {
		// A list with no resolvable card, or a server with no card
		// index, has no current version: all time is all there is.
		version = deckVersionAll
	}
	out.Version = version
	since := rng.Since(c.now())
	kept := make([]gamestats.SeatRow, 0, len(seats))
	for i := range seats {
		s := &seats[i]
		if s.EndedAt < since || !makeup.Matches(s) {
			continue
		}
		switch version {
		case deckVersionAll:
		case deckVersionCurrent:
			if s.DeckHash != current {
				continue
			}
		default:
			if s.DeckHash != version {
				continue
			}
		}
		kept = append(kept, *s)
	}
	// No game since the last edit at all, whatever the filters keep.
	out.NoGamesSinceEdit = version == deckVersionCurrent
	for i := range seats {
		if seats[i].DeckHash == current {
			out.NoGamesSinceEdit = false
			break
		}
	}
	out.Summary = gamestats.Summarise(kept)
	var isLand func(string) bool
	if c.Cards != nil {
		isLand = func(name string) bool {
			card, ok := c.Cards.FindByName(name)
			return ok && gamestats.IsLandCard(card)
		}
	}
	out.DeadCards = gamestats.DeadCards(kept, isLand)
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, out)
}

// deckVersionsOf lists each version the seats played, last played
// first, with the saved list's own version marked. Seats with no hash
// (a version-1 record) are in no version.
func deckVersionsOf(seats []gamestats.SeatRow, current string) []deckVersion {
	m := map[string]*deckVersion{}
	for i := range seats {
		s := &seats[i]
		if s.DeckHash == "" {
			continue
		}
		v := m[s.DeckHash]
		if v == nil {
			v = &deckVersion{Hash: s.DeckHash, Current: s.DeckHash == current, First: s.EndedAt, Last: s.EndedAt}
			m[s.DeckHash] = v
		}
		v.Games++
		v.First, v.Last = min(v.First, s.EndedAt), max(v.Last, s.EndedAt)
	}
	out := make([]deckVersion, 0, len(m))
	for _, v := range m {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Last != out[j].Last {
			return out[i].Last > out[j].Last
		}
		return out[i].Hash < out[j].Hash
	})
	return out
}

// otherDecksAverageOf is the mean count of each card type over the
// caller's saved decks other than except, nil when there are none.
func otherDecksAverageOf(idx *cards.Index, decks []decklibrary.Deck, except uuid.UUID) *otherDecksAverage {
	sums := make([]int, len(gamestats.CardTypes))
	n := 0
	for _, d := range decks {
		if d.ID == except {
			continue
		}
		f, ok := deckFactsFor(idx, d)
		if !ok {
			continue
		}
		n++
		for i, t := range f.facts.Types {
			sums[i] += t.Cards
		}
	}
	if n == 0 {
		return nil
	}
	out := &otherDecksAverage{Decks: n, Types: make([]typeAverage, len(gamestats.CardTypes))}
	for i, t := range gamestats.CardTypes {
		out.Types[i] = typeAverage{Type: t, Average: float64(sums[i]) / float64(n)}
	}
	return out
}
