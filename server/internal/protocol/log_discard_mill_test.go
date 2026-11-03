package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_discard_mill_test.go — #4445, ADR 0209 §3: the client's mill and
// discard cues ("cards go to the graveyard without dying") read two log
// lines, which this file pins for every viewer, an opponent included:
//
//   - a `mill` entry (CR 701.17a): `seat` the player milled, `amount`
//     the count, `batch`, and, additive, `card_ids`, the milled cards
//     the text names, so the cue can slide the cards themselves;
//   - a `discard` entry (CR 701.9a), new: one per card, with `card_id`,
//     `seat`, `old_zone` "hand", `new_zone` where it went, `cause`
//     ("effect", "cleanup" or "cost") and `batch`. Since #650 a discard
//     emitted EventDiscardCard instead of a zone move, and the log said
//     nothing at all about it.
//
// Neither is a `zone` entry from the battlefield, so neither can be
// read as a creature dying, which has its own cue. A commander that
// was milled or discarded is in its graveyard first, and its owner's
// CR 903.9a choice to send it home is a `zone` entry of its own after
// that one, which the commander_return cue reads. It can share the
// mill's or the discard's batch (nothing between them moves play on,
// CR 603.2c); the client caps a sound per name and batch, so the two
// cues' sounds both play.

// keptGame is an active two-seat game with both hands kept, and the
// number of cards in the first seat's library topped up to `library`.
func keptGame(t *testing.T, library int) *game.Game {
	t.Helper()
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	p := g.Seats[0]
	g.WithWriteLock(func() {
		for i := 0; p.Library.Size() < library; i++ {
			p.Library.PushBottom(game.NewCard("Spare "+string(rune('A'+i)), p.ID))
		}
	})
	return g
}

// handIDs is the first n cards of p's hand.
func handIDs(g *game.Game, p *game.Player, n int) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range p.Hand.Cards[:n] {
			out = append(out, c.InstanceID)
		}
	})
	return out
}

// topIDs is the top n cards of p's library, top first: the order a mill
// takes them in.
func topIDs(g *game.Game, p *game.Player, n int) []string {
	var out []string
	g.ReadSnapshot(func() {
		size := p.Library.Size()
		for i := 0; i < n; i++ {
			out = append(out, p.Library.Cards[size-1-i].InstanceID.String())
		}
	})
	return out
}

// discardsSeenBy is every discard entry viewer sees.
func discardsSeenBy(g *game.Game, viewer string) []LogEvent {
	return logEntriesOfKind(ViewOfGameFor(g, viewer).Log, LogDiscard)
}

// assertDiscards checks that every viewer sees one discard entry per
// card in ids, in order, all in one batch, with the wire fields the cue
// reads.
func assertDiscards(t *testing.T, g *game.Game, seat int, ids []uuid.UUID, cause string) []LogEvent {
	t.Helper()
	var last []LogEvent
	for _, viewer := range allViewers(g) {
		got := discardsSeenBy(g, viewer)
		if len(got) != len(ids) {
			t.Fatalf("viewer %q: %d discard entries, want %d: %+v", viewer, len(got), len(ids), got)
		}
		for i, e := range got {
			if e.Batch == 0 || e.Batch != got[0].Batch {
				t.Errorf("viewer %q: discard %d batch %d, want one shared non-zero batch (%d)", viewer, i, e.Batch, got[0].Batch)
			}
			raw := wireFields(t, e)
			want := map[string]any{
				"kind":     "discard",
				"card_id":  ids[i].String(),
				"old_zone": "hand",
				"new_zone": "graveyard",
				"cause":    cause,
				"seat":     float64(seat),
			}
			for key, w := range want {
				if raw[key] != w {
					t.Errorf("viewer %q: discard %d wire field %q = %v, want %v (entry %v)", viewer, i, key, raw[key], w, raw)
				}
			}
			assertNoUUID(t, e.Text)
		}
		last = got
	}
	return last
}

