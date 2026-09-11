package protocol

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // catalog hooks
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Catalog oracle IDs with the structured specs the enumerator needs:
// a targeted instant, a sacrifice-cost activated ability, a mana
// creature, a modal-ish cantrip. Same set internal/legal's own tests
// use, so the two suites move together.
const (
	oracleLightningBolt     = "4457ed35-7c10-48c8-9776-456485fdf070"
	oracleGoblinBombardment = "edad60c6-80de-4033-af1b-a703ac332983"
	oracleLlanowarElves     = "68954295-54e3-4303-a6bc-fc4547a4e3a3"
	oraclePreordain         = "ac641490-ca14-48d7-8cc4-b69ce984befa"
)

// --- four-player fixture ------------------------------------------

// busyTable builds the frame this PR is actually budgeted against: a
// full four-player Commander board, mid-game, with the active seat in
// its precombat main phase holding priority.
//
// Deliberately the expensive shape rather than a typical one. Every
// seat has a seven-card hand and a developed board; the active seat's
// hand holds the cards that expand worst — a targeted burn spell that
// can point at any of four players or any creature on the table, and
// an activated ability with a sacrifice cost that expands across
// everything the seat controls.
func busyTable(t testing.TB, creaturesPerSeat int) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := range 4 {
		deck := make([]game.Card, 0, 40)
		deck = append(deck, game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil))
		for j := range 39 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Seat %d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 13))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}

	for _, p := range g.Seats {
		p.Hand.Cards = nil
		// Seven-card hand: three lands, a bolt, a cantrip, two
		// creatures. The bolt and the cantrip are the two that
		// expand.
		put(p.Hand, p, basicLand("Mountain", "Mountain"))
		put(p.Hand, p, basicLand("Forest", "Forest"))
		put(p.Hand, p, basicLand("Island", "Island"))
		put(p.Hand, p, game.Card{
			Name: "Lightning Bolt", TypeLine: "Instant",
			ManaCost: "{R}", OracleID: oracleLightningBolt,
		})
		put(p.Hand, p, game.Card{
			Name: "Preordain", TypeLine: "Sorcery",
			ManaCost: "{U}", OracleID: oraclePreordain,
		})
		put(p.Hand, p, beast("Grizzly Bears", "{1}{G}", 2, 2))
		put(p.Hand, p, beast("Hill Giant", "{3}{R}", 3, 3))

		// Board: six untapped lands (so the enumerator's
		// affordability planner says yes to everything), a
		// sacrifice outlet, a mana creature, and the creature
		// count the caller asked for.
		for range 2 {
			put(g.Battlefield, p, basicLand("Mountain", "Mountain"))
			put(g.Battlefield, p, basicLand("Forest", "Forest"))
			put(g.Battlefield, p, basicLand("Island", "Island"))
		}
		put(g.Battlefield, p, game.Card{
			Name: "Goblin Bombardment", TypeLine: "Enchantment",
			ManaCost: "{1}{R}", OracleID: oracleGoblinBombardment,
		})
		put(g.Battlefield, p, game.Card{
			Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
			ManaCost: "{G}", Power: 1, Toughness: 1, OracleID: oracleLlanowarElves,
		})
		for j := range creaturesPerSeat {
			put(g.Battlefield, p, beast(fmt.Sprintf("Beast %d", j+1), "{2}{G}", 3, 3))
		}
	}

	advanceTo(t, g, game.StepPrecombatMain)
	return g
}

func put(z *game.Zone, owner *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner = owner.ID
	c.Controller = owner.ID
	z.PushTop(c)
	return c.InstanceID
}

func basicLand(name, sub string) game.Card {
	return game.Card{Name: name, TypeLine: "Basic Land — " + sub}
}

func beast(name, cost string, p, tgh int) game.Card {
	return game.Card{Name: name, TypeLine: "Creature — Beast", ManaCost: cost, Power: p, Toughness: tgh}
}

func advanceTo(t testing.TB, g *game.Game, step game.Step) {
	t.Helper()
	for range 40 {
		if g.Turn.Step == step {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to %s: %v", step, err)
		}
	}
	t.Fatalf("never reached step %s (at %s)", step, g.Turn.Step)
}

// --- own seat only -------------------------------------------------

// TestLegalMovesOwnSeatOnly is the hidden-information gate. A move
// list names the cards in a hand ("Cast Lightning Bolt targeting
// Kess"); handing one seat another seat's list would leak more than
// any other field on the wire.
func TestLegalMovesOwnSeatOnly(t *testing.T) {
	g := busyTable(t, 3)
	active := g.Seats[g.Turn.ActiveSeat]

	for _, p := range g.Seats {
		v := ViewOfGameFor(g, p.ID.String())
		for _, m := range v.LegalMoves {
			if m.Player != p.ID {
				t.Errorf("seat %s received a move for %s: %q", p.Name, m.Player, m.Label)
			}
		}
		if p.ID == active.ID && len(v.LegalMoves) == 0 {
			t.Error("active seat holding priority in its main phase got no moves")
		}
		if p.ID != active.ID && len(v.LegalMoves) != 0 {
			t.Errorf("seat %s holds no priority and owes no choice, got %d moves: %v",
				p.Name, len(v.LegalMoves), labelsOf(v.LegalMoves))
		}
	}
}

