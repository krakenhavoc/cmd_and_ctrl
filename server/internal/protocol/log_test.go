package protocol

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// logKinds returns the kinds of every entry, for order assertions.
func logKinds(entries []LogEvent) []LogKind {
	out := make([]LogKind, len(entries))
	for i, e := range entries {
		out[i] = e.Kind
	}
	return out
}

// findLog returns the first entry of the given kind, or fails.
func findLog(t *testing.T, entries []LogEvent, kind LogKind) LogEvent {
	t.Helper()
	for _, e := range entries {
		if e.Kind == kind {
			return e
		}
	}
	t.Fatalf("no %q entry in log: %v", kind, logKinds(entries))
	return LogEvent{}
}

func TestPublicLogProjectsTableEvents(t *testing.T) {
	g := buildActiveGame(t)
	caster := g.Seats[0]
	victim := g.Seats[1]

	boltID := uuid.New()
	bearID := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: bearID,
			Name:       "Grizzly Bears",
			TypeLine:   "Creature — Bear",
			Owner:      caster.ID,
			Controller: caster.ID,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: boltID,
			Name:       "Lightning Bolt",
			TypeLine:   "Instant",
			Owner:      caster.ID,
			Controller: caster.ID,
		})
		g.EmitEvent(game.Event{
			Kind: game.EventStepBegan, Actor: caster.ID, Amount: 3, Round: 2, Label: string(game.StepPrecombatMain),
		})
		g.EmitEvent(game.Event{
			Kind: game.EventCast, Actor: caster.ID, Source: boltID, CardID: boltID,
			OldZone: game.ZoneHand, NewZone: game.ZoneStack,
		})
		g.EmitEvent(game.Event{Kind: game.EventResolve, CardID: boltID})
		g.EmitEvent(game.Event{
			Kind: game.EventDealDamage, Actor: caster.ID, Source: boltID, Target: victim.ID, Amount: 3,
		})
		g.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: victim.ID, Amount: 2})
		g.EmitEvent(game.Event{
			Kind: game.EventAttack, Actor: caster.ID, CardID: bearID, Target: victim.ID,
		})
	})

	log := ViewOfGame(g).Log
	if len(log) == 0 {
		t.Fatal("empty log")
	}

	step := findLog(t, log, LogStep)
	if step.Turn != 3 || step.Round != 2 || step.Step != "precombat_main" {
		t.Errorf("step entry: turn %d step %q, want 3 / precombat_main", step.Turn, step.Step)
	}
	if want := "Turn 2 — P1 · precombat main"; step.Text != want {
		t.Errorf("step text: got %q, want %q", step.Text, want)
	}

	cast := findLog(t, log, LogCast)
	if cast.Turn != 3 {
		t.Errorf("cast entry did not inherit the turn: %+v", cast)
	}
	// `step` rides only on LogStep entries; everything else inherits it
	// from the last one. That is a wire-cost decision, so pin it.
	if cast.Step != "" {
		t.Errorf("non-step entry carries a step string: %+v", cast)
	}
	if want := "P1 cast Lightning Bolt"; cast.Text != want {
		t.Errorf("cast text: got %q, want %q", cast.Text, want)
	}

	if got := findLog(t, log, LogResolve).Text; got != "Lightning Bolt resolved" {
		t.Errorf("resolve text: got %q", got)
	}
	dmg := findLog(t, log, LogDamage)
	if want := "Lightning Bolt dealt 3 damage to P2"; dmg.Text != want {
		t.Errorf("damage text: got %q, want %q", dmg.Text, want)
	}
	if dmg.TargetSeat == nil || *dmg.TargetSeat != victim.Seat {
		t.Errorf("damage target seat: got %v, want seat %d", dmg.TargetSeat, victim.Seat)
	}
	if dmg.Target != "" {
		t.Errorf("damage to a player must not set the card target: %q", dmg.Target)
	}
	if got := findLog(t, log, LogLife).Text; got != "P2 gained 2 life" {
		t.Errorf("life text: got %q", got)
	}
	if got := findLog(t, log, LogAttack).Text; got != "Grizzly Bears attacks P2" {
		t.Errorf("attack text: got %q", got)
	}
}