func TestADiscardIsOneLogLinePerCardForEveryViewer(t *testing.T) {
	t.Run("an effect's discard of two cards", func(t *testing.T) {
		g := keptGame(t, 0)
		p := g.Seats[0]
		ids := handIDs(g, p, 2)
		names := make([]string, 0, 2)
		g.ReadSnapshot(func() {
			for _, c := range p.Hand.Cards[:2] {
				names = append(names, c.Name)
			}
		})
		var err error
		g.WithWriteLock(func() { err = g.DiscardCardsKeyedForEffect(p.ID, ids, uuid.Nil, game.KeyedTail{}) })
		if err != nil {
			t.Fatalf("DiscardCardsKeyedForEffect: %v", err)
		}
		got := assertDiscards(t, g, p.Seat, ids, "effect")
		for i, e := range got {
			if want := "P1 discarded " + names[i]; e.Text != want {
				t.Errorf("text %q, want %q", e.Text, want)
			}
		}
	})

	t.Run("a random discard", func(t *testing.T) {
		g := keptGame(t, 0)
		p := g.Seats[0]
		before := map[uuid.UUID]bool{}
		for _, id := range handIDs(g, p, p.Hand.Size()) {
			before[id] = true
		}
		var err error
		g.WithWriteLock(func() { err = g.DiscardRandomForEffect(p.ID, 2) })
		if err != nil {
			t.Fatalf("DiscardRandomForEffect: %v", err)
		}
		var ids []uuid.UUID
		for _, e := range discardsSeenBy(g, "") {
			id := uuid.MustParse(e.CardID)
			if !before[id] {
				t.Errorf("discarded %s was not in the hand", id)
			}
			ids = append(ids, id)
		}
		assertDiscards(t, g, p.Seat, ids, "effect")
	})

	t.Run("the cleanup step's discard to hand size (CR 514.1)", func(t *testing.T) {
		g := keptGame(t, 12)
		active := g.Seats[g.Turn.ActiveSeat]
		for active.Hand.Size() < 9 {
			if err := g.DrawCard(active.ID); err != nil {
				t.Fatalf("DrawCard: %v", err)
			}
		}
		advanceTo(t, g, game.StepEnd)
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep past End: %v", err)
		}
		if g.Turn.Step != game.StepCleanup || g.DiscardPending[active.ID] != 2 {
			t.Fatalf("expected a cleanup discard of 2, at %s pending %v", g.Turn.Step, g.DiscardPending)
		}
		ids := handIDs(g, active, 2)
		if err := g.DiscardSelection(active.ID, ids); err != nil {
			t.Fatalf("DiscardSelection: %v", err)
		}
		for _, e := range assertDiscards(t, g, active.Seat, ids, "cleanup") {
			if want := " to hand size"; len(e.Text) < len(want) || e.Text[len(e.Text)-len(want):] != want {
				t.Errorf("text %q does not end %q", e.Text, want)
			}
		}
	})

	t.Run("a cycling's discard is its cycle line, which says where the card went", func(t *testing.T) {
		g := buildActiveGame(t)
		actor, other := g.Seats[0], g.Seats[1]
		card := uuid.New()
		g.WithWriteLock(func() {
			actor.Graveyard.PushTop(game.Card{
				InstanceID: card, Name: "Shefet Monitor", TypeLine: "Creature — Naga",
				Owner: actor.ID, Controller: actor.ID,
				KnownBy: map[uuid.UUID]bool{actor.ID: true, other.ID: true},
			})
			g.EmitEvent(game.Event{
				Kind: game.EventDiscardCard, Actor: actor.ID, Source: card, CardID: card,
				OldZone: game.ZoneHand, NewZone: game.ZoneGraveyard, DiscardCause: game.DiscardCauseCost,
			})
			g.EmitEvent(game.Event{Kind: game.EventCycle, Actor: actor.ID, Source: card, CardID: card})
		})
		log := ViewOfGameFor(g, other.ID.String()).Log
		if got := logEntriesOfKind(log, LogDiscard); len(got) != 0 {
			t.Errorf("the cost's discard line survived beside the cycling: %+v", got)
		}
		cycle := findLog(t, log, LogCycle)
		if cycle.Text != "P1 cycled Shefet Monitor" {
			t.Errorf("text %q", cycle.Text)
		}
		raw := wireFields(t, cycle)
		for key, want := range map[string]any{"card_id": card.String(), "old_zone": "hand", "new_zone": "graveyard"} {
			if raw[key] != want {
				t.Errorf("wire field %q = %v, want %v (entry %v)", key, raw[key], want, raw)
			}
		}
	})
}

