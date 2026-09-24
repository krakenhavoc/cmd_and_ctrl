package aiseat_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// catalog_soak_test.go — bot games as a CATALOG exercise (#601).
//
// The whole-game tests next door prove the ENGINE survives four bots.
// They cannot say anything about the catalog, because the decks they
// deal are hand-written vanilla cards: Mountains, Bears, Ogres and a
// few Bolts. ~1,500 registered specs never see a battlefield.
//
// This file deals decks out of the catalog itself and asks two
// questions a bot game can honestly answer:
//
//  1. did any card's effect ERROR? The engine already emits
//     EventEffectError whenever a catalog primitive fails, and
//     deliberately keeps resolving so a partial failure cannot wedge
//     the stack (effects/primitives.go). Nothing read it, so a card
//     that throws on every resolution was invisible for as long as
//     the table still finished.
//
//  2. which cards did a bot actually REACH? Cast, resolved, entered,
//     triggered — counted per oracle ID from the event log.
//
// What it deliberately does NOT do is check a card did the RIGHT
// thing. That needs a per-card expectation, which is what the ~2,000
// tests in internal/cards/effects already are. This is a net, not an
// oracle: it catches cards that explode and cards nobody can reach,
// and both lists are where the next unit test should be written.
//
// Gated twice over: AISEAT_GAME_TESTS like every whole-game test, and
// CMDCTRL_SCRYFALL_DUMP because a catalog Spec carries an oracle ID
// and a name, not a type line or a mana cost. Without the dump there
// is no card to deal. The nightly sets both.

// catalogPool is the join of the registry and the Scryfall dump: every
// catalog card the deck builder can actually deal.
//
// Face keys ("<oracle>#1") are skipped. They are one face of a card
// whose whole-card key is registered separately when it is registered
// at all, and dealing a back face as a card is not a thing the deck
// path can do.
func catalogPool(t *testing.T, idx *cards.Index) []cards.Card {
	t.Helper()
	var pool []cards.Card
	for _, spec := range effects.All() {
		if strings.Contains(spec.OracleID, "#") {
			continue
		}
		oid, err := uuid.Parse(spec.OracleID)
		if err != nil {
			continue
		}
		c, ok := idx.FindByOracleID(oid)
		if !ok || c.TypeLine == "" {
			continue
		}
		pool = append(pool, c)
	}
	// Deterministic order before any seeded shuffle: effects.All()
	// walks a map, so without this the "same seed, same decks"
	// promise the failure message makes would be a lie.
	sort.Slice(pool, func(i, j int) bool { return pool[i].OracleID.String() < pool[j].OracleID.String() })
	return pool
}

// identityOf is the card's colour identity as a set.
func identityOf(c cards.Card) map[string]bool {
	out := map[string]bool{}
	for _, ch := range c.ColorIdentity {
		out[strings.ToUpper(ch)] = true
	}
	return out
}

// withinIdentity reports CR 903.4: every colour in the card's identity
// appears in the commander's.
func withinIdentity(c cards.Card, id map[string]bool) bool {
	for _, ch := range c.ColorIdentity {
		if !id[strings.ToUpper(ch)] {
			return false
		}
	}
	return true
}

var basicForColor = map[string]string{
	"W": "Plains",
	"U": "Island",
	"B": "Swamp",
	"R": "Mountain",
	"G": "Forest",
}