// TestLegalMovesNeverLeakThroughTheRawView locks down the two paths
// that bypass FilterViewFor entirely: the crash dump and the replay
// log both marshal the UNFILTERED GameView. The per-seat enumeration
// lives in an unexported field precisely so those paths carry no
// seat's moves at all.
func TestLegalMovesNeverLeakThroughTheRawView(t *testing.T) {
	g := busyTable(t, 3)
	raw := ViewOfGame(g)
	if len(raw.LegalMoves) != 0 {
		t.Errorf("unfiltered view carries %d moves; it must carry none", len(raw.LegalMoves))
	}
	b, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), "legal_moves") {
		t.Error("unfiltered view marshalled a legal_moves key")
	}
	if strings.Contains(string(b), "Lightning Bolt targeting") {
		t.Error("unfiltered view marshalled an enumerated move label")
	}
}

// TestLegalMovesSpectatorGetsNone: an unseated viewer has no "own
// seat", so there is no list they are entitled to. This is the one
// place the "empty viewer ID sees everything" convention deliberately
// does not apply.
func TestLegalMovesSpectatorGetsNone(t *testing.T) {
	g := busyTable(t, 3)
	if v := ViewOfGameFor(g, ""); len(v.LegalMoves) != 0 {
		t.Errorf("spectator got %d moves", len(v.LegalMoves))
	}
	if v := ViewOfGameFor(g, uuid.New().String()); len(v.LegalMoves) != 0 {
		t.Errorf("unknown viewer got %d moves", len(v.LegalMoves))
	}
}

// TestFilterViewForIsIdempotentOnLegalMoves: re-filtering an
// already-filtered view must not promote one seat's list onto
// another's frame. The map is unexported and FilterViewFor builds a
// fresh GameView literal, so the second pass sees no map at all.
func TestFilterViewForIsIdempotentOnLegalMoves(t *testing.T) {
	g := busyTable(t, 3)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%4]

	once := FilterViewFor(ViewOfGame(g), active.ID.String())
	if len(once.LegalMoves) == 0 {
		t.Fatal("active seat got no moves on the first filter")
	}
	twice := FilterViewFor(once, other.ID.String())
	if len(twice.LegalMoves) != 0 {
		t.Errorf("re-filtering for another seat produced %d moves", len(twice.LegalMoves))
	}
}

// TestLegalMovesFollowThePriorityCursor: the field is populated when
// the seat owes a decision and empty otherwise, which is what keeps
// the median frame free of it.
func TestLegalMovesFollowThePriorityCursor(t *testing.T) {
	g := busyTable(t, 3)
	active := g.Seats[g.Turn.ActiveSeat]
	defender := g.Seats[(g.Turn.ActiveSeat+1)%4]

	// Declare-blockers with an attacker pointed at the defender: the
	// defender owes a block declaration while the ACTIVE seat still
	// holds priority (#328), so both lists must be live.
	advanceTo(t, g, game.StepDeclareAttackers)
	var attacker uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == active.ID && c.IsCreature() && !c.Tapped && !game.HasSummoningSickness(c) {
			attacker = c.InstanceID
			break
		}
	}
	if attacker == uuid.Nil {
		t.Skip("no legal attacker in the fixture; combat shape covered by internal/legal")
	}
	if err := g.DeclareAttacker(attacker, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)

	dv := ViewOfGameFor(g, defender.ID.String())
	if !anyKind(dv.LegalMoves, "block") {
		t.Errorf("defender under attack got no block moves: %v", labelsOf(dv.LegalMoves))
	}
	for _, m := range dv.LegalMoves {
		if m.Player != defender.ID {
			t.Errorf("defender received a move for %s", m.Player)
		}
	}
}

func anyKind(moves []LegalMoveView, kind string) bool {
	for _, m := range moves {
		if string(m.Kind) == kind {
			return true
		}
	}
	return false
}

func labelsOf(moves []LegalMoveView) []string {
	out := make([]string, 0, len(moves))
	for _, m := range moves {
		out = append(out, m.Label)
	}
	return out
}

// --- frame budget --------------------------------------------------

// legalMovesFrameBudget is the ceiling this PR commits to for the
// legal_moves field alone, on the worst realistic frame: a full
// four-player board, the active seat holding priority in its main
// phase with a hand that expands.
//
// The number is a SHARE of a frame, not an absolute: S31 sub-PR 0 is
// adding a bounded public game log to the same GameView concurrently,
// and two agents each assuming the other's headroom is how a frame
// budget gets blown. 24 KiB is roughly a quarter of the ~100 KiB
// four-player snapshot the rest of the view already costs, and leaves
// the log the same order of room.
//
// If this test fails, the answer is NOT to raise the constant. It is
// to lower legal.Options.MaxExpansionPerSource — the cap exists for
// exactly this — or to stop enumerating a move class the client does
// not consume.
const legalMovesFrameBudget = 24 * 1024