// TestPublicLogCollapsesDrawsAndNeverNamesThem is the draw half of the
// visibility rule: a drawn card is hidden, so the log says a card was
// drawn and nothing else — no name, no instance ID for anyone,
// including the drawer.
func TestPublicLogCollapsesDrawsAndNeverNamesThem(t *testing.T) {
	g := buildActiveGame(t)
	p := g.Seats[0]
	secret := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: secret,
			Name:       "Demonic Tutor",
			Owner:      p.ID,
			Controller: p.ID,
			KnownBy:    map[uuid.UUID]bool{p.ID: true},
		})
		for range 3 {
			g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: p.ID, CardID: secret})
			g.EmitEvent(game.Event{
				Kind: game.EventZoneMove, Actor: p.ID, CardID: secret,
				OldZone: game.ZoneLibrary, NewZone: game.ZoneHand,
			})
		}
	})

	for _, viewer := range []string{"", p.ID.String(), g.Seats[1].ID.String()} {
		v := ViewOfGameFor(g, viewer)
		draws := 0
		for _, e := range v.Log {
			if e.Kind == LogZone {
				t.Errorf("viewer %q: library→hand move produced a zone entry: %+v", viewer, e)
			}
			if e.Kind != LogDraw {
				continue
			}
			draws++
			if e.CardID != "" || e.cardName != "" {
				t.Errorf("viewer %q: draw entry names a card: %+v", viewer, e)
			}
			if e.Amount != 3 {
				t.Errorf("viewer %q: draws did not collapse: amount %d, want 3", viewer, e.Amount)
			}
			if want := "P1 drew 3 cards"; e.Text != want {
				t.Errorf("viewer %q: draw text %q, want %q", viewer, e.Text, want)
			}
		}
		if draws != 1 {
			t.Errorf("viewer %q: %d draw entries, want 1 collapsed entry", viewer, draws)
		}
		// The strongest form of the assertion: the hidden card's
		// instance ID is nowhere in the serialised log at all.
		buf, err := json.Marshal(v.Log)
		if err != nil {
			t.Fatalf("marshal log: %v", err)
		}
		if strings.Contains(string(buf), secret.String()) {
			t.Errorf("viewer %q: hidden card instance ID leaked into the log wire: %s", viewer, buf)
		}
	}
}

// TestPublicLogRedactsCardsTheViewerDoesNotKnow is the central
// visibility test: the same log entry, filtered for two seats, names
// the card for the knower and not for anyone else.
func TestPublicLogRedactsCardsTheViewerDoesNotKnow(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	other := g.Seats[1]

	faceDown := uuid.New()
	g.WithWriteLock(func() {
		// A morph: on the battlefield (public zone, so the entry
		// exists) but known only to its controller.
		g.Battlefield.PushTop(game.Card{
			InstanceID: faceDown,
			Name:       "Sea Gate Restoration",
			TypeLine:   "Sorcery",
			Owner:      owner.ID,
			Controller: owner.ID,
			FaceDown:   true,
			KnownBy:    map[uuid.UUID]bool{owner.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventZoneMove, Actor: owner.ID, CardID: faceDown,
			OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield,
		})
	})

	ownerLog := ViewOfGameFor(g, owner.ID.String()).Log
	otherLog := ViewOfGameFor(g, other.ID.String()).Log
	adminLog := ViewOfGameFor(g, "").Log

	ownerEntry := findLog(t, ownerLog, LogZone)
	otherEntry := findLog(t, otherLog, LogZone)
	adminEntry := findLog(t, adminLog, LogZone)

	if want := "Sea Gate Restoration entered the battlefield"; ownerEntry.Text != want {
		t.Errorf("owner text: got %q, want %q", ownerEntry.Text, want)
	}
	if want := "a card entered the battlefield"; otherEntry.Text != want {
		t.Errorf("opponent text: got %q, want %q", otherEntry.Text, want)
	}
	if want := "Sea Gate Restoration entered the battlefield"; adminEntry.Text != want {
		t.Errorf("admin text: got %q, want %q", adminEntry.Text, want)
	}
	// The card is on the battlefield, so its instance ID is already on
	// the opponent's wire — withholding it from the log would be
	// theatre. The NAME is the secret, and it is what has to be gone.
	if otherEntry.CardID != faceDown.String() {
		t.Errorf("opponent lost the instance ID: %+v", otherEntry)
	}
	buf, err := json.Marshal(otherLog)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(buf), "Sea Gate Restoration") {
		t.Errorf("card name leaked to a non-knower: %s", buf)
	}
	// knowers must not survive onto the filtered value, or a second
	// filter pass for a DIFFERENT viewer would read a stale set.
	if otherEntry.cardKnowers != nil || ownerEntry.cardKnowers != nil {
		t.Error("knower sets survived the filter")
	}
}

