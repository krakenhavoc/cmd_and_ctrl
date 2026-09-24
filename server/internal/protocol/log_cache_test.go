package protocol

// log_cache_test.go holds the public log's resumable fold (#1401) to
// the one property that makes it safe: whatever sequence of views,
// events, undos and restores a game goes through, the log a view
// carries is exactly the log a fold from event 0 would have produced —
// before filtering, and after FilterViewFor for every seat and for a
// spectator.

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// publicLogUncached is the reference: a fresh fold over every event,
// finished against the same view. Caller must hold g's read lock.
func publicLogUncached(g *game.Game, v *GameView) []LogEvent {
	f := newLogFold(seatIDsOf(v), g.EventLogGeneration())
	f.feed(g.Events)
	return finishLog(f.entries(), v)
}

// spectatorID is a viewer seated nowhere.
var spectatorID = uuid.MustParse("5bec7a70-0000-4000-8000-000000000001").String()

// viewAndCheck builds a view the production way (the cached fold) and
// fails unless its log — raw, and filtered for every seat and a
// spectator — is identical to the reference fold's.
func viewAndCheck(t *testing.T, g *game.Game, where string) GameView {
	t.Helper()
	v := ViewOfGame(g)
	var ref []LogEvent
	g.ReadSnapshot(func() { ref = publicLogUncached(g, &v) })
	if !reflect.DeepEqual(v.Log, ref) {
		t.Fatalf("%s: cached log differs from a fresh fold (%d vs %d entries)\ncached: %s\nfresh:  %s",
			where, len(v.Log), len(ref), logDigest(v.Log), logDigest(ref))
	}
	refView := v
	refView.Log = ref
	viewers := []string{spectatorID}
	for _, s := range v.Seats {
		viewers = append(viewers, s.ID)
	}
	for _, viewer := range viewers {
		got, want := FilterViewFor(v, viewer).Log, FilterViewFor(refView, viewer).Log
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: viewer %s: filtered cached log differs from the filtered fresh fold\ncached: %s\nfresh:  %s",
				where, viewer, logDigest(got), logDigest(want))
		}
	}
	return v
}

func logDigest(l []LogEvent) string {
	if len(l) > 6 {
		l = l[len(l)-6:]
	}
	s := ""
	for _, e := range l {
		s += fmt.Sprintf("[%d %s t%d %q] ", e.Seq, e.Kind, e.Turn, e.Text)
	}
	return s
}

// --- synthetic streams -------------------------------------------------

// logFuzzer emits random events chosen to put every piece of the fold's
// carried state across a view boundary: step dedup, draw runs, the
// sacrifice → zone-move suppression, cycling's in-place replace, reveal
// and roll/flip groups (including a group reopened after its head was
// evicted), and turn/step context.
type logFuzzer struct {
	rng     *rand.Rand
	g       *game.Game
	players []uuid.UUID
	cards   []uuid.UUID
	turn    int
	step    int
	// keys are recent RevealSeq / BatchSeq values; reusing an old one
	// is how a group gets extended after its opening entry scrolled
	// out of the ring.
	revealKeys []uint64
	batchKeys  []uint64
	nextKey    uint64
	// follow is the second half of a pair the fold collapses (a
	// sacrifice's zone move, a cycle after its discard), emitted at the
	// start of the NEXT emit so a view can fall between the two.
	follow *game.Event
}

var fuzzSteps = []string{"untap", "upkeep", "draw", "main1", "combat_attackers", "combat_damage", "main2", "end", "cleanup"}

func newLogFuzzer(g *game.Game, seed uint64) *logFuzzer {
	f := &logFuzzer{rng: rand.New(rand.NewPCG(seed, seed^0x1401)), turn: 1, nextKey: 1 << 40}
	f.bind(g)
	return f
}

// bind points the fuzzer at g (a restored game has new *Zone values but
// the same IDs).
func (f *logFuzzer) bind(g *game.Game) {
	f.g = g
	f.players = f.players[:0]
	f.cards = f.cards[:0]
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			f.players = append(f.players, p.ID)
			for _, z := range []*game.Zone{p.Hand, p.Library, p.Graveyard, p.Command} {
				for _, c := range z.Cards {
					f.cards = append(f.cards, c.InstanceID)
				}
			}
		}
		for _, c := range g.Battlefield.Cards {
			f.cards = append(f.cards, c.InstanceID)
		}
	})
}