// TestLegalMovesFrameBudget measures the wire cost of legal_moves on
// the worst four-player frame and fails if it exceeds the budget.
// The measured numbers are logged on every run (go test -v) so the
// next person changing the enumerator sees the trend rather than
// rediscovering it.
func TestLegalMovesFrameBudget(t *testing.T) {
	for _, creatures := range []int{3, 6, 10} {
		t.Run(fmt.Sprintf("creatures=%d", creatures), func(t *testing.T) {
			g := busyTable(t, creatures)
			active := g.Seats[g.Turn.ActiveSeat]

			withMoves := ViewOfGameFor(g, active.ID.String())
			withoutMoves := withMoves
			withoutMoves.LegalMoves = nil

			full, err := json.Marshal(withMoves)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			base, err := json.Marshal(withoutMoves)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			growth := len(full) - len(base)

			t.Logf("board: 4 seats x %d beasts | frame %d B -> %d B | legal_moves %d B (+%.1f%%) across %d moves",
				creatures, len(base), len(full), growth,
				100*float64(growth)/float64(len(base)), len(withMoves.LegalMoves))

			if growth > legalMovesFrameBudget {
				t.Errorf("legal_moves cost %d B, over the %d B budget — lower MaxExpansionPerSource rather than raising this",
					growth, legalMovesFrameBudget)
			}
		})
	}
}

// TestLegalMovesPathologicalBoardStaysCapped is the case the
// per-source cap alone does not cover: the enumerator bounds ONE
// card's expansion at 12, so a seat with several crossed-product
// sources multiplies straight past any per-source ceiling. Four
// sacrifice outlets and four burn spells is ~100 moves and ~40 KB
// unbounded. The wire cap degrades it instead, and the invariant the
// client depends on — every playable card keeps at least one move —
// has to survive that.
func TestLegalMovesPathologicalBoardStaysCapped(t *testing.T) {
	g := busyTable(t, 4)
	active := g.Seats[g.Turn.ActiveSeat]

	// Three more of each crossed-product source, on top of what
	// busyTable already gave every seat.
	handSources := []uuid.UUID{}
	for range 3 {
		put(g.Battlefield, active, game.Card{
			Name: "Goblin Bombardment", TypeLine: "Enchantment",
			ManaCost: "{1}{R}", OracleID: oracleGoblinBombardment,
		})
		handSources = append(handSources, put(active.Hand, active, game.Card{
			Name: "Lightning Bolt", TypeLine: "Instant",
			ManaCost: "{R}", OracleID: oracleLightningBolt,
		}))
	}

	v := ViewOfGameFor(g, active.ID.String())
	if len(v.LegalMoves) > legalMovesWireCap {
		t.Errorf("capped list is %d moves, over the %d cap", len(v.LegalMoves), legalMovesWireCap)
	}

	full, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	bare := v
	bare.LegalMoves = nil
	base, err := json.Marshal(bare)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	growth := len(full) - len(base)
	t.Logf("pathological board (4 outlets + 4 burn spells): %d moves, legal_moves %d B", len(v.LegalMoves), growth)
	if growth > legalMovesFrameBudget {
		t.Errorf("legal_moves cost %d B, over the %d B budget", growth, legalMovesFrameBudget)
	}

	// The degradation must not cost a card its representation.
	sources := map[uuid.UUID]bool{}
	for _, m := range v.LegalMoves {
		sources[m.Source] = true
	}
	for _, id := range handSources {
		if !sources[id] {
			t.Errorf("degraded list dropped every move for hand card %s — that greys a playable card", id)
		}
	}
}

// BenchmarkViewOfGame measures the projection on the same busy
// four-player board, with and without the enumeration.
//
// The cost matters because ViewOfGame runs on EVERY Room.Apply and
// every Room.Snapshot — it is the hot path, not a debug surface — and
// it now enumerates once per seat rather than once. The seats that
// owe nothing are nearly free (the enumerator returns on its first
// priority check); the priority holder's pass is the real cost.
func BenchmarkViewOfGame(b *testing.B) {
	g := busyTable(b, 6)
	b.Run("with_legal_moves", func(b *testing.B) {
		for range b.N {
			_ = ViewOfGame(g)
		}
	})
	b.Run("enumeration_only", func(b *testing.B) {
		for range b.N {
			g.ReadSnapshot(func() { _ = enumerateLegalMoves(g) })
		}
	})
}

// TestLegalMovesQuietFrameIsFree is the other half of the budget
// story: on a frame where the viewer owes no decision — which is most
// of them, including every frame an opponent receives during your
// turn — the field costs nothing at all.
func TestLegalMovesQuietFrameIsFree(t *testing.T) {
	g := busyTable(t, 6)
	idle := g.Seats[(g.Turn.ActiveSeat+1)%4]

	v := ViewOfGameFor(g, idle.ID.String())
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), "legal_moves") {
		t.Error("an idle seat's frame carries a legal_moves key; omitempty should have dropped it")
	}
	t.Logf("idle four-player frame: %d B, 0 B of legal_moves", len(b))
}
