package protocol

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_library_search_test.go — #4447, ADR 0209 §3: the client's
// library_search cue riffles the library pile of a player whose library
// was searched (CR 701.23a). A library is a hidden zone, a face-down pile
// no player may look at (CR 400.2, 401.2): a search tells the table THAT a
// library was searched and whose, and never what was found, how many, or
// whether anything was (CR 701.23b), unless the effect reveals it
// (CR 701.23e). The log already wrote a `search` line; it now carries,
// additively, `library_seat`, whose library it was. When the search
// revealed what it took (a fetch land, a "reveal it" tutor, Bribery) the
// fold has always dropped the search line, because the `reveal` line
// before it already says "fetched" or "revealed"; that reveal line now
// carries `library_seat` instead, so every search is exactly one entry
// with it. This pins, for every viewer, an opponent included:
//
//   - a tutor to a hand (hidden): one `search` entry, no card reference,
//     and the found card's ID and name on no entry an opponent sees;
//   - a tutor to the battlefield with no reveal: the `search` entry, and
//     the card's own public `zone` entry onto the battlefield;
//   - a fetch land: its sacrifice is a `sacrifice` entry when it is
//     activated, in a batch of its own; on resolution the "fetched" reveal
//     carries `library_seat`, the found land has its `zone` entry, and
//     there is no second search line; a fetched shock land's entry
//     prompt comes between the reveal and the search line, and the
//     reveal carries `library_seat` from the frame the prompt is open;
//   - a search that finds nothing, then shuffles: the `search` entry;
//   - a search of another player's library: `library_seat` is that
//     player, `seat` the searcher, with a reveal (Bribery) or without;
//   - one search that finds several cards: still one entry;
//   - a plain shuffle is no search and has no `library_seat`.

// searchEntries is every entry viewer sees that carries library_seat.
func searchEntries(g *game.Game, viewer string) []LogEvent {
	var out []LogEvent
	for _, e := range ViewOfGameFor(g, viewer).Log {
		if e.LibrarySeat != nil {
			out = append(out, e)
		}
	}
	return out
}

// stackLibrary puts a new card named name, owned by p, on top of p's
// library, and returns its ID.
func stackLibrary(g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	g.WithWriteLock(func() { p.Library.PushTop(c) })
	return c.InstanceID
}

// answerSearch answers the one open search prompt as chooser, taking picks.
func answerSearch(t *testing.T, g *game.Game, chooser *game.Player, picks ...uuid.UUID) {
	t.Helper()
	var open *game.PendingChoice
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceSearchLibrary {
				open = c
			}
		}
	})
	if open == nil {
		t.Fatal("no search prompt is open")
	}
	if err := g.ResolveSearchLibrary(open.ID, chooser.ID, picks); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
}

// named is a predicate for the card named name.
func named(name string) func(game.Card) bool {
	return func(c game.Card) bool { return c.Name == name }
}

// assertOneSearch checks that every viewer sees exactly one entry with
// library_seat, of kind `kind`, naming `library` and searched by
// `searcher`, in a non-zero batch, with the wire fields the cue reads,
// and returns what the first viewer saw.
func assertOneSearch(t *testing.T, g *game.Game, kind LogKind, searcher, library int) LogEvent {
	t.Helper()
	var first LogEvent
	for i, viewer := range allViewers(g) {
		got := searchEntries(g, viewer)
		if len(got) != 1 {
			t.Fatalf("viewer %q: %d entries with library_seat, want 1: %+v", viewer, len(got), got)
		}
		e := got[0]
		raw := wireFields(t, e)
		want := map[string]any{
			"kind":         string(kind),
			"seat":         float64(searcher),
			"library_seat": float64(library),
		}
		for key, w := range want {
			if raw[key] != w {
				t.Errorf("viewer %q: wire field %q = %v, want %v (entry %v)", viewer, key, raw[key], w, raw)
			}
		}
		if e.Batch == 0 {
			t.Errorf("viewer %q: the search entry has no batch: %+v", viewer, e)
		}
		if e.CardID != "" || len(e.CardIDs) != 0 || e.Target != "" {
			t.Errorf("viewer %q: the search entry carries a card reference: %+v", viewer, e)
		}
		assertNoUUID(t, e.Text)
		if i == 0 {
			first = e
		}
	}
	return first
}