func TestAMillNamesItsCardsByIDForEveryViewer(t *testing.T) {
	for _, n := range []int{1, 3, 7} {
		g := keptGame(t, 10)
		p := g.Seats[0]
		want := topIDs(g, p, min(n, logRevealNamesMax))
		var err error
		g.WithWriteLock(func() { _, err = g.MillToZoneForEffect(p.ID, n, game.ZoneGraveyard) })
		if err != nil {
			t.Fatalf("MillToZoneForEffect: %v", err)
		}
		for _, viewer := range allViewers(g) {
			mills := logEntriesOfKind(ViewOfGameFor(g, viewer).Log, LogMill)
			if len(mills) != 1 {
				t.Fatalf("mill %d, viewer %q: %d mill entries, want 1: %+v", n, viewer, len(mills), mills)
			}
			e := mills[0]
			if e.Amount != n || e.Seat != p.Seat || e.Batch == 0 {
				t.Errorf("mill %d, viewer %q: amount %d seat %d batch %d", n, viewer, e.Amount, e.Seat, e.Batch)
			}
			raw := wireFields(t, e)
			ids, _ := raw["card_ids"].([]any)
			if len(ids) != len(want) {
				t.Fatalf("mill %d, viewer %q: card_ids %v, want %v", n, viewer, raw["card_ids"], want)
			}
			for i := range want {
				if ids[i] != want[i] {
					t.Errorf("mill %d, viewer %q: card_ids[%d] = %v, want %v", n, viewer, i, ids[i], want[i])
				}
			}
			// None of it is a zone line, so nothing reads it as a death.
			for _, z := range logEntriesOfKind(ViewOfGameFor(g, viewer).Log, LogZone) {
				t.Errorf("mill %d, viewer %q: a zone line as well: %+v", n, viewer, z)
			}
		}
	}
}

func TestAMilledOrDiscardedCommanderGoesHomeFromItsGraveyardAfterwards(t *testing.T) {
	// A commander of seat 0's on top of its library, or in its hand.
	commander := func(g *game.Game, p *game.Player, toHand bool) uuid.UUID {
		id := uuid.New()
		knownBy := map[uuid.UUID]bool{}
		for _, s := range g.Seats {
			knownBy[s.ID] = true
		}
		c := game.Card{
			InstanceID: id, Name: "Wandering Commander", TypeLine: "Legendary Creature — Elf Druid",
			Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID, IsCommander: true, KnownBy: knownBy,
		}
		g.WithWriteLock(func() {
			if toHand {
				p.Hand.PushTop(c)
			} else {
				p.Library.PushTop(c)
			}
		})
		return id
	}
	goHome := func(t *testing.T, g *game.Game, owner *game.Player, id uuid.UUID, kind LogKind) {
		t.Helper()
		g.RunStateChecksForTest()
		var prompt *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceCommanderReturn {
				prompt = c
			}
		}
		if prompt == nil {
			t.Fatalf("no CR 903.9a prompt (pending: %+v)", g.PendingChoices)
		}
		if err := g.ResolveCommanderReturn(prompt.ID, owner.ID, true); err != nil {
			t.Fatalf("ResolveCommanderReturn: %v", err)
		}
		opp := g.Seats[1]
		var went, home *LogEvent
		log := ViewOfGameFor(g, opp.ID.String()).Log
		for i := range log {
			e := &log[i]
			if e.Kind == kind {
				went = e
			}
			if e.Kind == LogZone && e.CardID == id.String() && e.NewZone == string(game.ZoneCommand) {
				home = e
			}
		}
		if went == nil || home == nil {
			t.Fatalf("want a %s entry and then a zone entry home: %+v", kind, log)
		}
		if home.OldZone != string(game.ZoneGraveyard) {
			t.Errorf("the commander went home from %q, want the graveyard", home.OldZone)
		}
		if home.Seq <= went.Seq {
			t.Errorf("home (seq %d) is not after the %s (seq %d)", home.Seq, kind, went.Seq)
		}
	}

	t.Run("milled", func(t *testing.T) {
		g := keptGame(t, 0)
		p := g.Seats[0]
		id := commander(g, p, false)
		var err error
		g.WithWriteLock(func() { _, err = g.MillToZoneForEffect(p.ID, 1, game.ZoneGraveyard) })
		if err != nil {
			t.Fatalf("MillToZoneForEffect: %v", err)
		}
		mills := logEntriesOfKind(ViewOfGameFor(g, "").Log, LogMill)
		if len(mills) != 1 || len(mills[0].CardIDs) != 1 || mills[0].CardIDs[0] != id.String() {
			t.Fatalf("the mill does not name the commander: %+v", mills)
		}
		goHome(t, g, p, id, LogMill)
	})

	t.Run("discarded", func(t *testing.T) {
		g := keptGame(t, 0)
		p := g.Seats[0]
		id := commander(g, p, true)
		var err error
		g.WithWriteLock(func() { err = g.DiscardCardsKeyedForEffect(p.ID, []uuid.UUID{id}, uuid.Nil, game.KeyedTail{}) })
		if err != nil {
			t.Fatalf("DiscardCardsKeyedForEffect: %v", err)
		}
		assertDiscards(t, g, p.Seat, []uuid.UUID{id}, "effect")
		goHome(t, g, p, id, LogDiscard)
	})
}
