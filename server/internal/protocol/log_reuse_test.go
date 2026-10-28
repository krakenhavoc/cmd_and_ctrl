package protocol

// log_reuse_test.go holds #3823's reuse of the finished public log to
// the property log_cache_test.go holds the fold to: whatever changed
// between two views, the log a view carries is exactly the log a fresh
// fold finishes against that view, raw and for every viewer
// (viewAndCheck). Reuse is only a shortcut to the same answer: a frame
// whose ring is unchanged and whose names and knower sets resolve the
// same gets the earlier frame's log, and every other frame gets a new
// one.

import (
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sameLog reports whether two views carry one shared log.
func sameLog(a, b GameView) bool {
	return len(a.Log) > 0 && len(b.Log) > 0 && &a.Log[0] == &b.Log[0]
}

func reusedSoFar(t *testing.T, g *game.Game) int {
	t.Helper()
	n := -1
	g.LogProjectionCache().Do(func(slot *any) {
		if c, ok := (*slot).(*logCache); ok {
			n = c.reused
		}
	})
	if n < 0 {
		t.Fatal("no log cached after a view")
	}
	return n
}

func emitEvent(g *game.Game, ev game.Event) {
	g.WithWriteLock(func() { g.EmitEvent(ev) })
}

// knownHandCard is a card in seat i's hand, by instance ID, and makes its
// owner its only knower.
func knownHandCard(t *testing.T, g *game.Game, i int) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.WithWriteLock(func() {
		p := g.Seats[i]
		if len(p.Hand.Cards) == 0 {
			t.Fatalf("seat %d has an empty hand", i)
		}
		c := &p.Hand.Cards[0]
		c.KnownBy = map[uuid.UUID]bool{p.ID: true}
		id = c.InstanceID
	})
	return id
}

// writeHandCard edits a hand card in place, with no event: the change
// reaches the view and nothing reaches the log.
func writeHandCard(g *game.Game, i int, id uuid.UUID, fn func(*game.Card)) {
	g.WithWriteLock(func() {
		for j := range g.Seats[i].Hand.Cards {
			if c := &g.Seats[i].Hand.Cards[j]; c.InstanceID == id {
				fn(c)
			}
		}
	})
}

func TestPublicLogReusesAnUnchangedLog(t *testing.T) {
	g := buildActiveGame(t)
	a := g.Seats[0].ID
	emitEvent(g, game.Event{Kind: game.EventChangeLife, Target: a, Amount: -2})
	v1 := viewAndCheck(t, g, "first view")
	base := reusedSoFar(t, g)

	v2 := viewAndCheck(t, g, "no new events")
	if !sameLog(v1, v2) {
		t.Fatal("a view with nothing new built a new log; want the last one reused")
	}
	// An event the log writes no line for leaves the ring as it was.
	emitEvent(g, game.Event{Kind: game.EventChangeLife, Target: a, Amount: 0})
	v3 := viewAndCheck(t, g, "an event with no line")
	if !sameLog(v1, v3) {
		t.Fatal("an event with no line built a new log; want the last one reused")
	}
	if got := reusedSoFar(t, g) - base; got != 2 {
		t.Errorf("reused the log %d times, want 2", got)
	}

	emitEvent(g, game.Event{Kind: game.EventChangeLife, Target: a, Amount: -1})
	v4 := viewAndCheck(t, g, "a new line")
	if sameLog(v1, v4) {
		t.Fatal("a view after a new line reused the old log")
	}
	if n := len(v4.Log); n != len(v1.Log)+1 {
		t.Fatalf("the log has %d entries after one new line, want %d", n, len(v1.Log)+1)
	}
}

// An event that collapses into the last line (a draw run) edits the
// ring in place, without a push.
func TestPublicLogRefinishesAfterAnInPlaceEdit(t *testing.T) {
	g := buildActiveGame(t)
	a := g.Seats[0].ID
	emitEvent(g, game.Event{Kind: game.EventDrawCard, Actor: a})
	v1 := viewAndCheck(t, g, "one draw")
	emitEvent(g, game.Event{Kind: game.EventDrawCard, Actor: a})
	v2 := viewAndCheck(t, g, "a second draw, folded into the first line")
	if sameLog(v1, v2) {
		t.Fatal("a view after an in-place edit of the last line reused the old log")
	}
	if len(v1.Log) != len(v2.Log) {
		t.Fatalf("the draw run was meant to collapse into one line: %d entries, then %d", len(v1.Log), len(v2.Log))
	}
}

// The hidden-information half: a knower set is compared by content, so
// a card revealed to one more seat, or hidden from it again, with no
// event and no new line, gives every viewer the log a fresh fold would.
func TestPublicLogRefinishesWhenAKnowerSetChanges(t *testing.T) {
	g := buildActiveGame(t)
	owner, other := g.Seats[0].ID, g.Seats[1].ID
	card := knownHandCard(t, g, 0)
	emitEvent(g, game.Event{Kind: game.EventChangeLife, Target: owner, Amount: 3, Source: card})
	named := func(v GameView, viewer uuid.UUID) bool {
		for _, e := range FilterViewFor(v, viewer.String()).Log {
			if e.Kind == LogLife && e.CardID != "" && e.cardName != "" {
				return true
			}
		}
		return false
	}

	v1 := viewAndCheck(t, g, "known to its owner")
	if !named(v1, owner) || named(v1, other) {
		t.Fatalf("before the reveal: owner sees the name %v, opponent %v; want true, false", named(v1, owner), named(v1, other))
	}

	writeHandCard(g, 0, card, func(c *game.Card) { c.KnownBy[other] = true })
	v2 := viewAndCheck(t, g, "revealed to the opponent")
	if sameLog(v1, v2) {
		t.Fatal("a knower set grew and the log was reused")
	}
	if !named(v2, other) {
		t.Fatal("the opponent now knows the card and its log line still hides it")
	}

	writeHandCard(g, 0, card, func(c *game.Card) { delete(c.KnownBy, other) })
	v3 := viewAndCheck(t, g, "hidden again")
	if sameLog(v2, v3) {
		t.Fatal("a knower set shrank and the log was reused")
	}
	if named(v3, other) {
		t.Fatal("the opponent no longer knows the card and its log line still names it")
	}

	// The per-viewer logs are each viewer's own copies, never the
	// shared one.
	v4 := viewAndCheck(t, g, "unchanged")
	if !sameLog(v3, v4) {
		t.Fatal("an unchanged view built a new log")
	}
	for _, viewer := range []uuid.UUID{owner, other} {
		if f := FilterViewFor(v4, viewer.String()); sameLog(f, v4) {
			t.Fatalf("viewer %s's log is the shared one; want its own copy", viewer)
		}
	}
}

