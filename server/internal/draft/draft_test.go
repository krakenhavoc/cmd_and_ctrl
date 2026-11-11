package draft

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draftpacks"
)

var testNS = uuid.MustParse("00000000-0000-0000-0000-000000005438")

func synthID(parts ...any) uuid.UUID {
	return uuid.NewSHA1(testNS, []byte(fmt.Sprint(parts...)))
}

// synthIndex is a made-up set about the size of a real one, released in
// the play-booster era (14 cards a pack).
func synthIndex(set, released string) *cards.Index {
	idx := cards.NewIndex()
	n := 1
	colors := []string{"W", "U", "B", "R", "G"}
	put := func(prefix, rarity string, count int) {
		for i := range count {
			name := fmt.Sprintf("%s %s %02d", set, prefix, i)
			idx.Put(cards.Card{
				ID: synthID("print", set, n), OracleID: synthID("oracle", name), Name: name,
				SetCode: set, SetName: "Synthetic " + set, SetType: "expansion",
				CollectorNo: fmt.Sprint(n), Lang: "en", Layout: "normal",
				TypeLine: "Creature — Test", BorderColor: "black", ReleasedAt: released,
				Rarity: rarity, Booster: true,
				Colors: []string{colors[i%5]}, ManaCost: fmt.Sprintf("{%d}{%s}", i%5, colors[i%5]),
			})
			n++
		}
	}
	put("Common", "common", 80)
	put("Uncommon", "uncommon", 100)
	put("Rare", "rare", 60)
	put("Mythic", "mythic", 20)
	for _, b := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest"} {
		idx.Put(cards.Card{
			ID: synthID("print", set, n), OracleID: synthID("oracle", b), Name: b,
			SetCode: set, SetName: "Synthetic " + set, SetType: "expansion",
			CollectorNo: fmt.Sprint(n), Lang: "en", Layout: "normal",
			TypeLine: "Basic Land — " + b, BorderColor: "black", ReleasedAt: released,
			Rarity: "common", Booster: true,
		})
		n++
	}
	return idx
}

func allFull(uuid.UUID) bool { return true }

func testSource(t *testing.T) *SetSource {
	t.Helper()
	src, err := NewSetSource(synthIndex("syn", "2026-03-01"), "syn", allFull)
	if err != nil {
		t.Fatalf("NewSetSource: %v", err)
	}
	return src
}

var t0 = time.Date(2026, 11, 11, 19, 0, 0, 0, time.UTC)