// TestPublicLogRedactionIsIdempotent guards the replay path, which
// re-filters a stored (already-filtered) frame. Re-filtering must
// never widen what an entry says.
func TestPublicLogRedactionIsIdempotent(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	secret := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: secret,
			Name:       "Hidden Horror",
			Owner:      owner.ID,
			Controller: owner.ID,
			KnownBy:    map[uuid.UUID]bool{owner.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventZoneMove, Actor: owner.ID, CardID: secret,
			OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield,
		})
	})
	once := FilterViewFor(ViewOfGame(g), owner.ID.String())
	twice := FilterViewFor(once, g.Seats[1].ID.String())
	entry := findLog(t, twice.Log, LogZone)
	if strings.Contains(entry.Text, "Hidden Horror") {
		t.Errorf("re-filtering for another seat re-revealed the name: %q", entry.Text)
	}
}

// TestPublicLogIsBounded holds the wire budget: however long the game
// runs, the log is a ring of PublicLogMax entries keeping the tail.
func TestPublicLogIsBounded(t *testing.T) {
	g := buildActiveGame(t)
	p := g.Seats[0]
	const emitted = PublicLogMax * 4
	g.WithWriteLock(func() {
		for i := range emitted {
			g.EmitEvent(game.Event{
				Kind: game.EventStepBegan, Actor: p.ID, Amount: i + 1, Label: string(game.StepUpkeep),
			})
		}
	})
	log := ViewOfGame(g).Log
	if len(log) != PublicLogMax {
		t.Fatalf("log length %d, want %d", len(log), PublicLogMax)
	}
	// Oldest-first, and it is the TAIL that survives.
	if log[len(log)-1].Turn != emitted {
		t.Errorf("newest entry turn %d, want %d", log[len(log)-1].Turn, emitted)
	}
	if log[0].Turn != emitted-PublicLogMax+1 {
		t.Errorf("oldest entry turn %d, want %d", log[0].Turn, emitted-PublicLogMax+1)
	}
	for i := 1; i < len(log); i++ {
		if log[i].Seq <= log[i-1].Seq {
			t.Fatalf("log not in seq order at %d: %d after %d", i, log[i].Seq, log[i-1].Seq)
		}
	}
}

// TestPublicLogSuppressesRedundantEntries pins the de-duplication
// rules that keep the log readable: a sacrifice is not also reported
// as a zone move, and a resolved spell is not also reported as going
// to the graveyard.
func TestPublicLogSuppressesRedundantEntries(t *testing.T) {
	g := buildActiveGame(t)
	p := g.Seats[0]
	bearID := uuid.New()
	boltID := uuid.New()
	g.WithWriteLock(func() {
		for _, c := range []game.Card{
			{InstanceID: bearID, Name: "Grizzly Bears", Owner: p.ID, Controller: p.ID},
			{InstanceID: boltID, Name: "Lightning Bolt", Owner: p.ID, Controller: p.ID},
		} {
			g.Battlefield.PushTop(c)
		}
		g.EmitEvent(game.Event{Kind: game.EventSacrifice, Actor: p.ID, CardID: bearID})
		g.EmitEvent(game.Event{
			Kind: game.EventZoneMove, Actor: p.ID, CardID: bearID,
			OldZone: game.ZoneBattlefield, NewZone: game.ZoneGraveyard,
		})
		g.EmitEvent(game.Event{Kind: game.EventResolve, CardID: boltID})
		g.EmitEvent(game.Event{
			Kind: game.EventZoneMove, CardID: boltID,
			OldZone: game.ZoneStack, NewZone: game.ZoneGraveyard,
		})
	})
	log := ViewOfGame(g).Log
	var zones int
	for _, e := range log {
		if e.Kind == LogZone {
			zones++
		}
	}
	if zones != 0 {
		t.Errorf("%d redundant zone entries survived: %v", zones, log)
	}
	if got := findLog(t, log, LogSacrifice).Text; got != "P1 sacrificed Grizzly Bears" {
		t.Errorf("sacrifice text: got %q", got)
	}
}