func (f *logFuzzer) player() uuid.UUID { return f.players[f.rng.IntN(len(f.players))] }
func (f *logFuzzer) card() uuid.UUID   { return f.cards[f.rng.IntN(len(f.cards))] }

var fuzzZones = []game.ZoneKind{game.ZoneBattlefield, game.ZoneGraveyard, game.ZoneExile, game.ZoneHand, game.ZoneLibrary, game.ZoneStack, game.ZoneCommand}

func (f *logFuzzer) zone() game.ZoneKind { return fuzzZones[f.rng.IntN(len(fuzzZones))] }

func (f *logFuzzer) key(keys *[]uint64) uint64 {
	if len(*keys) > 0 && f.rng.IntN(3) > 0 {
		// Mostly the newest open group; sometimes a much older one.
		i := len(*keys) - 1
		if f.rng.IntN(4) == 0 {
			i = f.rng.IntN(len(*keys))
		}
		return (*keys)[i]
	}
	f.nextKey++
	*keys = append(*keys, f.nextKey)
	return f.nextKey
}

// emit adds one or two events under the write lock.
func (f *logFuzzer) emit() {
	f.g.WithWriteLock(func() {
		g := f.g
		if f.follow != nil {
			ev := *f.follow
			f.follow = nil
			if f.rng.IntN(4) > 0 {
				g.EmitEvent(ev)
				return
			}
		}
		switch f.rng.IntN(16) {
		case 0, 1:
			if f.rng.IntN(3) == 0 {
				f.step++
				if f.step == len(fuzzSteps) {
					f.step = 0
					f.turn++
				}
			}
			// Unchanged step half the time: the mulligan dedup.
			g.EmitEvent(game.Event{Kind: game.EventStepBegan, Actor: f.players[f.turn%len(f.players)], Amount: f.turn, Label: fuzzSteps[f.step], Round: f.turn})
		case 2, 3:
			g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: f.player(), CardID: f.card()})
		case 4:
			id := f.card()
			g.EmitEvent(game.Event{Kind: game.EventSacrifice, Actor: f.player(), CardID: id})
			f.follow = &game.Event{Kind: game.EventZoneMove, Actor: f.player(), CardID: id, OldZone: game.ZoneBattlefield, NewZone: game.ZoneGraveyard}
		case 5:
			id := f.card()
			g.EmitEvent(game.Event{Kind: game.EventZoneMove, Actor: f.player(), CardID: id, OldZone: game.ZoneBattlefield, NewZone: game.ZoneGraveyard})
		case 6:
			id, who := f.card(), f.player()
			g.EmitEvent(game.Event{Kind: game.EventZoneMove, Actor: who, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneGraveyard})
			f.follow = &game.Event{Kind: game.EventCycle, Actor: who, CardID: id}
		case 7:
			g.EmitEvent(game.Event{Kind: game.EventZoneMove, Actor: f.player(), CardID: f.card(), OldZone: f.zone(), NewZone: f.zone()})
		case 8:
			ev := game.Event{Kind: game.EventRevealCards, Actor: f.player(), RevealSeq: f.key(&f.revealKeys), OldZone: game.ZoneHand}
			if f.rng.IntN(4) == 0 {
				ev.Target = f.player()
			} else {
				ev.CardID = f.card()
			}
			g.EmitEvent(ev)
		case 9:
			if f.rng.IntN(2) == 0 {
				g.EmitEvent(game.Event{Kind: game.EventRollDie, Actor: f.player(), Source: f.card(), BatchSeq: f.key(&f.batchKeys), Sides: 6, Amount: 1 + f.rng.IntN(6)})
			} else {
				g.EmitEvent(game.Event{Kind: game.EventFlipCoin, Actor: f.player(), Source: f.card(), BatchSeq: f.key(&f.batchKeys), Label: "heads", Won: f.rng.IntN(2) == 0})
			}
		case 10:
			g.EmitEvent(game.Event{Kind: game.EventCast, Actor: f.player(), CardID: f.card(), OldZone: game.ZoneHand})
		case 11:
			g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: f.player(), CardID: f.card()})
		case 12:
			g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: f.player(), Source: f.card(), Target: f.player(), Amount: 1 + f.rng.IntN(5)})
		case 13:
			g.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: f.player(), Amount: f.rng.IntN(7) - 3})
		case 14:
			g.EmitEvent(game.Event{Kind: game.EventAttack, Actor: f.player(), CardID: f.card(), Target: f.player()})
		default:
			// Silent kinds still advance the cursor and clear the
			// sacrifice marker.
			g.EmitEvent(game.Event{Kind: game.EventTapCard, Actor: f.player(), CardID: f.card()})
		}
	})
}