// catalogDeck deals one seat: a legendary creature out of the catalog
// as commander, catalog cards inside its colour identity, and real
// basics for the mana.
//
// The basics come from the dump rather than being built by hand
// because ProducedMana rides the Scryfall record, and a Mountain that
// produces nothing is a deck that never casts anything.
//
// Not a tournament-legal list: the engine's AddPlayer takes cards, not
// a validated decklist, and singleton is not enforced here. It is a
// deck that CASTS, which is the property this test needs.
func catalogDeck(t *testing.T, idx *cards.Index, pool []cards.Card, rng *rand.Rand, owner uuid.UUID) ([]game.Card, string) {
	t.Helper()

	var legends []cards.Card
	for _, c := range pool {
		if strings.Contains(c.TypeLine, "Legendary Creature") && len(c.ColorIdentity) > 0 {
			legends = append(legends, c)
		}
	}
	if len(legends) == 0 {
		t.Skip("no legendary creature in the catalog/dump join — nothing to command with")
	}
	cmdrCard := legends[rng.IntN(len(legends))]
	id := identityOf(cmdrCard)

	var castable []cards.Card
	for _, c := range pool {
		if c.OracleID == cmdrCard.OracleID || !withinIdentity(c, id) {
			continue
		}
		castable = append(castable, c)
	}
	rng.Shuffle(len(castable), func(i, j int) { castable[i], castable[j] = castable[j], castable[i] })

	const spells, lands = 60, 39
	out := []game.Card{deck.ToGameCard(cmdrCard, true)}
	out[0].Owner, out[0].Controller = owner, owner

	picked := spells
	if len(castable) < picked {
		picked = len(castable)
	}
	for _, c := range castable[:picked] {
		gc := deck.ToGameCard(c, false)
		gc.Owner, gc.Controller = owner, owner
		out = append(out, gc)
	}

	// Basics, spread evenly over the commander's colours. A colourless
	// commander would deal no basics at all, which is why identityOf
	// is required to be non-empty above.
	colors := make([]string, 0, len(id))
	for ch := range id {
		colors = append(colors, ch)
	}
	sort.Strings(colors)
	for n := 0; n < lands+(spells-picked); n++ {
		name := basicForColor[colors[n%len(colors)]]
		bc, ok := idx.FindByName(name)
		if !ok {
			t.Fatalf("basic land %q is not in the dump", name)
		}
		gc := deck.ToGameCard(bc, false)
		gc.Owner, gc.Controller = owner, owner
		out = append(out, gc)
	}
	return out, cmdrCard.Name
}