// TestPublicLogSurvivesSnapshotRestore is the drift-test claim made
// executable. The log is a projection of game.Events, which
// snapshot_drift_test.go classifies `carried` — so the history is
// carried across a deploy without the log itself being state.
func TestPublicLogSurvivesSnapshotRestore(t *testing.T) {
	g := buildActiveGame(t)
	p := g.Seats[0]
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventStepBegan, Actor: p.ID, Amount: 9, Label: string(game.StepEnd),
		})
	})
	before := ViewOfGame(g).Log

	snap := g.CaptureSnapshot()
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	after := ViewOfGame(restored).Log
	if len(before) != len(after) {
		t.Fatalf("log length across restore: %d → %d", len(before), len(after))
	}
	for i := range before {
		if before[i].Text != after[i].Text || before[i].Seq != after[i].Seq {
			t.Fatalf("entry %d drifted across restore: %q (seq %d) → %q (seq %d)",
				i, before[i].Text, before[i].Seq, after[i].Text, after[i].Seq)
		}
	}
}

// TestPublicLogWireCost is the measurement the sub-PR 2 view-size
// budget asked for: how much does a FULL log add to a snapshot frame
// on a saturated four-player board? Reported, not asserted on
// absolute bytes — the assertion is on the share of the frame, which
// is the thing the budget cares about.
func TestPublicLogWireCost(t *testing.T) {
	g := buildFourPlayerBoard(t)
	v := FilterViewFor(ViewOfGame(g), g.Seats[0].ID.String())
	if len(v.Log) != PublicLogMax {
		t.Fatalf("log is not full: %d entries, want %d", len(v.Log), PublicLogMax)
	}
	withLog, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	v.Log = nil
	withoutLog, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	delta := len(withLog) - len(withoutLog)
	perEntry := delta / PublicLogMax
	t.Logf("four-player frame: %d B without the log, %d B with a full %d-entry log (+%d B, +%.1f%%; %d B/entry)",
		len(withoutLog), len(withLog), PublicLogMax, delta,
		100*float64(delta)/float64(len(withoutLog)), perEntry)
	t.Logf("deflated (what permessage-deflate would cost if the hub ever enables it): "+
		"%d B without the log, %d B with it (+%d B)",
		deflatedLen(t, withoutLog), deflatedLen(t, withLog),
		deflatedLen(t, withLog)-deflatedLen(t, withoutLog))

	// The gate is on the two numbers this package controls: what one
	// entry costs, and what the whole log costs. A ratio against the
	// rest of the frame is NOT the gate — the synthetic cards here
	// carry no oracle text, faces, or ability lists, so the frame they
	// produce is smaller than a real one and the ratio would flatter
	// the log.
	if perEntry > 160 {
		t.Errorf("log entry costs %d B on the wire, budget is 160 B", perEntry)
	}
	if delta > 32<<10 {
		t.Errorf("a full log costs %d B on the wire, budget is %d B", delta, 32<<10)
	}
}

// deflatedLen reports the size of b under the same DEFLATE a
// permessage-deflate WebSocket extension would apply.
func deflatedLen(t *testing.T, b []byte) int {
	t.Helper()
	var buf bytes.Buffer
	w, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		t.Fatalf("flate.NewWriter: %v", err)
	}
	if _, err := w.Write(b); err != nil {
		t.Fatalf("deflate: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("deflate close: %v", err)
	}
	return buf.Len()
}

// BenchmarkPublicLogProjection isolates what the log adds to a
// broadcast: the projection runs once per frame (not once per viewer),
// over every event the game has ever emitted.
func BenchmarkPublicLogProjection(b *testing.B) {
	g := buildFourPlayerBoardB(b)
	v := ViewOfGame(g)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = publicLogOf(g, &v)
	}
}

