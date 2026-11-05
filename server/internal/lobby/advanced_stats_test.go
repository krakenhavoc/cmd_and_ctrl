package lobby

// Tests for advanced_stats.go (ADR 0217 slice 5): who may read the
// advanced-stats routes, the makeup and range filters, the deck route's
// ownership rule and its version selection.

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decklibrary"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
)

// advancedNow is the stack's clock, so the range filter is fixed.
var advancedNow = time.UnixMilli(1_800_000_000_000)

type advancedStack struct {
	statsPrefsStack
	stats   *gamestats.SQLStore
	library decklibrary.Store
	oracle  map[string]uuid.UUID
}

// advancedCards is the test index: a commander, a basic land and three
// spells, by name.
func advancedCards() (*cards.Index, map[string]uuid.UUID) {
	idx := cards.NewIndex()
	oracle := map[string]uuid.UUID{}
	put := func(name, typeLine, cost string, identity ...string) {
		c := cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost,
			ColorIdentity: identity, Layout: "normal", Lang: "en", SetType: "core",
			Legalities: map[string]string{"commander": "legal"}}
		idx.Put(c)
		oracle[name] = c.OracleID
	}
	put("Atraxa, Praetors' Voice", "Legendary Creature — Phyrexian Angel Horror", "{G}{W}{U}{B}", "W", "U", "B", "G")
	put("Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", "{2}{R}{R}", "R")
	put("Forest", "Basic Land — Forest", "")
	put("Sol Ring", "Artifact", "{1}")
	put("Lightning Bolt", "Instant", "{R}", "R")
	put("Counterspell", "Instant", "{U}{U}", "U")
	return idx, oracle
}

func newAdvancedStack(t *testing.T) advancedStack {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	d := openTestDB(t, t.TempDir())
	idx, oracle := advancedCards()
	stats, library := gamestats.NewSQLStore(d), decklibrary.NewSQLStore(d)
	s := newAdminStackWith(t, idx, func(c *Config) {
		c.StatsPrefs = gamestats.NewSQLPrefsStore(d)
		c.AdvancedStats = stats
		c.DeckLibrary = library
		c.Now = func() time.Time { return advancedNow }
	}, statsAdminDiscordID)
	return advancedStack{statsPrefsStack: statsPrefsStack{adminStack: s, db: d}, stats: stats, library: library, oracle: oracle}
}

// on switches the person's advanced stats on.
func (s advancedStack) on(t *testing.T, tok string) {
	t.Helper()
	if code := status(t, s.srv, "PUT", "/me/stats/prefs", tok, map[string]bool{"advanced": true}); code != http.StatusOK {
		t.Fatalf("switching advanced stats on: %d", code)
	}
}

const atraxaList = "Commander:\n1 Atraxa, Praetors' Voice\n\nDeck:\n1 Sol Ring\n10 Forest\n1 Counterspell\n1 Not A Real Card\n"