// newCatalogRoom seats `seats` bots on freshly dealt catalog decks.
func newCatalogRoom(t *testing.T, seats int, seed uint64, idx *cards.Index, pool []cards.Card) *ws.Room {
	t.Helper()
	rng := rand.New(rand.NewPCG(seed, seed+7))
	g := game.NewGame()
	for i := 0; i < seats; i++ {
		d, cmdr := catalogDeck(t, idx, pool, rng, uuid.Nil)
		p, err := g.AddPlayer(fmt.Sprintf("Bot%d", i), d)
		if err != nil {
			t.Fatalf("AddPlayer %d: %v", i, err)
		}
		t.Logf("seat %d (%s): %s", i, p.Name, cmdr)
	}
	if err := g.Start(rand.New(rand.NewPCG(seed, seed+1))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return ws.NewRoom(g, testLogger(), "")
}

// cardStat is one catalog card's tally across every game in a run.
type cardStat struct {
	Name      string `json:"name"`
	OracleID  string `json:"oracle_id"`
	Cast      int    `json:"cast"`
	Resolved  int    `json:"resolved"`
	Entered   int    `json:"entered"`
	Triggered int    `json:"triggered"`
	Errors    int    `json:"errors"`
}

// effectError is one EventEffectError, named.
type effectError struct {
	Card string `json:"card"`
	Msg  string `json:"msg"`
}

// tally walks the finished game's event log and attributes each event
// to the card that caused it.
//
// The identity map is built by walking every zone first: a card that
// resolved is in a graveyard, on the battlefield, or exiled by the
// time we look, and Event carries instance IDs rather than names.
// Tokens have no oracle ID and are counted under their name, which is
// the honest answer — a Treasure is not a catalog card.
func tally(g *game.Game, stats map[string]*cardStat) []effectError {
	var errs []effectError
	g.ReadSnapshot(func() {
		named := map[uuid.UUID]game.Card{}
		note := func(z *game.Zone) {
			if z == nil {
				return
			}
			for _, c := range z.Cards {
				named[c.InstanceID] = c
			}
		}
		note(g.Battlefield)
		note(g.Stack)
		note(g.Exile)
		for _, p := range g.Seats {
			note(p.Library)
			note(p.Hand)
			note(p.Graveyard)
			note(p.Command)
		}

		bump := func(id uuid.UUID, f func(*cardStat)) {
			c, ok := named[id]
			if !ok {
				return
			}
			key := c.OracleID
			if key == "" {
				key = "token:" + c.Name
			}
			st, ok := stats[key]
			if !ok {
				st = &cardStat{Name: c.Name, OracleID: c.OracleID}
				stats[key] = st
			}
			f(st)
		}

		for _, ev := range g.Events {
			switch ev.Kind {
			case game.EventCast:
				bump(ev.CardID, func(s *cardStat) { s.Cast++ })
			case game.EventResolve:
				bump(ev.CardID, func(s *cardStat) { s.Resolved++ })
			case game.EventETB:
				bump(ev.CardID, func(s *cardStat) { s.Entered++ })
			case game.EventTrigger:
				bump(ev.Source, func(s *cardStat) { s.Triggered++ })
			case game.EventEffectError:
				id := ev.Source
				if id == uuid.Nil {
					id = ev.CardID
				}
				name := "(unattributed)"
				if c, ok := named[id]; ok {
					name = c.Name
				}
				bump(id, func(s *cardStat) { s.Errors++ })
				errs = append(errs, effectError{Card: name, Msg: ev.ErrorMsg})
			}
		}
	})
	return errs
}

// catalogRun is what one game produced.
type catalogRun struct {
	state       game.State
	turns       int
	applied     int64
	elapsed     time.Duration
	stalled     bool
	stallKind   game.PendingChoiceKind
	stallKinds  string
	stallReason string
	dump        string
}

// playCatalogGame is playGameIn with the stall RETURNED rather than
// fatal.
//
// The difference matters twice over. A stalled table still played
// thirteen turns of real cards, and t.Fatalf would throw that away
// before the tally runs — the first run of this test stalled on #499
// and reported "0 of 1521 cards reached", which was an artifact of
// the abort, not a fact about the game. And a stall whose cause is a
// filed, known enumerator gap is not this test going red; it is this
// test telling you the gap is still there.
func playCatalogGame(t *testing.T, room *ws.Room, policies []aiseat.Policy, turnBudget int, wall time.Duration) catalogRun {
	t.Helper()
	g := room.Game
	ctx, cancel := context.WithTimeout(context.Background(), wall)
	defer cancel()

	started := time.Now()
	runners := make([]*aiseat.Runner, 0, len(policies))
	for i, p := range g.Seats {
		runners = append(runners, aiseat.Start(ctx, room, p.ID, policies[i], aiseat.Config{}, nil, testLogger()))
	}

	var out catalogRun
	stall := envDuration("AISEAT_STALL", 15*time.Second)
	lastSeq, lastMove := room.Seq(), time.Now()
	for {
		snap := g.Snapshot()
		if snap.State != game.StateActive || snap.Turn.Round > turnBudget {
			break
		}
		if seq := room.Seq(); seq != lastSeq {
			lastSeq, lastMove = seq, time.Now()
		} else if time.Since(lastMove) > stall {
			var pending int
			var kinds []string
			g.ReadSnapshot(func() {
				pending = len(g.PendingChoices)
				seen := map[game.PendingChoiceKind]bool{}
				for _, c := range g.PendingChoices {
					if c == nil || seen[c.Kind] {
						continue
					}
					seen[c.Kind] = true
					kinds = append(kinds, string(c.Kind))
					if out.stallKind == "" {
						out.stallKind, out.stallReason = c.Kind, c.Reason
					}
				}
			})
			sort.Strings(kinds)
			out.stallKinds = strings.Join(kinds, ", ")
			// describeSeats takes the read lock itself, so it is
			// called outside the block above rather than inside it.
			out.stalled = true
			out.dump = fmt.Sprintf("turn %d step %s priority=%d pending=%d kinds=[%s]\n%s",
				snap.Turn.Round, snap.Turn.Step, snap.Turn.PriorityHolder, pending, out.stallKinds, describeSeats(g))
			break
		}
		if ctx.Err() != nil {
			out.stalled = true
			out.dump = fmt.Sprintf("wall clock (%s) exhausted at turn %d", wall, snap.Turn.Round)
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	cancel()
	for _, r := range runners {
		<-r.Done()
		out.applied += r.Stats().Applied
	}
	snap := g.Snapshot()
	out.state, out.turns, out.elapsed = snap.State, snap.Turn.Round, time.Since(started)
	return out
}

// soakBaseSeed is the default first seed: the UTC date as YYYYMMDD,
// times 100.
//
// A fixed default dealt the same four decks every night, so the
// nightly reached the same few dozen cards forever and the report
// never grew. Keying on the date gives each night new decks, so
// coverage accumulates across runs, while a whole day stays
// reproducible: the seed is printed at the top of the log, and every
// failure prints the exact seed to replay.
//
// The x100 leaves each day 100 seeds of its own, so a run of
// AISEAT_CATALOG_GAMES up to 100 never replays tomorrow's first deck.
func soakBaseSeed(now time.Time) uint64 {
	y, m, d := now.UTC().Date()
	return uint64(y*10000+int(m)*100+d) * 100
}

// TestCatalogSoak plays N four-bot games on decks dealt from the
// catalog, fails on any effect error, and reports what the bots
// reached.
//
// Count: AISEAT_CATALOG_GAMES (default 1). Seed: AISEAT_CATALOG_SEED,
// defaulting to one derived from today's UTC date (see soakBaseSeed).
// Report: AISEAT_CATALOG_REPORT, a path to write the per-card JSON to.
func TestCatalogSoak(t *testing.T) {
	requireGameTests(t)
	dump := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if dump == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to deal decks out of the catalog")
	}
	idx := cards.NewIndex()
	n, err := idx.Load(dump)
	if err != nil {
		t.Fatalf("load scryfall dump %s: %v", dump, err)
	}
	pool := catalogPool(t, idx)
	t.Logf("dump: %d cards; catalog pool: %d of %d specs are dealable", n, len(pool), len(effects.All()))
	if len(pool) == 0 {
		t.Fatal("no catalog card resolved against the dump — the join is broken, not empty")
	}

	games := 1
	if v, err := strconv.Atoi(os.Getenv("AISEAT_CATALOG_GAMES")); err == nil && v > 0 {
		games = v
	}
	base := soakBaseSeed(time.Now())
	if v, err := strconv.ParseUint(os.Getenv("AISEAT_CATALOG_SEED"), 10, 64); err == nil {
		base = v
	}
	t.Logf("catalog soak: %d game(s) from base seed %d", games, base)

	const seats, turnBudget = 4, 200
	wall := envDuration("AISEAT_WALLCLOCK", 300*time.Second)
	stats := map[string]*cardStat{}

	for i := 0; i < games; i++ {
		seed := base + uint64(i)
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			room := newCatalogRoom(t, seats, seed, idx, pool)
			policies := make([]aiseat.Policy, seats)
			for s := range policies {
				policies[s] = aiseat.NewRandomPolicy(rand.NewPCG(seed, uint64(s)))
			}
			res := playCatalogGame(t, room, policies, turnBudget, wall)
			// Unconditional, and before any assertion: a game that
			// stalled still exercised every card it managed to play,
			// and those counts are the report.
			errs := tally(room.Game, stats)

			t.Logf("seed %d: state=%s turns=%d applied=%d in %v",
				seed, res.state, res.turns, res.applied, res.elapsed)

			// A stall is REPORTED here, not failed, and the reason is
			// worth stating because the opposite is the obvious
			// choice. Stall-hunting on random tables is already
			// owned next door by TestRandomBotSoak and
			// TestFourRandomBotsPlayToAWinner, which fuzz FIXED decks
			// under pinned seeds — so a stall there is a regression
			// against a stable input and a gate worth having.
			//
			// This test's input is the catalog, which changes every
			// time a card lands. A stall here is a lead with a seed
			// attached, not a regression signal, and gating the
			// nightly on it would mean a card PR turning the build
			// red for an engine bug filed weeks earlier. Both stalls
			// the first runs produced were exactly that: #499's
			// unenumerated choice kinds (fixed since), and #605's
			// wedge on stale commander-zone prompts.
			//
			// The hard assertion is the one below: a card whose
			// effect threw.
			if res.stalled {
				t.Logf("seed %d: table stalled — kinds=[%s] first=%q reason=%q\n%s\nreproduce: AISEAT_CATALOG_SEED=%d AISEAT_CATALOG_GAMES=1",
					seed, res.stallKinds, res.stallKind, res.stallReason, res.dump, seed)
			}

			// The assertion this file exists for. An effect error is a
			// card whose primitive failed mid-resolution: the engine
			// logs it and carries on, so nothing else in the suite
			// would ever go red over it.
			for _, e := range errs {
				t.Errorf("seed %d: %s errored during resolution: %s\nreproduce: AISEAT_CATALOG_SEED=%d AISEAT_CATALOG_GAMES=1",
					seed, e.Card, e.Msg, seed)
			}

			// A game that never cast anything exercised no cards, and
			// would report a clean sheet for the wrong reason.
			var casts int
			for _, s := range stats {
				casts += s.Cast
			}
			if casts == 0 {
				t.Errorf("seed %d: not one spell was cast across %d turns — the decks did not work", seed, res.turns)
			}
		})
	}

	reportCatalogSoak(t, stats, len(pool))
}