// TestPublicLogCacheMatchesAFreshFold is the property test: random
// event streams, viewed at random points, with random undos (to a
// shallower and a deeper point than the fold's cursor, then regrown)
// and persisted-snapshot restores. Every view's log must equal a fresh
// fold's, raw and per viewer.
func TestPublicLogCacheMatchesAFreshFold(t *testing.T) {
	const events = 800
	for seed := uint64(1); seed <= 3; seed++ {
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			g := fourPlayerBoard(t)
			fz := newLogFuzzer(g, seed)
			var undo []*game.Game
			views, undos, restores := 0, 0, 0
			for i := 0; i < events; i++ {
				fz.emit()
				switch r := fz.rng.IntN(100); {
				case r < 15:
					viewAndCheck(t, g, fmt.Sprintf("event %d", i))
					views++
				case r < 30:
					undo = append(undo, g.Clone())
				case r < 33 && len(undo) > 0:
					// Pop to a random depth, the way repeated undos do.
					at := fz.rng.IntN(len(undo))
					snap := undo[at]
					undo = undo[:at]
					g.WithWriteLock(func() { g.RestoreFrom(snap) })
					undos++
				case r < 34:
					restored, err := g.CaptureSnapshot().Restore()
					if err != nil {
						t.Fatalf("restore: %v", err)
					}
					g = restored
					fz.bind(g)
					undo = nil
					restores++
				}
			}
			viewAndCheck(t, g, "end")
			if views == 0 || undos == 0 || restores == 0 {
				t.Fatalf("the stream exercised views=%d undos=%d restores=%d; every one must be non-zero", views, undos, restores)
			}
		})
	}
}

// TestPublicLogCacheSeesAnUndoThatRegrowsToTheSameLength is the case
// the generation exists for. An undo truncates the log and eventSeq
// rewinds with it, so new events regrow the log to the SAME length
// with the SAME Seq at the cursor. Only the generation tells that log
// from the one the fold saw.
func TestPublicLogCacheSeesAnUndoThatRegrowsToTheSameLength(t *testing.T) {
	g := buildActiveGame(t)
	a, b := g.Seats[0].ID, g.Seats[1].ID
	pre := g.Clone()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: a, Amount: -3})
	})
	before := viewAndCheck(t, g, "before the undo")

	g.WithWriteLock(func() {
		g.RestoreFrom(pre)
		g.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: b, Amount: 5})
	})
	after := viewAndCheck(t, g, "after the undo and regrowth")

	if len(before.Log) != len(after.Log) || before.Log[len(before.Log)-1].Seq != after.Log[len(after.Log)-1].Seq {
		t.Fatalf("the regrown log is meant to match the old one in length and last Seq; got %d/%d entries", len(before.Log), len(after.Log))
	}
	last := after.Log[len(after.Log)-1]
	if last.Kind != LogLife || last.Amount != 5 || last.Seat != 1 {
		t.Fatalf("the last entry is %+v; want seat 1 gaining 5 — the undone line survived the undo", last)
	}
}

// TestPublicLogCacheRefoldsWhenTheSeatsChange: a fold's entries carry
// seat INDICES, so a view whose seat order differs from the fold's must
// not reuse it.
func TestPublicLogCacheRefoldsWhenTheSeatsChange(t *testing.T) {
	g := buildActiveGame(t)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: g.Seats[1].ID, Amount: -2})
	})
	v := ViewOfGame(g)
	swapped := v
	swapped.Seats = []PlayerView{v.Seats[1], v.Seats[0]}
	var got, want []LogEvent
	g.ReadSnapshot(func() {
		got = publicLogOf(g, &swapped)
		want = publicLogUncached(g, &swapped)
	})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("a view with a different seat order reused the fold\ngot:  %s\nwant: %s", logDigest(got), logDigest(want))
	}
}

// --- played games ------------------------------------------------------