// assertHiddenFrom fails if viewer's log names card id, or a card called
// name, anywhere: an ID field, or a line's text.
func assertHiddenFrom(t *testing.T, g *game.Game, viewer string, id uuid.UUID, name string) {
	t.Helper()
	for _, e := range ViewOfGameFor(g, viewer).Log {
		s := id.String()
		if e.CardID == s || e.Target == s || e.Source == s || e.StackItemID == s || slices.Contains(e.CardIDs, s) {
			t.Errorf("viewer %q: a %s entry carries the found card's ID: %+v", viewer, e.Kind, e)
		}
		if strings.Contains(e.Text, name) {
			t.Errorf("viewer %q: a %s entry names the found card: %q", viewer, e.Kind, e.Text)
		}
	}
}

func TestALibrarySearchIsOneEntryWithItsLibrarySeat(t *testing.T) {
	t.Run("a tutor to a hand names no card to an opponent", func(t *testing.T) {
		g := keptGame(t, 0)
		me, opp := g.Seats[0], g.Seats[1]
		found := stackLibrary(g, me, "Secret Tutor Target", "Sorcery")
		var err error
		g.WithWriteLock(func() {
			err = g.SearchLibraryForEffect(me.ID, named("Secret Tutor Target"), game.ZoneHand, 1, false, true)
		})
		if err != nil {
			t.Fatalf("SearchLibraryForEffect: %v", err)
		}
		answerSearch(t, g, me, found)
		e := assertOneSearch(t, g, LogSearch, me.Seat, me.Seat)
		if e.Text != "P1 searched their library and shuffled it" || e.Label != game.SearchShuffledLabel {
			t.Errorf("search entry %q label %q", e.Text, e.Label)
		}
		if e.TargetSeat == nil || *e.TargetSeat != me.Seat {
			t.Errorf("target_seat %v, want the searcher's own library %d", e.TargetSeat, me.Seat)
		}
		// The redaction check: an opponent, and a spectator, learn that P1
		// searched, and nothing about what P1 found.
		assertHiddenFrom(t, g, opp.ID.String(), found, "Secret Tutor Target")
		assertHiddenFrom(t, g, "", found, "Secret Tutor Target")
		// The card did reach the hand.
		var inHand bool
		g.ReadSnapshot(func() { inHand = me.Hand.Contains(found) })
		if !inHand {
			t.Fatal("the tutor did not put the card into the hand")
		}
	})

	t.Run("a tutor to the battlefield with no reveal", func(t *testing.T) {
		g := keptGame(t, 0)
		me := g.Seats[0]
		found := stackLibrary(g, me, "Found Giant", "Creature — Giant")
		var err error
		g.WithWriteLock(func() {
			err = g.SearchLibraryForEffect(me.ID, named("Found Giant"), game.ZoneBattlefield, 1, false, true)
		})
		if err != nil {
			t.Fatalf("SearchLibraryForEffect: %v", err)
		}
		answerSearch(t, g, me, found)
		e := assertOneSearch(t, g, LogSearch, me.Seat, me.Seat)
		// The card is public once it is on the battlefield: its own zone
		// entry names it, for every viewer, and is what the arrival reads.
		for _, viewer := range allViewers(g) {
			var zones []LogEvent
			for _, z := range logEntriesOfKind(ViewOfGameFor(g, viewer).Log, LogZone) {
				if z.CardID == found.String() {
					zones = append(zones, z)
				}
			}
			if len(zones) != 1 || zones[0].OldZone != "library" || zones[0].NewZone != "battlefield" {
				t.Fatalf("viewer %q: zone entries for the found card: %+v", viewer, zones)
			}
			if zones[0].Batch != e.Batch {
				t.Errorf("viewer %q: the arrival's batch %d, the search's %d: want one moment", viewer, zones[0].Batch, e.Batch)
			}
		}
	})

	t.Run("a search that finds nothing, then shuffles", func(t *testing.T) {
		g := keptGame(t, 0)
		me := g.Seats[0]
		var err error
		g.WithWriteLock(func() {
			err = g.SearchLibraryForEffect(me.ID, func(game.Card) bool { return false }, game.ZoneHand, 1, false, true)
		})
		if err != nil {
			t.Fatalf("SearchLibraryForEffect: %v", err)
		}
		answerSearch(t, g, me)
		e := assertOneSearch(t, g, LogSearch, me.Seat, me.Seat)
		if e.Text != "P1 searched their library and shuffled it" {
			t.Errorf("text %q", e.Text)
		}
	})

	t.Run("a search that reveals what it found carries library_seat on its reveal", func(t *testing.T) {
		g := keptGame(t, 0)
		me, opp := g.Seats[0], g.Seats[1]
		found := stackLibrary(g, me, "Shown Forest", "Basic Land — Forest")
		var err error
		g.WithWriteLock(func() {
			err = g.SearchLibraryForEffect(me.ID, named("Shown Forest"), game.ZoneHand, 1, true, true)
		})
		if err != nil {
			t.Fatalf("SearchLibraryForEffect: %v", err)
		}
		answerSearch(t, g, me, found)
		e := assertOneSearch(t, g, LogReveal, me.Seat, me.Seat)
		// Revealed, so the whole table may read its name (CR 701.23e).
		if e.Text != "P1 revealed Shown Forest from their library" {
			t.Errorf("reveal text %q", e.Text)
		}
		if n := len(logEntriesOfKind(ViewOfGameFor(g, opp.ID.String()).Log, LogSearch)); n != 0 {
			t.Errorf("the search was said twice: %d search lines beside its reveal", n)
		}
	})

	t.Run("several cards found by one search are one entry", func(t *testing.T) {
		g := keptGame(t, 0)
		me := g.Seats[0]
		a := stackLibrary(g, me, "Claimed Forest", "Basic Land — Forest")
		b := stackLibrary(g, me, "Claimed Forest", "Basic Land — Forest")
		var err error
		g.WithWriteLock(func() {
			err = g.SearchLibraryForEffect(me.ID, named("Claimed Forest"), game.ZoneBattlefield, 2, true, true)
		})
		if err != nil {
			t.Fatalf("SearchLibraryForEffect: %v", err)
		}
		answerSearch(t, g, me, a, b)
		e := assertOneSearch(t, g, LogReveal, me.Seat, me.Seat)
		if e.Amount != 2 || e.Text != "P1 fetched 2 cards from their library onto the battlefield: Claimed Forest, Claimed Forest" {
			t.Errorf("reveal amount %d text %q", e.Amount, e.Text)
		}
	})

	t.Run("a search of another player's library names that library", func(t *testing.T) {
		for _, reveal := range []bool{false, true} {
			g := keptGame(t, 0)
			me, opp := g.Seats[0], g.Seats[1]
			found := stackLibrary(g, opp, "Bribed Titan", "Creature — Giant")
			var err error
			g.WithWriteLock(func() {
				err = g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
					Player:       me.ID,
					LibraryOwner: opp.ID,
					Pred:         named("Bribed Titan"),
					Dest:         game.ZoneBattlefield,
					Limit:        1,
					Reveal:       reveal,
					Shuffle:      true,
				})
			})
			if err != nil {
				t.Fatalf("SearchLibraryThenForEffect: %v", err)
			}
			answerSearch(t, g, me, found)
			kind := LogSearch
			if reveal {
				kind = LogReveal
			}
			e := assertOneSearch(t, g, kind, me.Seat, opp.Seat)
			if !reveal && e.Text != "P1 searched P2's library and shuffled it" {
				t.Errorf("text %q", e.Text)
			}
		}
	})

	t.Run("a plain shuffle is no search", func(t *testing.T) {
		g := keptGame(t, 0)
		me := g.Seats[0]
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventSearchLibrary, Actor: me.ID, Target: me.ID, Label: "shuffle"})
		})
		for _, viewer := range allViewers(g) {
			if got := searchEntries(g, viewer); len(got) != 0 {
				t.Errorf("viewer %q: a plain shuffle carries library_seat: %+v", viewer, got)
			}
			shuffles := logEntriesOfKind(ViewOfGameFor(g, viewer).Log, LogSearch)
			if len(shuffles) != 1 || shuffles[0].Text != "P1 shuffled their library" {
				t.Errorf("viewer %q: shuffle entries %+v", viewer, shuffles)
			}
		}
	})
}