// BenchmarkFilterViewForWithLog is the per-VIEWER half: FilterViewFor
// runs once per connected client on every frame.
func BenchmarkFilterViewForWithLog(b *testing.B) {
	g := buildFourPlayerBoardB(b)
	v := ViewOfGame(g)
	viewer := g.Seats[0].ID.String()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FilterViewFor(v, viewer)
	}
}

// buildFourPlayerBoard is a saturated table: four seats, a wide
// battlefield, full hands and graveyards, and enough game history to
// fill the log ring several times over.
func buildFourPlayerBoard(t *testing.T) *game.Game {
	t.Helper()
	return fourPlayerBoard(t)
}

func buildFourPlayerBoardB(b *testing.B) *game.Game {
	b.Helper()
	return fourPlayerBoard(b)
}

func fourPlayerBoard(tb testing.TB) *game.Game {
	tb.Helper()
	g := game.NewGame()
	for i := range 4 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := range 99 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i+1, j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Player %d", i+1), deck); err != nil {
			tb.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(4, 4))); err != nil {
		tb.Fatalf("Start: %v", err)
	}

	g.WithWriteLock(func() {
		// A wide board: 12 permanents a seat, which is a late-game
		// four-player table.
		perms := make([]uuid.UUID, 0, 48)
		for si, p := range g.Seats {
			for j := range 12 {
				id := uuid.New()
				perms = append(perms, id)
				g.Battlefield.PushTop(game.Card{
					InstanceID: id,
					Name:       fmt.Sprintf("Permanent %d-%d", si+1, j+1),
					TypeLine:   "Creature — Elemental Warrior",
					ManaCost:   "{3}{G}{G}",
					Power:      3,
					Toughness:  4,
					Owner:      p.ID,
					Controller: p.ID,
					KnownBy:    map[uuid.UUID]bool{p.ID: true},
				})
			}
			for j := range 8 {
				p.Graveyard.PushTop(game.Card{
					InstanceID: uuid.New(),
					Name:       fmt.Sprintf("Dead %d-%d", si+1, j+1),
					TypeLine:   "Creature — Zombie",
					Owner:      p.ID,
					Controller: p.ID,
				})
			}
		}
		// History: enough events to overflow the ring, in the mix a
		// real game produces.
		for turn := range 40 {
			active := g.Seats[turn%len(g.Seats)]
			g.EmitEvent(game.Event{
				Kind: game.EventStepBegan, Actor: active.ID, Amount: turn + 1,
				Label: string(game.StepPrecombatMain),
			})
			for i := range 4 {
				card := perms[(turn*4+i)%len(perms)]
				victim := g.Seats[(turn+1)%len(g.Seats)]
				g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: active.ID})
				g.EmitEvent(game.Event{
					Kind: game.EventCast, Actor: active.ID, Source: card, CardID: card,
					OldZone: game.ZoneHand, NewZone: game.ZoneStack,
				})
				g.EmitEvent(game.Event{Kind: game.EventResolve, CardID: card})
				g.EmitEvent(game.Event{
					Kind: game.EventZoneMove, Actor: active.ID, CardID: card,
					OldZone: game.ZoneStack, NewZone: game.ZoneBattlefield,
				})
				// #1257: an ETB trigger resolving, the way the engine
				// emits it — source and label, no CardID — so the
				// budget below prices the label an ability's line now
				// carries on the wire.
				g.EmitEvent(game.Event{
					Kind: game.EventResolve, Actor: active.ID, Source: card,
					Label: fmt.Sprintf("Permanent %d-%d — draw a card", (turn*4+i)%len(perms)/12+1, (turn*4+i)%12+1),
				})
				g.EmitEvent(game.Event{
					Kind: game.EventAttack, Actor: active.ID, CardID: card, Target: victim.ID,
				})
				g.EmitEvent(game.Event{
					Kind: game.EventDealDamage, Actor: active.ID, Source: card,
					Target: victim.ID, Amount: 3, Combat: true,
				})
			}
		}
	})
	return g
}