// logCacheDeck is a small deck the enumerator can play end to end:
// lands, vanilla creatures to attack with, and the catalog's Lightning
// Bolt for targets, damage and the stack.
func logCacheDeck() []game.Card {
	deck := []game.Card{}
	cmdr := game.NewCommander("Commander Bear", uuid.Nil)
	cmdr.TypeLine = "Legendary Creature — Bear"
	cmdr.ManaCost = "{2}{R}"
	cmdr.Power, cmdr.Toughness = 3, 3
	deck = append(deck, cmdr)
	for i := 0; i < 14; i++ {
		c := game.NewCard("Mountain", uuid.Nil)
		c.TypeLine = "Basic Land — Mountain"
		deck = append(deck, c)
	}
	for i := 0; i < 10; i++ {
		c := game.NewCard("Bear", uuid.Nil)
		c.TypeLine = "Creature — Bear"
		c.ManaCost = "{1}{R}"
		c.Power, c.Toughness = 2, 2
		deck = append(deck, c)
	}
	for i := 0; i < 6; i++ {
		c := game.NewCard("Lightning Bolt", uuid.Nil)
		c.TypeLine = "Instant"
		c.ManaCost = "{R}"
		c.OracleID = "4457ed35-7c10-48c8-9776-456485fdf070"
		deck = append(deck, c)
	}
	return deck
}

// randomMove picks a uniformly random seat among those with a decision,
// then a uniformly random move of that seat's — what a table of random
// bots does, without the goroutines.
func randomMove(g *game.Game, rng *rand.Rand) (legal.Move, uuid.UUID, bool) {
	type seatMoves struct {
		seat  uuid.UUID
		moves []legal.Move
	}
	var open []seatMoves
	for _, p := range g.Seats {
		if ms := legal.EnumerateFor(g, p.ID); len(ms) > 0 {
			open = append(open, seatMoves{p.ID, ms})
		}
	}
	if len(open) == 0 {
		return legal.Move{}, uuid.Nil, false
	}
	s := open[rng.IntN(len(open))]
	return s.moves[rng.IntN(len(s.moves))], s.seat, true
}

// TestPublicLogCacheMatchesAFreshFoldInPlayedGames runs the same
// property over real games: random legal moves from the enumerator,
// dispatched through actions like a bot's, with a view after most of
// them, undos through the same Clone / RestoreFrom pair ws.Room uses,
// and a persisted-snapshot restore partway through.
func TestPublicLogCacheMatchesAFreshFoldInPlayedGames(t *testing.T) {
	const moves = 1200
	for seed := uint64(1); seed <= 2; seed++ {
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			g := game.NewGame()
			for i := 0; i < 4; i++ {
				if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), logCacheDeck()); err != nil {
					t.Fatalf("AddPlayer: %v", err)
				}
			}
			if err := g.Start(rand.New(rand.NewPCG(seed, seed+1))); err != nil {
				t.Fatalf("Start: %v", err)
			}
			rng := rand.New(rand.NewPCG(seed, 1401))
			var undo []*game.Game
			applied, undos, restores := 0, 0, 0
			for i := 0; i < moves; i++ {
				mv, seat, ok := randomMove(g, rng)
				if !ok {
					break
				}
				pre := g.Clone()
				err := actions.Dispatch(g, actions.Action{Type: actions.Type(mv.Type), Player: mv.Player, Caller: seat, Params: mv.Params})
				if err != nil {
					continue
				}
				applied++
				undo = append(undo, pre)
				switch r := rng.IntN(100); {
				case r < 6 && len(undo) > 1:
					at := len(undo) - 1 - rng.IntN(min(len(undo), 4))
					snap := undo[at]
					undo = undo[:at]
					g.WithWriteLock(func() { g.RestoreFrom(snap) })
					undos++
				case r < 7 && restores == 0:
					restored, err := g.CaptureSnapshot().Restore()
					if err != nil {
						t.Fatalf("restore: %v", err)
					}
					g = restored
					undo = nil
					restores++
				}
				if rng.IntN(8) == 0 {
					viewAndCheck(t, g, fmt.Sprintf("move %d (%s)", i, mv.Label))
				}
			}
			viewAndCheck(t, g, "end")
			var n int
			g.ReadSnapshot(func() { n = len(g.Events) })
			t.Logf("seed %d: %d moves applied, %d undos, %d restores, %d events", seed, applied, undos, restores, n)
			if applied < 100 || undos == 0 {
				t.Fatalf("the game exercised applied=%d undos=%d; too thin to be a property test", applied, undos)
			}
		})
	}
}

// --- cost ----------------------------------------------------------------