// reportCatalogSoak logs the summary and, when asked, writes the whole
// per-card table out as JSON.
//
// The two lists worth acting on are the cards that errored (bugs) and
// the cards nothing reached (where a unit test is worth more than
// another bot game).
func reportCatalogSoak(t *testing.T, stats map[string]*cardStat, pool int) {
	t.Helper()
	rows := make([]*cardStat, 0, len(stats))
	for _, s := range stats {
		rows = append(rows, s)
	}
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.Errors != b.Errors {
			return a.Errors > b.Errors
		}
		if a.Cast != b.Cast {
			return a.Cast > b.Cast
		}
		return a.Name < b.Name
	})

	var touched, errored int
	for _, s := range rows {
		if s.Cast+s.Resolved+s.Entered+s.Triggered > 0 {
			touched++
		}
		if s.Errors > 0 {
			errored++
		}
	}
	t.Logf("catalog soak: %d of %d dealable cards were reached; %d errored", touched, pool, errored)
	for i, s := range rows {
		if i >= 15 {
			break
		}
		t.Logf("  %-34s cast=%d resolved=%d etb=%d trig=%d err=%d", s.Name, s.Cast, s.Resolved, s.Entered, s.Triggered, s.Errors)
	}

	path := os.Getenv("AISEAT_CATALOG_REPORT")
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Errorf("create report dir: %v", err)
		return
	}
	b, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		t.Errorf("marshal report: %v", err)
		return
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Errorf("write report %s: %v", path, err)
		return
	}
	t.Logf("catalog soak: per-card report written to %s", path)
}

// TestSoakBaseSeedRollsDaily pins the two properties the nightly
// relies on: a day is stable, and adjacent days cannot overlap.
func TestSoakBaseSeedRollsDaily(t *testing.T) {
	morning := time.Date(2026, 9, 16, 1, 0, 0, 0, time.UTC)
	evening := time.Date(2026, 9, 16, 23, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 9, 17, 1, 0, 0, 0, time.UTC)

	if got, want := soakBaseSeed(morning), uint64(2026091600); got != want {
		t.Fatalf("soakBaseSeed(2026-09-16) = %d, want %d", got, want)
	}
	if soakBaseSeed(morning) != soakBaseSeed(evening) {
		t.Error("two runs on the same UTC day dealt different decks")
	}
	if soakBaseSeed(tomorrow)-soakBaseSeed(morning) < 100 {
		t.Error("adjacent days are fewer than 100 seeds apart, so a long run would replay tomorrow's decks")
	}
	// Month and year boundaries keep moving forward.
	if soakBaseSeed(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)) <= soakBaseSeed(time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)) {
		t.Error("the seed went backwards across a month boundary")
	}
}