// newDraft makes a draft with `humans` people seated; the first is the
// host.
func newDraft(t *testing.T, mode Mode, seats, humans int, seed uint64) (*Draft, []uuid.UUID) {
	t.Helper()
	d, err := New(Config{Name: "Test draft", Set: "syn", SetName: "Synthetic syn", Mode: mode, Seats: seats, Seed: seed}, t0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	var ids []uuid.UUID
	for i := range humans {
		s, err := d.AddSeat(fmt.Sprintf("Person %d", i+1), uuid.Nil)
		if err != nil {
			t.Fatalf("AddSeat: %v", err)
		}
		ids = append(ids, s.ID)
	}
	return d, ids
}

// firstCards names the first n cards of a seat's pack.
func firstCards(d *Draft, seat uuid.UUID) []string {
	pack := d.HeldPack(seat)
	n := d.Take(seat)
	out := make([]string, 0, n)
	for _, c := range pack[:n] {
		out = append(out, c.ID)
	}
	return out
}

func cardIDs(cs []Card) []string {
	out := make([]string, 0, len(cs))
	for _, c := range cs {
		out = append(out, c.ID)
	}
	return out
}

// A full eight-seat pick-one draft, every seat a person, from seeded
// packs: every card of every pack ends in exactly one pool, and each
// round passes the way CR 905.1b says.
func TestEightSeatPickOneDraft(t *testing.T) {
	src := testSource(t)
	d, seats := newDraft(t, PickOne, 8, 8, 42)
	if err := d.Start(seats[0], src, t0); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if d.State() != StatePicking || d.Round() != 1 || d.PickNumber() != 1 {
		t.Fatalf("after Start: state %s round %d pick %d", d.State(), d.Round(), d.PickNumber())
	}

	// Every card dealt, by ID, from the first view of each round.
	dealt := map[string]bool{}
	wantDirection := map[int]Direction{1: Left, 2: Right, 3: Left}
	now := t0
	for d.State() == StatePicking {
		round, pick := d.Round(), d.PickNumber()
		if got := DirectionOf(round); got != wantDirection[round] {
			t.Fatalf("round %d passes %s, want %s", round, got, wantDirection[round])
		}
		// What each seat holds, and what it will leave behind.
		before := make([][]string, len(seats))
		for i, id := range seats {
			pack := d.HeldPack(id)
			if len(pack) != 14-(pick-1) {
				t.Fatalf("round %d pick %d seat %d holds %d cards, want %d", round, pick, i, len(pack), 14-(pick-1))
			}
			if pick == 1 {
				for _, c := range pack {
					if dealt[c.ID] {
						t.Fatalf("card %s was dealt twice", c.ID)
					}
					dealt[c.ID] = true
				}
			}
			before[i] = cardIDs(pack)
		}
		for _, id := range seats {
			if err := d.Pick(id, firstCards(d, id), now); err != nil {
				t.Fatalf("round %d pick %d: %v", round, pick, err)
			}
		}
		now = now.Add(time.Second)
		if d.State() != StatePicking || d.Round() != round {
			continue // the round ended; nothing was passed
		}
		// Seat i's leftovers are now held by its left neighbour (i+1) in
		// rounds one and three, by its right neighbour (i-1) in round two.
		for i := range seats {
			to := (i + 1) % 8
			if wantDirection[round] == Right {
				to = (i + 7) % 8
			}
			got := cardIDs(d.HeldPack(seats[to]))
			if want := before[i][1:]; !reflect.DeepEqual(got, want) {
				t.Fatalf("round %d pick %d: seat %d's pack went wrong: seat %d holds %v, want %v", round, pick, i, to, got, want)
			}
		}
	}

	if d.State() != StateDone {
		t.Fatalf("state %s, want done", d.State())
	}
	if len(dealt) != 8*3*14 {
		t.Fatalf("%d cards were dealt, want %d", len(dealt), 8*3*14)
	}
	inPools := map[string]int{}
	for i, pool := range d.Pools() {
		if len(pool) != 3*14 {
			t.Fatalf("seat %d's pool has %d cards, want %d", i, len(pool), 3*14)
		}
		for _, c := range pool {
			inPools[c.ID]++
		}
	}
	for id := range dealt {
		if inPools[id] != 1 {
			t.Fatalf("card %s is in %d pools, want exactly 1", id, inPools[id])
		}
	}
	if len(inPools) != len(dealt) {
		t.Fatalf("pools hold %d different cards, %d were dealt", len(inPools), len(dealt))
	}
	for _, id := range seats {
		if got := d.Pool(id); len(got) != 42 {
			t.Fatalf("Pool: %d cards", len(got))
		}
	}
}

// The same seed, source and picks give the same draft.
func TestASeedReplaysADraft(t *testing.T) {
	run := func(seed uint64) [][]Card {
		d, seats := newDraft(t, PickOne, 8, 1, seed)
		if err := d.Start(seats[0], testSource(t), t0); err != nil {
			t.Fatalf("Start: %v", err)
		}
		for d.State() == StatePicking {
			if err := d.Pick(seats[0], firstCards(d, seats[0]), t0); err != nil {
				t.Fatalf("Pick: %v", err)
			}
		}
		return d.Pools()
	}
	names := func(pools [][]Card) [][]string {
		out := make([][]string, len(pools))
		for i, p := range pools {
			for _, c := range p {
				out[i] = append(out[i], c.ID+" "+c.Name)
			}
		}
		return out
	}
	a, b, c := names(run(7)), names(run(7)), names(run(8))
	if !reflect.DeepEqual(a, b) {
		t.Fatal("two drafts with one seed and the same picks differ")
	}
	if reflect.DeepEqual(a, c) {
		t.Fatal("two seeds gave the same draft")
	}
}

// A four-seat pick-two draft: each pick takes two cards, so a 14-card
// pack is seven picks, and every card still ends in exactly one pool.
func TestFourSeatPickTwoDraft(t *testing.T) {
	d, seats := newDraft(t, PickTwo, 4, 4, 99)
	if err := d.Start(seats[0], testSource(t), t0); err != nil {
		t.Fatalf("Start: %v", err)
	}
	v, _ := d.ViewFor(seats[0], t0, nil)
	if v.Take != 2 || v.PicksInRound != 7 {
		t.Fatalf("take %d, picks in round %d; want 2 and 7", v.Take, v.PicksInRound)
	}
	// One card is not a pick two pick, nor are three, nor one card twice.
	pack := d.HeldPack(seats[0])
	for _, bad := range [][]string{
		{pack[0].ID},
		{pack[0].ID, pack[1].ID, pack[2].ID},
		{pack[0].ID, pack[0].ID},
		{pack[0].ID, "c0"},
	} {
		if err := d.Pick(seats[0], bad, t0); !errors.Is(err, ErrBadPick) {
			t.Fatalf("Pick(%v) = %v, want ErrBadPick", bad, err)
		}
	}
	picks := 0
	for d.State() == StatePicking {
		for _, id := range seats {
			if got := d.Take(id); got != 2 {
				t.Fatalf("Take = %d, want 2", got)
			}
			if err := d.Pick(id, firstCards(d, id), t0); err != nil {
				t.Fatalf("Pick: %v", err)
			}
		}
		picks++
	}
	if picks != Rounds*7 {
		t.Fatalf("%d picks, want %d", picks, Rounds*7)
	}
	seen := map[string]int{}
	for i, pool := range d.Pools() {
		if len(pool) != 42 {
			t.Fatalf("seat %d's pool has %d cards, want 42", i, len(pool))
		}
		for _, c := range pool {
			seen[c.ID]++
		}
	}
	if len(seen) != 4*3*14 {
		t.Fatalf("pools hold %d different cards, want %d", len(seen), 4*3*14)
	}
	for id, n := range seen {
		if n != 1 {
			t.Fatalf("card %s is in %d pools", id, n)
		}
	}
}

// Pick two is allowed at any size, and a 15-card pack ends on a pick of
// one card.
func TestPickTwoTakesOneFromAnOddPack(t *testing.T) {
	// A set from the draft-booster era: 15 cards a pack.
	src, err := NewSetSource(synthIndex("old", "2015-03-01"), "old", allFull)
	if err != nil {
		t.Fatalf("NewSetSource: %v", err)
	}
	d, seats := newDraft(t, PickTwo, 8, 1, 3)
	if err := d.Start(seats[0], src, t0); err != nil {
		t.Fatalf("Start: %v", err)
	}
	v, _ := d.ViewFor(seats[0], t0, nil)
	if len(v.Pack) != 15 || v.PicksInRound != 8 {
		t.Fatalf("pack of %d, %d picks in the round; want 15 and 8", len(v.Pack), v.PicksInRound)
	}
	var takes []int
	for d.State() == StatePicking && d.Round() == 1 {
		takes = append(takes, d.Take(seats[0]))
		if err := d.Pick(seats[0], firstCards(d, seats[0]), t0); err != nil {
			t.Fatalf("Pick: %v", err)
		}
	}
	if want := []int{2, 2, 2, 2, 2, 2, 2, 1}; !reflect.DeepEqual(takes, want) {
		t.Fatalf("takes %v, want %v", takes, want)
	}
}

// One person and seven bots is a normal draft: it completes with no
// input but that person's, and the bots draft a full pool each.
func TestOnePersonAndSevenBots(t *testing.T) {
	d, seats := newDraft(t, PickOne, 8, 1, 5)
	if err := d.Start(seats[0], testSource(t), t0); err != nil {
		t.Fatalf("Start: %v", err)
	}
	bots := 0
	for _, s := range d.Seats() {
		if s.Bot {
			bots++
		}
	}
	if bots != 7 {
		t.Fatalf("%d bots, want 7", bots)
	}
	mine := 0
	for d.State() == StatePicking {
		pending := d.Pending()
		if len(pending) != 1 || pending[0].ID != seats[0] {
			t.Fatalf("pending %d seats; only the person should ever be waited on", len(pending))
		}
		if err := d.Pick(seats[0], firstCards(d, seats[0]), t0); err != nil {
			t.Fatalf("Pick: %v", err)
		}
		mine++
	}
	if mine != 42 {
		t.Fatalf("the person made %d picks, want 42", mine)
	}
	for i, pool := range d.Pools() {
		if len(pool) != 42 {
			t.Fatalf("seat %d's pool has %d cards, want 42", i, len(pool))
		}
	}
}

func TestOnlyTheHostStartsAndOnlyOnce(t *testing.T) {
	d, seats := newDraft(t, PickOne, 4, 2, 1)
	src := testSource(t)
	if err := d.Start(seats[1], src, t0); !errors.Is(err, ErrNotHost) {
		t.Fatalf("Start by a guest seat = %v, want ErrNotHost", err)
	}
	if err := d.Start(seats[0], src, t0); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := d.Start(seats[0], src, t0); !errors.Is(err, ErrStarted) {
		t.Fatalf("second Start = %v, want ErrStarted", err)
	}
	if _, err := d.AddSeat("Late", uuid.Nil); !errors.Is(err, ErrStarted) {
		t.Fatalf("AddSeat after the start = %v, want ErrStarted", err)
	}
}

func TestBadSettingsAreRefused(t *testing.T) {
	for _, cfg := range []Config{
		{Mode: PickOne, Seats: 1},
		{Mode: PickOne, Seats: 9},
		{Mode: "pick_three", Seats: 8},
		{Mode: PickOne, Seats: 8, PickTimeout: -time.Second},
	} {
		if _, err := New(cfg, t0); !errors.Is(err, ErrBadConfig) {
			t.Fatalf("New(%+v) = %v, want ErrBadConfig", cfg, err)
		}
	}
	d, _ := newDraft(t, PickOne, 2, 2, 1)
	if _, err := d.AddSeat("Third", uuid.Nil); !errors.Is(err, ErrFull) {
		t.Fatalf("a third seat in a draft of two = %v, want ErrFull", err)
	}
}

// A seat cannot pick twice from one pack, nor name a card it does not
// hold (another seat's pack is not its to pick from).
func TestAPickMustComeFromTheSeatsOwnPack(t *testing.T) {
	d, seats := newDraft(t, PickOne, 2, 2, 11)
	if err := d.Start(seats[0], testSource(t), t0); err != nil {
		t.Fatalf("Start: %v", err)
	}
	theirs := d.HeldPack(seats[1])[0].ID
	if err := d.Pick(seats[0], []string{theirs}, t0); !errors.Is(err, ErrBadPick) {
		t.Fatalf("picking from another seat's pack = %v, want ErrBadPick", err)
	}
	if err := d.Pick(seats[0], firstCards(d, seats[0]), t0); err != nil {
		t.Fatalf("Pick: %v", err)
	}
	if err := d.Pick(seats[0], []string{"c1"}, t0); !errors.Is(err, ErrAlreadyPick) {
		t.Fatalf("a second pick from one pack = %v, want ErrAlreadyPick", err)
	}
	if err := d.Pick(uuid.New(), []string{"c1"}, t0); !errors.Is(err, ErrNoSuchSeat) {
		t.Fatalf("a pick by no seat = %v, want ErrNoSuchSeat", err)
	}
}

// The timer picks for a stalled seat: the placeholder's choice from the
// pack that seat holds, and the draft moves on.
func TestExpirePicksForAStalledSeat(t *testing.T) {
	d, seats := newDraft(t, PickOne, 4, 2, 21)
	if err := d.Start(seats[0], testSource(t), t0); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if got := d.Deadline(); !got.Equal(t0.Add(DefaultPickTimeout)) {
		t.Fatalf("deadline %s, want %s", got, t0.Add(DefaultPickTimeout))
	}
	if err := d.Pick(seats[0], firstCards(d, seats[0]), t0); err != nil {
		t.Fatalf("Pick: %v", err)
	}
	held := cardIDs(d.HeldPack(seats[1]))
	later := t0.Add(DefaultPickTimeout)
	d.Expire(later)
	pool := d.Pool(seats[1])
	if len(pool) != 1 {
		t.Fatalf("the stalled seat has %d picks, want 1", len(pool))
	}
	found := false
	for _, id := range held {
		found = found || id == pool[0].ID
	}
	if !found {
		t.Fatalf("the timer's pick %s is not from the pack the seat held", pool[0].ID)
	}
	if d.PickNumber() != 2 {
		t.Fatalf("pick %d after the timer, want 2", d.PickNumber())
	}
	if got := d.Deadline(); !got.Equal(later.Add(DefaultPickTimeout)) {
		t.Fatalf("the next pick's deadline is %s, want %s", got, later.Add(DefaultPickTimeout))
	}
}

// A person who leaves during the picks becomes a bot: the draft goes on
// and their seat still drafts a full pool.
func TestLeavingDuringThePicksHandsTheSeatToABot(t *testing.T) {
	d, seats := newDraft(t, PickOne, 4, 2, 31)
	if err := d.Start(seats[0], testSource(t), t0); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := d.Leave(seats[1], t0); err != nil {
		t.Fatalf("Leave: %v", err)
	}
	for d.State() == StatePicking {
		if err := d.Pick(seats[0], firstCards(d, seats[0]), t0); err != nil {
			t.Fatalf("Pick: %v", err)
		}
	}
	if got := len(d.Pool(seats[1])); got != 42 {
		t.Fatalf("the seat that was left drafted %d cards, want 42", got)
	}
	if d.Host() != seats[0] {
		t.Fatal("the host changed")
	}
}

func TestLeavingTheLobbyFreesTheSeatAndPassesTheHost(t *testing.T) {
	d, seats := newDraft(t, PickOne, 4, 3, 1)
	if err := d.Leave(seats[0], t0); err != nil {
		t.Fatalf("Leave: %v", err)
	}
	if d.Host() != seats[1] {
		t.Fatal("the next person should host")
	}
	if d.Humans() != 2 || len(d.Seats()) != 2 {
		t.Fatalf("%d people in %d seats, want 2 in 2", d.Humans(), len(d.Seats()))
	}
	if err := d.Leave(seats[0], t0); !errors.Is(err, ErrNoSuchSeat) {
		t.Fatalf("leaving twice = %v, want ErrNoSuchSeat", err)
	}
}

// A Picker that answers with nonsense is overruled; the draft never
// stalls on a bot.
type brokenPicker struct{}

func (brokenPicker) Pick(PickRequest) []string { return []string{"not a card"} }

func TestABrokenPickerIsOverruled(t *testing.T) {
	d, err := New(Config{Mode: PickTwo, Seats: 4, Seed: 1, Picker: brokenPicker{}}, t0)
	if err != nil {
		t.Fatal(err)
	}
	host, _ := d.AddSeat("Host", uuid.Nil)
	if err := d.Start(host.ID, testSource(t), t0); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for d.State() == StatePicking {
		d.Expire(t0)
	}
	for i, pool := range d.Pools() {
		if len(pool) != 42 {
			t.Fatalf("seat %d's pool has %d cards, want 42", i, len(pool))
		}
	}
}

// The Picker is asked with the pack the seat holds, its picks so far
// and the pick mode, and nothing of any other seat.
type recordingPicker struct {
	inner Picker
	reqs  []PickRequest
}

func (p *recordingPicker) Pick(req PickRequest) []string {
	p.reqs = append(p.reqs, req)
	return p.inner.Pick(req)
}

func TestThePickerSeesOnlyItsOwnSeat(t *testing.T) {
	rec := &recordingPicker{inner: RandomPicker{Seed: 9}}
	d, err := New(Config{Mode: PickTwo, Seats: 4, Seed: 9, Picker: rec}, t0)
	if err != nil {
		t.Fatal(err)
	}
	host, _ := d.AddSeat("Host", uuid.Nil)
	if err := d.Start(host.ID, testSource(t), t0); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for d.State() == StatePicking {
		d.Expire(t0)
	}
	if len(rec.reqs) != 4*3*7 {
		t.Fatalf("the picker was asked %d times, want %d", len(rec.reqs), 4*3*7)
	}
	picked := map[int]int{}
	for _, req := range rec.reqs {
		if req.Mode != PickTwo || req.Take != 2 || req.Seats != 4 {
			t.Fatalf("request %+v: want pick two, take 2, 4 seats", req)
		}
		if len(req.Picked) != picked[req.Seat] {
			t.Fatalf("seat %d was shown %d picks, it has made %d", req.Seat, len(req.Picked), picked[req.Seat])
		}
		if len(req.Pack) != 14-2*(req.Pick-1) {
			t.Fatalf("round %d pick %d: a pack of %d", req.Round, req.Pick, len(req.Pack))
		}
		picked[req.Seat] += req.Take
	}
}

func TestRandomPickerIsAFunctionOfTheSeedAndThePick(t *testing.T) {
	pack := make([]Card, 14)
	for i := range pack {
		pack[i].ID = fmt.Sprint("c", i+1)
	}
	req := PickRequest{Pack: pack, Mode: PickTwo, Take: 2, Seat: 3, Seats: 8, Round: 2, Pick: 5}
	a, b := RandomPicker{Seed: 1}.Pick(req), RandomPicker{Seed: 1}.Pick(req)
	if !reflect.DeepEqual(a, b) || len(a) != 2 || a[0] == a[1] {
		t.Fatalf("picks %v and %v: want the same two different cards", a, b)
	}
	// Over many seats and picks, it does not always take the same slot.
	first := map[string]bool{}
	for seat := range 8 {
		for pick := 1; pick <= 14; pick++ {
			req.Seat, req.Pick = seat, pick
			first[RandomPicker{Seed: 1}.Pick(req)[0]] = true
		}
	}
	if len(first) < 8 {
		t.Fatalf("the random picker took only %d different cards over 112 picks", len(first))
	}
}

// --- hidden information ---------------------------------------------------

// cardFacts is everything in a marshalled view that could identify a
// card: every string and every number under a key that names one.
func collectStrings(v any, out map[string]bool) {
	switch x := v.(type) {
	case map[string]any:
		for _, e := range x {
			collectStrings(e, out)
		}
	case []any:
		for _, e := range x {
			collectStrings(e, out)
		}
	case string:
		out[x] = true
	}
}

// TestNoSeatSeesAnotherSeatsCards walks every view every seat is given
// through a whole draft, as the JSON that goes on the wire, and fails
// on any card the seat may not see (CR 905.1c): a card is allowed only
// if it is in the pack the seat holds or among the seat's own picks.
// It checks by card ID, Scryfall ID and name, anywhere in the message,
// so a new field that carried a card would be caught too.
func TestNoSeatSeesAnotherSeatsCards(t *testing.T) {
	for _, mode := range []Mode{PickOne, PickTwo} {
		t.Run(string(mode), func(t *testing.T) {
			d, seats := newDraft(t, mode, 4, 4, 77)
			if err := d.Start(seats[0], testSource(t), t0); err != nil {
				t.Fatalf("Start: %v", err)
			}
			// The server's own reading of who holds what.
			type facts struct{ id, scry, name string }
			all := map[string]facts{}
			note := func(cs []Card) {
				for _, c := range cs {
					all[c.ID] = facts{c.ID, c.ScryfallID.String(), c.Name}
				}
			}
			views := 0
			check := func(when string) {
				for _, id := range seats {
					note(d.HeldPack(id))
				}
				for i, id := range seats {
					allowed := map[string]bool{}
					for _, c := range d.HeldPack(id) {
						allowed[c.ID] = true
					}
					for _, c := range d.Pool(id) {
						allowed[c.ID] = true
					}
					v, ok := d.ViewFor(id, t0, nil)
					if !ok {
						t.Fatalf("no view for seat %d", i)
					}
					raw, err := json.Marshal(v)
					if err != nil {
						t.Fatal(err)
					}
					var decoded any
					if err := json.Unmarshal(raw, &decoded); err != nil {
						t.Fatal(err)
					}
					got := map[string]bool{}
					collectStrings(decoded, got)
					// Names and printings a seat may see, from its own cards.
					okNames, okScry := map[string]bool{}, map[string]bool{}
					for cid := range allowed {
						okNames[all[cid].name] = true
						okScry[all[cid].scry] = true
					}
					for _, f := range all {
						if allowed[f.id] {
							continue
						}
						if got[f.id] {
							t.Fatalf("%s: seat %d's view carries card %s, which it does not hold and has not drafted", when, i, f.id)
						}
						if got[f.scry] && !okScry[f.scry] {
							t.Fatalf("%s: seat %d's view carries the printing of %s (%s)", when, i, f.id, f.name)
						}
						if got[f.name] && !okNames[f.name] {
							t.Fatalf("%s: seat %d's view names %s (%s)", when, i, f.name, f.id)
						}
					}
					// And it does see its own.
					for cid := range allowed {
						if !got[cid] {
							t.Fatalf("%s: seat %d's view is missing its own card %s", when, i, cid)
						}
					}
					views++
				}
			}
			check("at the start")
			for d.State() == StatePicking {
				for n, id := range seats {
					if err := d.Pick(id, firstCards(d, id), t0); err != nil {
						t.Fatalf("Pick: %v", err)
					}
					check(fmt.Sprintf("round %d pick %d after seat %d picked", d.Round(), d.PickNumber(), n))
				}
			}
			if len(all) != 4*3*14 {
				t.Fatalf("the walk saw %d cards, want %d", len(all), 4*3*14)
			}
			if views < 100 {
				t.Fatalf("only %d views were walked", views)
			}
		})
	}
}

// The view type itself: cards live in the viewer's three fields and
// nowhere else, and a PlayerView has no field that could hold one.
func TestOnlyTheViewersOwnFieldsHoldCards(t *testing.T) {
	cardType := reflect.TypeOf(Card{})
	var holds func(t reflect.Type) bool
	holds = func(t reflect.Type) bool {
		switch t.Kind() {
		case reflect.Slice, reflect.Pointer, reflect.Array:
			return holds(t.Elem())
		case reflect.Map:
			return holds(t.Elem()) || holds(t.Key())
		case reflect.Struct:
			if t == cardType {
				return true
			}
			for i := range t.NumField() {
				if holds(t.Field(i).Type) {
					return true
				}
			}
		}
		return false
	}
	var got []string
	vt := reflect.TypeOf(View{})
	for i := range vt.NumField() {
		if holds(vt.Field(i).Type) {
			got = append(got, vt.Field(i).Name)
		}
	}
	sort.Strings(got)
	if want := []string{"Pack", "Picks", "Pool"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("View fields that can hold a card: %v, want %v. A new one must be the viewer's own, and TestNoSeatSeesAnotherSeatsCards must cover it.", got, want)
	}
	if holds(reflect.TypeOf(PlayerView{})) {
		t.Fatal("PlayerView can hold a card: every seat sees every PlayerView")
	}
	if holds(reflect.TypeOf(Preview{})) {
		t.Fatal("Preview can hold a card")
	}
}

func TestTheViewOfEachState(t *testing.T) {
	d, seats := newDraft(t, PickOne, 4, 2, 13)
	v, ok := d.ViewFor(seats[1], t0, func(id uuid.UUID) bool { return id == seats[0] })
	if !ok {
		t.Fatal("no view")
	}
	if v.State != StateLobby || v.You != 1 || len(v.Players) != 2 || v.Pack != nil || v.Deadline != nil || v.Pool != nil {
		t.Fatalf("lobby view: %+v", v)
	}
	if !v.Players[0].Host || !v.Players[0].Connected || v.Players[1].Connected {
		t.Fatalf("lobby players: %+v", v.Players)
	}
	if _, ok := d.ViewFor(uuid.New(), t0, nil); ok {
		t.Fatal("a view for a seat the draft does not have")
	}

	if err := d.Start(seats[0], testSource(t), t0); err != nil {
		t.Fatalf("Start: %v", err)
	}
	v, _ = d.ViewFor(seats[0], t0.Add(5*time.Second), nil)
	if v.State != StatePicking || v.Round != 1 || v.Rounds != 3 || v.Pick != 1 || v.PicksInRound != 14 ||
		v.Take != 1 || v.Passing != Left || len(v.Pack) != 14 || len(v.Players) != 4 || v.PickSeconds != 60 {
		t.Fatalf("picking view: %+v", v)
	}
	if v.Deadline == nil || !v.Deadline.Equal(t0.Add(time.Minute)) || v.Now == nil || !v.Now.Equal(t0.Add(5*time.Second)) {
		t.Fatalf("deadline %v now %v", v.Deadline, v.Now)
	}
	if !v.Players[0].Picking || v.Players[2].Picking || !v.Players[2].Bot || v.Players[2].Picked != 1 {
		t.Fatalf("players: %+v", v.Players)
	}
	if c := v.Pack[0]; c.ID == "" || c.ScryfallID == uuid.Nil || c.Name == "" || c.Set != "syn" || c.Rarity == "" || c.TypeLine == "" {
		t.Fatalf("a pack card is missing a field: %+v", c)
	}
	// Having picked, the seat holds no pack and waits.
	if err := d.Pick(seats[0], firstCards(d, seats[0]), t0); err != nil {
		t.Fatal(err)
	}
	v, _ = d.ViewFor(seats[0], t0, nil)
	if v.Pack != nil || v.Take != 0 || len(v.Picks) != 1 || v.Players[0].Picking || !v.Players[1].Picking {
		t.Fatalf("waiting view: pack %d take %d picks %d players %+v", len(v.Pack), v.Take, len(v.Picks), v.Players)
	}

	for d.State() == StatePicking {
		d.Expire(t0)
	}
	v, _ = d.ViewFor(seats[0], t0, nil)
	if v.State != StateDone || len(v.Pool) != 42 || len(v.Picks) != 42 || v.Pack != nil || v.Deadline != nil || v.Passing != "" {
		t.Fatalf("done view: state %s pool %d picks %d", v.State, len(v.Pool), len(v.Picks))
	}
	raw, _ := json.Marshal(v)
	if !strings.Contains(string(raw), `"state":"done"`) || !strings.Contains(string(raw), `"pool":[`) {
		t.Fatalf("the done view's JSON has no pool: %.200s", raw)
	}
}

// --- the pack source --------------------------------------------------------

func TestASetThatIsNotDraftableIsRefused(t *testing.T) {
	idx := synthIndex("syn", "2026-03-01")
	oneMissing := func(id uuid.UUID) bool { return id != synthID("oracle", "syn Common 03") }
	if _, err := NewSetSource(idx, "syn", oneMissing); !errors.Is(err, ErrNotDraftable) {
		t.Fatalf("a set with one card not Full = %v, want ErrNotDraftable", err)
	}
	if _, err := NewSetSource(idx, "nope", allFull); !errors.Is(err, ErrNotDraftable) {
		t.Fatalf("an unknown set = %v, want ErrNotDraftable", err)
	}
	if _, err := NewSetSource(nil, "syn", allFull); !errors.Is(err, ErrNotDraftable) {
		t.Fatalf("no index = %v, want ErrNotDraftable", err)
	}
	if _, err := NewSetSource(idx, " SYN ", allFull); err != nil {
		t.Fatalf("the set code is trimmed and lowercased: %v", err)
	}
}

func TestTheSourceIsTheDraftpacksGenerator(t *testing.T) {
	idx := synthIndex("syn", "2026-03-01")
	src, err := NewSetSource(idx, "syn", allFull)
	if err != nil {
		t.Fatal(err)
	}
	if src.Booster() != string(draftpacks.PoolFor(idx, "syn").Recipe.Kind) {
		t.Fatalf("booster %q", src.Booster())
	}
	got, err := src.Pack(rand.New(rand.NewPCG(1, 2)))
	if err != nil {
		t.Fatal(err)
	}
	want, err := draftpacks.Generate(idx, "syn", rand.New(rand.NewPCG(1, 2)))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want.Cards) {
		t.Fatalf("a pack of %d, draftpacks made %d", len(got), len(want.Cards))
	}
	for i, c := range got {
		w := want.Cards[i]
		if c.ScryfallID != w.ID || c.OracleID != w.OracleID || c.Name != w.Name || c.ID != "" {
			t.Fatalf("card %d: %+v, want %+v", i, c, w)
		}
		if !c.Basic && (len(c.Colors) != 1 || c.ManaCost == "" || c.ManaValue < 1) {
			t.Fatalf("card %d has no colour or cost: %+v", i, c)
		}
		if c.Colors == nil {
			t.Fatalf("card %d: colors must marshal as [], not null", i)
		}
	}
}

func TestManaValue(t *testing.T) {
	for cost, want := range map[string]int{
		"":                 0,
		"{0}":              0,
		"{2}{W}{W}":        4,
		"{X}{R}":           1,
		"{X}{X}{G}":        1,
		"{2/W}{2/W}":       4,
		"{W/U}{W/U}":       2,
		"{G/P}{G}":         2,
		"{10}{U}{U}":       12,
		"{1}{R} // {2}{U}": 5,
		"{C}{C}":           2,
		"{S}":              1,
		"{3}{B":            3,
		"{1}{G/U/P}":       2,
	} {
		if got := ManaValue(cost); got != want {
			t.Errorf("ManaValue(%q) = %d, want %d", cost, got, want)
		}
	}
}