// foldedSoFar reads how many events this game's folds have projected,
// across refolds.
func foldedSoFar(t *testing.T, g *game.Game) int {
	t.Helper()
	n := -1
	g.LogProjectionCache().Do(func(slot *any) {
		if c, ok := (*slot).(*logCache); ok {
			n = c.folded
		}
	})
	if n < 0 {
		t.Fatal("no fold cached after a view")
	}
	return n
}

// TestPublicLogViewFoldsOnlyNewEvents is #1401's acceptance test. At two
// game lengths, a view after k new events must project exactly k events
// — not the whole history — so per-frame cost no longer grows with the
// length of the game.
func TestPublicLogViewFoldsOnlyNewEvents(t *testing.T) {
	const k = 7
	perFrame := map[int]int{}
	for _, history := range []int{500, 5000} {
		g := fourPlayerBoard(t)
		fz := newLogFuzzer(g, uint64(history))
		for i := 0; i < history; i++ {
			fz.emit()
		}
		ViewOfGame(g)
		base := foldedSoFar(t, g)
		var total int
		g.ReadSnapshot(func() { total = len(g.Events) })
		if base != total {
			t.Fatalf("history %d: the first view folded %d events, want all %d", history, base, total)
		}

		// A frame with nothing new folds nothing.
		ViewOfGame(g)
		if got := foldedSoFar(t, g) - base; got != 0 {
			t.Errorf("history %d: a view with no new events folded %d", history, got)
		}

		g.WithWriteLock(func() {
			for i := 0; i < k; i++ {
				g.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: g.Seats[i%4].ID, Amount: -1})
			}
		})
		ViewOfGame(g)
		perFrame[history] = foldedSoFar(t, g) - base
		if perFrame[history] != k {
			t.Errorf("history %d: view N+1 folded %d events after %d were emitted; want exactly %d", history, perFrame[history], k, k)
		}
	}
	if perFrame[500] != perFrame[5000] {
		t.Errorf("per-frame fold work grows with the game: %d events at 500, %d at 5000", perFrame[500], perFrame[5000])
	}
}

// TestPublicLogViewRefoldsAfterAnUndo pins the other side of the
// bargain: the view after an undo pays for the whole log once, and the
// one after that is incremental again.
func TestPublicLogViewRefoldsAfterAnUndo(t *testing.T) {
	g := fourPlayerBoard(t)
	fz := newLogFuzzer(g, 9)
	for i := 0; i < 300; i++ {
		fz.emit()
	}
	pre := g.Clone()
	fz.emit()
	ViewOfGame(g)
	base := foldedSoFar(t, g)
	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	ViewOfGame(g)
	var total int
	g.ReadSnapshot(func() { total = len(g.Events) })
	if got := foldedSoFar(t, g) - base; got != total {
		t.Fatalf("the view after an undo folded %d events; want a full refold of %d", got, total)
	}
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: g.Seats[0].ID, Amount: 1}) })
	ViewOfGame(g)
	if got := foldedSoFar(t, g) - base - total; got != 1 {
		t.Fatalf("the next view folded %d events; want 1 (incremental again)", got)
	}
}

// TestPublicLogViewsShareNoGrowableSlices: the fold keeps growing a roll
// batch's Results in place after a view has been handed out, so each
// view must get its own copy. Otherwise two views — possibly on two
// goroutines, the room's and a bot seat's — share a backing array the
// fold is still appending into.
func TestPublicLogViewsShareNoGrowableSlices(t *testing.T) {
	g := buildActiveGame(t)
	a := g.Seats[0].ID
	roll := func(n int) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventRollDie, Actor: a, BatchSeq: 77, Sides: 6, Amount: n})
		})
	}
	lastRoll := func(v GameView) LogEvent {
		for i := len(v.Log) - 1; i >= 0; i-- {
			if v.Log[i].Kind == LogRoll {
				return v.Log[i]
			}
		}
		t.Fatal("no roll entry in the log")
		return LogEvent{}
	}
	roll(1)
	roll(2)
	roll(3)
	first := lastRoll(ViewOfGame(g))
	second := lastRoll(ViewOfGame(g))
	if &first.Results[0] == &second.Results[0] {
		t.Fatal("two views share one Results backing array")
	}
	roll(4)
	third := lastRoll(ViewOfGame(g))
	if !reflect.DeepEqual(first.Results, []int{1, 2, 3}) || !reflect.DeepEqual(third.Results, []int{1, 2, 3, 4}) {
		t.Fatalf("results: first view %v, third view %v", first.Results, third.Results)
	}
}