func TestPublicLogRefinishesWhenANameChanges(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	card := knownHandCard(t, g, 0)
	emitEvent(g, game.Event{Kind: game.EventChangeLife, Target: owner, Amount: 3, Source: card})
	v1 := viewAndCheck(t, g, "before the rename")

	writeHandCard(g, 0, card, func(c *game.Card) { c.Name = "Renamed In Place" })
	v2 := viewAndCheck(t, g, "after the card's rename")
	if sameLog(v1, v2) {
		t.Fatal("the card a line names was renamed and the log was reused")
	}

	g.WithWriteLock(func() { g.Seats[0].Name = "Renamed Seat" })
	v3 := viewAndCheck(t, g, "after the seat's rename")
	if sameLog(v2, v3) {
		t.Fatal("the seat a line names was renamed and the log was reused")
	}
}

// StampHostOnLog writes on a copy: the log it was handed may be shared
// with the next frame, which must not inherit the stamp.
func TestStampHostOnLogLeavesASharedLogAlone(t *testing.T) {
	g := buildActiveGame(t)
	host := g.Seats[0]
	emitEvent(g, game.Event{
		Kind: game.EventSpawned, Actor: host.ID, Target: host.ID,
		NewZone: game.ZoneBattlefield, Label: "Sol Ring", Amount: 1,
	})
	v1 := ViewOfGame(g)
	shared := v1.Log
	v1.Seats[0].IsHost = true
	StampHostOnLog(&v1)
	if !findLog(t, v1.Log, LogSpawn).ActorIsHost {
		t.Fatal("the stamp did not reach the view it was asked of")
	}
	if findLog(t, shared, LogSpawn).ActorIsHost {
		t.Fatal("the stamp was written into the shared log")
	}
	v2 := ViewOfGame(g)
	if !sameLog(GameView{Log: shared}, v2) {
		t.Fatal("the next frame did not reuse the unstamped log")
	}
	if findLog(t, v2.Log, LogSpawn).ActorIsHost {
		t.Fatal("the next frame inherited the earlier frame's host stamp")
	}
}

// TestLogResolutionFieldsAreAllCompared is the drift guard: every
// unexported LogEvent field is either written by the fold (part of the
// ring, which the version covers) or by the resolver (compared by
// sameLogResolution). A new field must be put in one list; a resolver
// field must be compared, or a reused log keeps its old value.
func TestLogResolutionFieldsAreAllCompared(t *testing.T) {
	fold := []string{"revealSeq", "revealIDs", "revealTo", "batchSeq", "ability", "cardHint", "targetHint"}
	resolved := map[string]func(*LogEvent){
		"cardName":       func(e *LogEvent) { e.cardName = "x" },
		"cardKnowers":    func(e *LogEvent) { e.cardKnowers = map[string]bool{"x": true} },
		"targetName":     func(e *LogEvent) { e.targetName = "x" },
		"targetKnowers":  func(e *LogEvent) { e.targetKnowers = map[string]bool{"x": true} },
		"sourceKnowers":  func(e *LogEvent) { e.sourceKnowers = map[string]bool{"x": true} },
		"actorName":      func(e *LogEvent) { e.actorName = "x" },
		"targetSeatName": func(e *LogEvent) { e.targetSeatName = "x" },
		"seatNames":      func(e *LogEvent) { e.seatNames = []string{"x"} },
		"revealNames":    func(e *LogEvent) { e.revealNames = []string{"x"} },
		"cardFound":      func(e *LogEvent) { e.cardFound = true },
	}
	typ := reflect.TypeFor[LogEvent]()
	for i := range typ.NumField() {
		f := typ.Field(i)
		if f.IsExported() {
			continue
		}
		if _, ok := resolved[f.Name]; !ok && !slices.Contains(fold, f.Name) {
			t.Errorf("LogEvent.%s is in neither list: say whether the fold or the resolver writes it (and compare it in sameLogResolution if the resolver does)", f.Name)
		}
	}
	for name, set := range resolved {
		var a, b LogEvent
		set(&b)
		if sameLogResolution(&a, &b) {
			t.Errorf("sameLogResolution does not compare %s", name)
		}
	}
	// A knower set is compared by content, and nil is not empty.
	if !sameKnowers(map[string]bool{"a": true}, map[string]bool{"a": true}) ||
		sameKnowers(map[string]bool{"a": true}, map[string]bool{"b": true}) ||
		sameKnowers(nil, map[string]bool{}) {
		t.Error("sameKnowers compares knower sets wrongly")
	}
}