// A fetch land, played for real: the sacrifice is its cost, logged as it
// is activated (the self-sacrifice cue's moment, #4427); the search is
// its effect, logged as it resolves, in a later batch.
func TestAFetchLandIsASacrificeThenOneSearchEntry(t *testing.T) {
	g, me, opp := attachTable(t)
	wilds := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: wilds, Name: "Evolving Wilds", TypeLine: "Land",
			OracleID: "a75445d3-1303-4bb5-89ad-26ea93fecd48",
			Owner:    me.ID, Controller: me.ID,
			KnownBy: map[uuid.UUID]bool{me.ID: true, opp.ID: true},
		})
	})
	forest := stackLibrary(g, me, "Fetched Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, wilds, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Evolving Wilds: %v", err)
	}
	sacs := logEntriesOfKind(ViewOfGameFor(g, opp.ID.String()).Log, LogSacrifice)
	if len(sacs) != 1 || sacs[0].CardID != wilds.String() {
		t.Fatalf("the activation's sacrifice entries: %+v", sacs)
	}
	if got := searchEntries(g, opp.ID.String()); len(got) != 0 {
		t.Fatalf("the activation already searched: %+v", got)
	}
	for i := 0; len(g.StackMeta) > 0; i++ {
		if i > 8 {
			t.Fatal("the fetch ability never resolved")
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	answerSearch(t, g, me, forest)
	e := assertOneSearch(t, g, LogReveal, me.Seat, me.Seat)
	if e.Text != "P1 fetched Fetched Forest from their library onto the battlefield" {
		t.Errorf("reveal text %q", e.Text)
	}
	if e.Batch == sacs[0].Batch {
		t.Errorf("the search shares the sacrifice's batch %d: want the cost and the effect apart", e.Batch)
	}
	for _, viewer := range allViewers(g) {
		log := ViewOfGameFor(g, viewer).Log
		if n := len(logEntriesOfKind(log, LogSearch)); n != 0 {
			t.Errorf("viewer %q: %d search lines beside the fetch's reveal", viewer, n)
		}
		var arrived bool
		for _, z := range logEntriesOfKind(log, LogZone) {
			if z.CardID == forest.String() && z.NewZone == "battlefield" && z.OldZone == "library" {
				arrived = true
				if z.Batch != e.Batch {
					t.Errorf("viewer %q: the land's arrival batch %d, the search's %d: want one moment", viewer, z.Batch, e.Batch)
				}
			}
		}
		if !arrived {
			t.Errorf("viewer %q: no zone entry for the fetched land", viewer)
		}
	}
}

// A fetched shock land asks "pay 2 life?" as it enters, and the search
// finishes (its line, its shuffle) only once that is answered, so a frame
// goes out between the reveal and the search line. The reveal says it was
// a search, and whose library, from that first frame on, and the answer
// adds no second entry.
func TestAFetchedShockLandIsOneSearchEntryAcrossItsPrompt(t *testing.T) {
	g, me, opp := attachTable(t)
	delta := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: delta, Name: "Polluted Delta", TypeLine: "Land",
			OracleID: "ef86989d-ce80-4e55-aece-7d11710eeffa",
			Owner:    me.ID, Controller: me.ID,
			KnownBy: map[uuid.UUID]bool{me.ID: true, opp.ID: true},
		})
	})
	grave := uuid.Nil
	g.WithWriteLock(func() {
		c := game.NewCard("Watery Grave", me.ID)
		c.TypeLine = "Land — Island Swamp"
		c.OracleID = "fc9ec820-4245-4a96-b009-5308a818ca58"
		me.Library.PushTop(c)
		grave = c.InstanceID
	})
	if err := g.ActivateCatalogAbility(me.ID, delta, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Polluted Delta: %v", err)
	}
	for i := 0; len(g.StackMeta) > 0; i++ {
		if i > 8 {
			t.Fatal("the fetch ability never resolved")
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	answerSearch(t, g, me, grave)
	var shock *game.PendingChoice
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceEntryPayLife {
				shock = c
			}
		}
	})
	if shock == nil {
		t.Fatal("the fetched shock land did not ask to pay 2 life")
	}
	// The frame that goes out with the prompt open.
	before := assertOneSearch(t, g, LogReveal, me.Seat, me.Seat)
	if err := g.ResolveEntryPayLife(shock.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveEntryPayLife: %v", err)
	}
	after := assertOneSearch(t, g, LogReveal, me.Seat, me.Seat)
	if after.Seq != before.Seq {
		t.Errorf("the search entry moved from seq %d to %d across the prompt", before.Seq, after.Seq)
	}
	if n := len(logEntriesOfKind(ViewOfGameFor(g, opp.ID.String()).Log, LogSearch)); n != 0 {
		t.Errorf("%d search lines beside the fetch's reveal after the prompt", n)
	}
}