func (s advancedStack) saveDeck(t *testing.T, owner uuid.UUID, name, text string, commanders ...string) decklibrary.Deck {
	t.Helper()
	d, err := s.library.Upsert(context.Background(), owner, name, "text", text, commanders, 14)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// seededGame is one seeded game: the caller's seat 0 and the other seats'
// kinds ("bot" or "person").
type seededGame struct {
	daysAgo int
	deckID  string
	hash    string
	result  gamestats.Result
	others  []string
}

// seed writes a finished game with a version-2 record for user.
func (s advancedStack) seed(t *testing.T, user uuid.UUID, g seededGame) {
	t.Helper()
	id := uuid.New()
	ended := advancedNow.Add(-time.Duration(g.daysAgo) * 24 * time.Hour).UnixMilli()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := s.db.Exec(q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	exec(`INSERT INTO games (id, name, state, created_at, started_at, ended_at) VALUES (?, 'Friday', 'ended', 1, ?, ?)`,
		id.String(), ended-40*60000, ended)
	var deckID any
	if g.deckID != "" {
		deckID = g.deckID
	}
	exec(`INSERT INTO seats (game_id, seat, player_id, user_id, deck_id, deck_name) VALUES (?, 0, 'p0', ?, ?, 'Atraxa')`,
		id.String(), user.String(), deckID)
	for i, kind := range g.others {
		if kind == "bot" {
			exec(`INSERT INTO seats (game_id, seat, player_id, bot_tier) VALUES (?, ?, ?, 'heuristic')`, id.String(), i+1, uuid.NewString())
		} else {
			exec(`INSERT INTO seats (game_id, seat, player_id, guest_name) VALUES (?, ?, ?, 'Guest')`, id.String(), i+1, uuid.NewString())
		}
	}
	place := func(p int) *int { return &p }
	rec := gamestats.Game{GameID: id, Version: gamestats.Version, Source: gamestats.SourceLive, LogComplete: true,
		Turns: 10, Rounds: 3, StartingSeat: place(0), ComputedAt: time.UnixMilli(ended), FramesComplete: true}
	mine := gamestats.Seat{Seat: 0, Result: g.result, Place: place(2), Commanders: []string{"Atraxa, Praetors' Voice"},
		TurnPosition: place(1), DeckHash: g.hash,
		Cards: []gamestats.Card{{Name: "Sol Ring", Casts: 1, FirstCastTurn: 1}},
		Play: &gamestats.Play{OwnTurns: 1, LandDropsHit: 1, Turns: []gamestats.TurnStat{{OwnTurn: 1, Turn: 1, LandPlayed: 1, LandDrops: 1}},
			Seen: []gamestats.SeenCard{{Name: "Counterspell", Copies: 1}, {Name: "Forest", Copies: 2}}}}
	if g.result == gamestats.ResultWin {
		mine.Place = place(1)
		rec.WinKind = gamestats.KindCombat
	}
	rec.Seats = append(rec.Seats, mine)
	for i := range g.others {
		rec.Seats = append(rec.Seats, gamestats.Seat{Seat: i + 1, Result: gamestats.ResultLoss, Place: place(i + 2), Commanders: []string{}})
	}
	if err := s.stats.Put(context.Background(), rec); err != nil {
		t.Fatal(err)
	}
}

func (s advancedStack) currentHash() string {
	return gamestats.DeckHash(map[string]int{
		s.oracle["Atraxa, Praetors' Voice"].String(): 1, s.oracle["Sol Ring"].String(): 1,
		s.oracle["Forest"].String(): 10, s.oracle["Counterspell"].String(): 1,
	})
}

func TestAdvancedStatsRoutesRefuseWithoutTheSetting(t *testing.T) {
	s := newAdvancedStack(t)
	tok, user := s.person(t, "222", "Bea")
	d := s.saveDeck(t, user, "Atraxa", atraxaList, "Atraxa, Praetors' Voice")
	guest := s.issue(t, auth.Principal{Role: auth.RolePlayer, GameID: uuid.New(), PlayerID: uuid.New()})
	for _, path := range []string{"/me/stats/advanced", "/me/stats/decks/" + d.ID.String()} {
		if code := status(t, s.srv, "GET", path, "", nil); code != http.StatusUnauthorized {
			t.Errorf("signed out %s: %d, want 401", path, code)
		}
		for name, who := range map[string]string{"guest": guest, "admin token": s.adminToken(t)} {
			if code := status(t, s.srv, "GET", path, who, nil); code != http.StatusForbidden {
				t.Errorf("%s %s: %d, want 403", name, path, code)
			}
		}
		// Off by default: refused.
		if code := status(t, s.srv, "GET", path, tok, nil); code != http.StatusForbidden {
			t.Errorf("setting off %s: %d, want 403", path, code)
		}
	}
	s.on(t, tok)
	for _, path := range []string{"/me/stats/advanced", "/me/stats/decks/" + d.ID.String()} {
		if code := status(t, s.srv, "GET", path, tok, nil); code != http.StatusOK {
			t.Errorf("setting on %s: %d, want 200", path, code)
		}
	}
	// Switched off again: refused again.
	if code := status(t, s.srv, "PUT", "/me/stats/prefs", tok, map[string]bool{"advanced": false}); code != http.StatusOK {
		t.Fatal(code)
	}
	if code := status(t, s.srv, "GET", "/me/stats/advanced", tok, nil); code != http.StatusForbidden {
		t.Errorf("switched off: %d, want 403", code)
	}
}

// The admin's default turns it on for someone who never touched it.
func TestAdvancedStatsFollowTheDefault(t *testing.T) {
	s := newAdvancedStack(t)
	tok, _ := s.person(t, "222", "Bea")
	admin, adminID := s.person(t, statsAdminDiscordID, "Ian")
	s.adminModeOn(t, adminID)
	if code := status(t, s.srv, "PUT", "/admin/stats/advanced-default", admin, map[string]bool{"on": true}); code != http.StatusOK {
		t.Fatalf("default on: %d", code)
	}
	if code := status(t, s.srv, "GET", "/me/stats/advanced", tok, nil); code != http.StatusOK {
		t.Errorf("following a default of on: %d, want 200", code)
	}
}

func TestAdvancedOverviewFilters(t *testing.T) {
	s := newAdvancedStack(t)
	tok, user := s.person(t, "222", "Bea")
	other, otherUser := s.person(t, "333", "Cal")
	s.on(t, tok)
	s.on(t, other)
	atraxa := s.saveDeck(t, user, "Atraxa", atraxaList, "Atraxa, Praetors' Voice")
	s.saveDeck(t, user, "Krenko", "Commander:\n1 Krenko, Mob Boss\n\nDeck:\n1 Sol Ring\n1 Lightning Bolt\n10 Forest\n1 Counterspell\n", "Krenko, Mob Boss")
	h := s.currentHash()
	people := []string{"person", "person", "person"}
	for _, g := range []seededGame{
		{daysAgo: 1, deckID: atraxa.ID.String(), hash: h, result: gamestats.ResultWin, others: people},
		{daysAgo: 2, deckID: atraxa.ID.String(), hash: h, result: gamestats.ResultLoss, others: people},
		{daysAgo: 3, deckID: atraxa.ID.String(), hash: h, result: gamestats.ResultWin, others: people},
		{daysAgo: 4, deckID: atraxa.ID.String(), hash: h, result: gamestats.ResultLoss, others: []string{"bot", "person", "person"}},
		{daysAgo: 5, deckID: atraxa.ID.String(), hash: h, result: gamestats.ResultLoss, others: []string{"bot", "bot", "bot"}},
		{daysAgo: 100, deckID: "", hash: "", result: gamestats.ResultWin, others: people},
	} {
		s.seed(t, user, g)
	}
	s.seed(t, otherUser, seededGame{daysAgo: 1, result: gamestats.ResultWin, others: people})

	get := func(query string) advancedOverviewResponse {
		t.Helper()
		code, out := decodeInto[advancedOverviewResponse](t, s.statsPrefsStack, "GET", "/me/stats/advanced"+query, tok, nil)
		if code != http.StatusOK {
			t.Fatalf("GET %s: %d", query, code)
		}
		return out
	}
	all := get("")
	if all.Makeup != gamestats.MakeupAll || all.Range != gamestats.RangeAll || all.Record.Games != 6 || all.Record.Wins != 3 {
		t.Errorf("all: %s %s %+v", all.Makeup, all.Range, all.Record)
	}
	if all.WinRate.Rate == nil || *all.WinRate.Rate != 0.5 || all.Coverage.Frames != 6 {
		t.Errorf("win rate %+v coverage %+v", all.WinRate, all.Coverage)
	}
	for q, want := range map[string]int{
		"?makeup=people": 4, "?makeup=with_bots": 1, "?makeup=bots": 1, "?range=30d": 5, "?range=90d": 5,
		"?makeup=people&range=30d": 3, "?makeup=all&range=all": 6,
	} {
		if got := get(q).Record.Games; got != want {
			t.Errorf("%s: %d games, want %d", q, got, want)
		}
	}
	for _, bad := range []string{"?makeup=humans", "?range=7d"} {
		if code := status(t, s.srv, "GET", "/me/stats/advanced"+bad, tok, nil); code != http.StatusBadRequest {
			t.Errorf("%s: %d, want 400", bad, code)
		}
	}

	if all.MostPlayedDeck == nil || all.MostPlayedDeck.DeckID != atraxa.ID.String() || all.MostPlayedDeck.Games != 5 {
		t.Errorf("most played %+v", all.MostPlayedDeck)
	}
	if len(all.Decks) != 2 || all.Decks[1].DeckID != "" || all.Decks[1].Name != "Atraxa" || len(all.Decks[0].Last) != 5 {
		t.Errorf("deck rows %+v", all.Decks)
	}
	// Colours: every game's commander is Atraxa (WUBG), none red.
	colours := map[string]gamestats.Rate{}
	for _, c := range all.Colours {
		colours[c.Colour] = c.Games
	}
	if colours["W"].N != 6 || colours["R"].N != 0 || colours["R"].Of != 6 {
		t.Errorf("colours %+v", all.Colours)
	}
	lib := all.Library
	if !lib.CardIndex || lib.Decks != 2 || lib.RolesAvailable {
		t.Errorf("library %+v", lib)
	}
	inBoth := map[string]int{}
	for _, c := range lib.CardsInMostDecks {
		inBoth[c.Name] = c.Decks
	}
	if inBoth["Sol Ring"] != 2 || inBoth["Counterspell"] != 2 || len(inBoth) != 2 {
		t.Errorf("cards in the most decks %+v: Sol Ring and Counterspell, Forest left out", lib.CardsInMostDecks)
	}
	// Atraxa's non-land cards: Atraxa (4), Sol Ring (1), Counterspell (2):
	// 7/3, over its 5 games; the Krenko deck played none.
	if lib.AvgManaValue == nil || lib.AvgManaValueGames != 5 || *lib.AvgManaValue < 2.33 || *lib.AvgManaValue > 2.34 {
		t.Errorf("avg mana value %v over %d", lib.AvgManaValue, lib.AvgManaValueGames)
	}
	if len(all.MostCast) != 1 || all.MostCast[0].Games != 6 {
		t.Errorf("most cast %+v", all.MostCast)
	}
	if len(all.Commander.ByCommander) != 1 || all.Commander.ByCommander[0].Tally.Games != 6 {
		t.Errorf("by commander %+v", all.Commander.ByCommander)
	}

	// The other person sees only their own game.
	code, theirs := decodeInto[advancedOverviewResponse](t, s.statsPrefsStack, "GET", "/me/stats/advanced", other, nil)
	if code != http.StatusOK || theirs.Record.Games != 1 || theirs.Library.Decks != 0 {
		t.Errorf("another person: %d %+v", code, theirs.Record)
	}
}

func TestAdvancedDeckVersionsAndOwnership(t *testing.T) {
	s := newAdvancedStack(t)
	tok, user := s.person(t, "222", "Bea")
	other, otherUser := s.person(t, "333", "Cal")
	s.on(t, tok)
	s.on(t, other)
	d := s.saveDeck(t, user, "Atraxa", atraxaList, "Atraxa, Praetors' Voice")
	theirs := s.saveDeck(t, otherUser, "Theirs", atraxaList, "Atraxa, Praetors' Voice")
	cur, old := s.currentHash(), "0123456789abcdef"
	people := []string{"person", "person", "person"}
	for _, g := range []seededGame{
		{daysAgo: 1, deckID: d.ID.String(), hash: cur, result: gamestats.ResultWin, others: people},
		{daysAgo: 2, deckID: d.ID.String(), hash: cur, result: gamestats.ResultLoss, others: []string{"bot", "bot", "bot"}},
		{daysAgo: 10, deckID: d.ID.String(), hash: old, result: gamestats.ResultLoss, others: people},
		{daysAgo: 11, deckID: d.ID.String(), hash: old, result: gamestats.ResultLoss, others: people},
		{daysAgo: 12, deckID: d.ID.String(), hash: old, result: gamestats.ResultWin, others: people},
		// A version-1 record has no hash: in all time only.
		{daysAgo: 13, deckID: d.ID.String(), hash: "", result: gamestats.ResultWin, others: people},
	} {
		s.seed(t, user, g)
	}
	s.seed(t, otherUser, seededGame{daysAgo: 1, deckID: theirs.ID.String(), hash: cur, result: gamestats.ResultWin, others: people})

	path := "/me/stats/decks/" + d.ID.String()
	get := func(query string) advancedDeckResponse {
		t.Helper()
		code, out := decodeInto[advancedDeckResponse](t, s.statsPrefsStack, "GET", path+query, tok, nil)
		if code != http.StatusOK {
			t.Fatalf("GET %s: %d", query, code)
		}
		return out
	}
	def := get("")
	if def.Version != "current" || def.Record.Games != 2 || def.NoGamesSinceEdit {
		t.Errorf("since last edit: %s %+v %v", def.Version, def.Record, def.NoGamesSinceEdit)
	}
	f := def.Deck.Facts
	if f == nil || f.Hash != cur || f.Unresolved != 1 || f.Cards != 13 || f.ManaCurve[1] != 1 || f.ManaCurve[2] != 1 || f.ManaCurve[4] != 1 {
		t.Fatalf("facts %+v", f)
	}
	if def.Deck.Name != "Atraxa" || def.Deck.UpdatedAt == 0 || len(def.Deck.Commanders) != 1 {
		t.Errorf("deck %+v", def.Deck)
	}
	if len(def.Versions) != 2 || def.Versions[0].Hash != cur || !def.Versions[0].Current || def.Versions[0].Games != 2 ||
		def.Versions[1].Hash != old || def.Versions[1].Games != 3 || def.Versions[1].Current {
		t.Errorf("versions %+v", def.Versions)
	}
	if v := def.Versions[1]; v.First >= v.Last {
		t.Errorf("version range %+v", v)
	}
	// The dead cards: Counterspell sat in hand uncast; Forest is a land.
	if len(def.DeadCards) != 1 || def.DeadCards[0].Name != "Counterspell" {
		t.Errorf("dead cards %+v", def.DeadCards)
	}
	if def.OtherDecks != nil {
		t.Errorf("other decks %+v with no other deck", def.OtherDecks)
	}
	if got := get("?version=current").Record.Games; got != 2 {
		t.Errorf("current: %d", got)
	}
	if all := get("?version=all"); all.Version != "all" || all.Record.Games != 6 {
		t.Errorf("all time: %s %d", all.Version, all.Record.Games)
	}
	if o := get("?version=" + old); o.Version != old || o.Record.Games != 3 || o.Record.Wins != 1 {
		t.Errorf("old version: %s %+v", o.Version, o.Record)
	}
	if got := get("?version=all&makeup=people").Record.Games; got != 5 {
		t.Errorf("all time, people only: %d", got)
	}
	if got := get("?makeup=bots").Record.Games; got != 1 {
		t.Errorf("current, bots only: %d", got)
	}
	if f := get("?makeup=bots&range=30d"); f.NoGamesSinceEdit {
		t.Error("a filter that keeps a current game says no games since the edit")
	}
	if f := get("?makeup=with_bots"); f.Record.Games != 0 || f.NoGamesSinceEdit {
		t.Errorf("a filter that keeps nothing is not 'no games since the edit': %d %v", f.Record.Games, f.NoGamesSinceEdit)
	}
	if got := get("?version=all&range=30d").Record.Games; got != 6 {
		t.Errorf("all time, 30 days: %d", got)
	}
	for _, bad := range []string{"?version=latest", "?version=ABCDEF0123456789", "?makeup=x"} {
		if code := status(t, s.srv, "GET", path+bad, tok, nil); code != http.StatusBadRequest {
			t.Errorf("%s: %d, want 400", bad, code)
		}
	}

	// Another person's deck, a deck that does not exist and a bad id
	// are all the same 404.
	for _, p := range []string{"/me/stats/decks/" + theirs.ID.String(), "/me/stats/decks/" + uuid.NewString(), "/me/stats/decks/nope"} {
		if code := status(t, s.srv, "GET", p, tok, nil); code != http.StatusNotFound {
			t.Errorf("%s: %d, want 404", p, code)
		}
	}

	// Editing the deck makes a new version with no games yet.
	s.saveDeck(t, user, "Atraxa", atraxaList+"1 Lightning Bolt\n", "Atraxa, Praetors' Voice")
	edited := get("")
	if !edited.NoGamesSinceEdit || edited.Record.Games != 0 || edited.Deck.Facts.Hash == cur {
		t.Errorf("after an edit: no games %v, %d games, hash %s", edited.NoGamesSinceEdit, edited.Record.Games, edited.Deck.Facts.Hash)
	}
	for _, v := range edited.Versions {
		if v.Current {
			t.Errorf("a played version is current after the edit: %+v", v)
		}
	}

	// A second deck gives the composition comparison.
	s.saveDeck(t, user, "Krenko", "Commander:\n1 Krenko, Mob Boss\n\nDeck:\n20 Forest\n", "Krenko, Mob Boss")
	cmp := get("")
	if cmp.OtherDecks == nil || cmp.OtherDecks.Decks != 1 || cmp.OtherDecks.Types[0].Type != "land" || cmp.OtherDecks.Types[0].Average != 20 {
		t.Errorf("other decks %+v", cmp.OtherDecks)
	}
}

// A server with no database has nothing to read: a 503 for a person
// whose advanced stats are on (by the default, with no store).
func TestAdvancedStatsWithNoStore(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	s := newAdminStackWith(t, nil, nil, statsAdminDiscordID)
	tok, _ := s.signedIn(t, "222", "Bea")
	if gamestats.AdvancedDefault {
		if code := status(t, s.srv, "GET", "/me/stats/advanced", tok, nil); code != http.StatusServiceUnavailable {
			t.Errorf("no store: %d, want 503", code)
		}
		return
	}
	if code := status(t, s.srv, "GET", "/me/stats/advanced", tok, nil); code != http.StatusForbidden {
		t.Errorf("no store, default off: %d, want 403", code)
	}
}
