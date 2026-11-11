package draft

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// finishPicks plays every pick of a draft with the first cards of each
// pack, for the people; the bots pick for themselves.
func finishPicks(t *testing.T, d *Draft, people []uuid.UUID) {
	t.Helper()
	for guard := 0; d.State() == StatePicking; guard++ {
		if guard > 10000 {
			t.Fatal("the picks never ended")
		}
		for _, id := range people {
			if d.Take(id) > 0 {
				if err := d.Pick(id, firstCards(d, id), t0); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

// limitedDraft is a draft whose picks are done: seats seats, the first
// humans of them people.
func limitedDraft(t *testing.T, seats, humans int, seed uint64) (*Draft, []uuid.UUID) {
	t.Helper()
	d, people := newDraft(t, PickOne, seats, humans, seed)
	if err := d.Start(people[0], testSource(t), t0); err != nil {
		t.Fatal(err)
	}
	finishPicks(t, d, people)
	if d.State() != StateDone || d.Phase() != PhaseBuilding {
		t.Fatalf("after the picks: state %s, phase %q", d.State(), d.Phase())
	}
	return d, people
}

// legalDeck is a 40-card deck from a seat's pool: its first 23 cards
// and 17 basics.
func legalDeck(d *Draft, seat uuid.UUID) LimitedDeck {
	pool := d.Pool(seat)
	out := LimitedDeck{Basics: map[string]int{"Island": 9, "Swamp": 8}}
	for _, c := range pool[:23] {
		out.Main = append(out.Main, c.ID)
	}
	return out
}

func TestTheLastPickOpensTheBuild(t *testing.T) {
	d, people := limitedDraft(t, 8, 2, 11)
	if got := d.BuildDeadline(); !got.Equal(t0.Add(DefaultBuildTimeout)) {
		t.Fatalf("build deadline = %v", got)
	}
	// Every bot has a deck already; no person has.
	for i, s := range d.Seats() {
		_, built := d.Deck(i)
		if built != s.Bot {
			t.Fatalf("seat %d (bot %v) built = %v", i, s.Bot, built)
		}
	}
	v, _ := d.ViewFor(people[0], t0, nil)
	if v.Phase != PhaseBuilding || v.Build == nil || v.Build.DeckMin != 40 || v.Build.Submitted || v.Deck != nil {
		t.Fatalf("building view = %+v, build %+v", v.Phase, v.Build)
	}
	if len(v.Pool) == 0 {
		t.Fatal("the building view has no pool")
	}
}

func TestSubmitDeckRunsTheLimitedCheck(t *testing.T) {
	d, people := limitedDraft(t, 4, 2, 12)
	me, other := people[0], people[1]
	pool := d.Pool(me)

	short := legalDeck(d, me)
	short.Basics = map[string]int{"Island": 16}
	notMine := legalDeck(d, me)
	notMine.Main[0] = d.Pool(other)[0].ID
	twice := legalDeck(d, me)
	twice.Main[1] = twice.Main[0]
	notBasic := legalDeck(d, me)
	notBasic.Basics = map[string]int{pool[0].Name: 17}
	for name, deck := range map[string]LimitedDeck{
		"too few cards":           short,
		"a card of another seat":  notMine,
		"a card named twice":      twice,
		"a basic that is not one": notBasic,
	} {
		if err := d.SubmitDeck(me, deck, t0); !errors.Is(err, ErrDeckIllegal) {
			t.Errorf("%s: err = %v, want ErrDeckIllegal", name, err)
		}
	}
	if _, built := d.Deck(0); built {
		t.Fatal("a refused deck was kept")
	}

	// Basics are free: a deck of one drafted card and 39 Forests.
	if err := d.SubmitDeck(me, LimitedDeck{Main: []string{pool[0].ID}, Basics: map[string]int{"Forest": 39}}, t0); err != nil {
		t.Fatalf("one card and 39 basics: %v", err)
	}
	// The rounds wait for the other person, and a deck may be changed
	// until then.
	if d.Phase() != PhaseBuilding {
		t.Fatalf("phase = %s with a person still building", d.Phase())
	}
	if err := d.SubmitDeck(me, legalDeck(d, me), t0); err != nil {
		t.Fatal(err)
	}
	v, _ := d.ViewFor(me, t0, nil)
	if !v.Build.Submitted || v.Deck == nil || v.Deck.Size() != 40 {
		t.Fatalf("after submitting: %+v, deck %+v", v.Build, v.Deck)
	}
	// The sideboard is the rest of the pool (CR 100.4b).
	main, basics, side, _ := d.DeckCards(0)
	if len(main) != 23 || basics["Island"] != 9 || len(main)+len(side) != len(pool) {
		t.Fatalf("deck cards: %d main, %v basics, %d side of a pool of %d", len(main), basics, len(side), len(pool))
	}
	if err := d.SubmitDeck(other, legalDeck(d, other), t0); err != nil {
		t.Fatal(err)
	}
	if d.Phase() != PhaseRounds || d.MatchRound() != 1 {
		t.Fatalf("after the last deck: phase %s, round %d", d.Phase(), d.MatchRound())
	}
	if err := d.SubmitDeck(me, legalDeck(d, me), t0); !errors.Is(err, ErrNotBuilding) {
		t.Fatalf("a deck after the build: %v", err)
	}
}

// When the build timer runs out, a person who has not submitted gets
// the DeckBuilder's deck from their own pool.
func TestTheBuildTimerFallsBackToTheDeckBuilder(t *testing.T) {
	d, people := limitedDraft(t, 4, 2, 13)
	if err := d.SubmitDeck(people[0], legalDeck(d, people[0]), t0); err != nil {
		t.Fatal(err)
	}
	d.ExpireBuild(t0.Add(DefaultBuildTimeout - time.Second))
	if d.Phase() != PhaseBuilding {
		t.Fatal("the build ended before its deadline")
	}
	d.ExpireBuild(t0.Add(DefaultBuildTimeout))
	if d.Phase() != PhaseRounds {
		t.Fatalf("phase = %s after the deadline", d.Phase())
	}
	mine, _ := d.Deck(0)
	if len(mine.Main) != 23 || mine.Basics["Island"] != 9 {
		t.Fatalf("the submitted deck was replaced: %+v", mine)
	}
	late, ok := d.Deck(1)
	if !ok || late.Size() < 40 {
		t.Fatalf("the late seat's deck = %+v, %v", late, ok)
	}
	pool := map[string]bool{}
	for _, c := range d.Pool(people[1]) {
		pool[c.ID] = true
	}
	for _, id := range late.Main {
		if !pool[id] {
			t.Fatalf("the DeckBuilder's deck holds %s, which is not in the seat's pool", id)
		}
	}
}

type fixedBuilder struct{ deck LimitedDeck }

func (b fixedBuilder) BuildLimitedDeck([]Card) LimitedDeck { return b.deck }

// Config.Builder is the seam slice 5 fills; a builder whose deck is not
// legal is overruled.
func TestTheConfiguredDeckBuilderIsUsedAndABrokenOneOverruled(t *testing.T) {
	for _, tc := range []struct {
		name  string
		deck  LimitedDeck
		wants func(LimitedDeck) bool
	}{
		{"used", LimitedDeck{Basics: map[string]int{"Mountain": 41}}, func(d LimitedDeck) bool { return d.Basics["Mountain"] == 41 }},
		{"overruled", LimitedDeck{Main: []string{"not a card"}}, func(d LimitedDeck) bool { return d.Size() >= 40 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := New(Config{Mode: PickOne, Seats: 2, Seed: 5, Builder: fixedBuilder{tc.deck}}, t0)
			if err != nil {
				t.Fatal(err)
			}
			s, _ := d.AddSeat("A", uuid.Nil)
			if err := d.Start(s.ID, testSource(t), t0); err != nil {
				t.Fatal(err)
			}
			finishPicks(t, d, []uuid.UUID{s.ID})
			bot, ok := d.Deck(1)
			if !ok || !tc.wants(bot) {
				t.Fatalf("bot deck = %+v, %v", bot, ok)
			}
		})
	}
}

func TestSimpleDeckBuilder(t *testing.T) {
	var pool []Card
	add := func(n int, colour string, mv int, typeLine string) {
		for range n {
			c := Card{ID: fmt.Sprintf("c%d", len(pool)+1), Name: fmt.Sprintf("Card %d", len(pool)+1), ManaValue: mv, TypeLine: typeLine}
			if colour != "" {
				c.Colors = []string{colour}
			}
			pool = append(pool, c)
		}
	}
	add(14, "R", 2, "Creature")
	add(10, "G", 3, "Creature")
	add(6, "G", 7, "Sorcery")
	add(8, "U", 1, "Instant")
	add(3, "", 4, "Artifact")
	add(2, "", 0, "Land")
	deck := SimpleDeckBuilder{}.BuildLimitedDeck(pool)
	if len(deck.Main) != 23 || deck.Size() != 40 {
		t.Fatalf("deck has %d pool cards and %d in all", len(deck.Main), deck.Size())
	}
	byID := map[string]Card{}
	for _, c := range pool {
		byID[c.ID] = c
	}
	dear := 0
	for _, id := range deck.Main {
		c := byID[id]
		if len(c.Colors) > 0 && c.Colors[0] == "U" {
			t.Fatalf("the deck plays %s, a card of a third colour", c.Name)
		}
		if c.TypeLine == "Land" {
			t.Fatal("the deck counts a land among its 23 cards")
		}
		if c.ManaValue == 7 {
			dear++
		}
	}
	// 14 red, 10 green and 3 colourless are cheaper than the sevens.
	if dear != 0 {
		t.Fatalf("the deck plays %d seven-drops ahead of cheaper cards", dear)
	}
	if deck.Basics["Mountain"] < 3 || deck.Basics["Forest"] < 3 || deck.Basics["Mountain"]+deck.Basics["Forest"] != 17 || len(deck.Basics) != 2 {
		t.Fatalf("basics = %v", deck.Basics)
	}

	// A thin pool is topped up with lands to 40.
	thin := SimpleDeckBuilder{}.BuildLimitedDeck(pool[:5])
	if thin.Size() != 40 {
		t.Fatalf("a thin pool made a deck of %d", thin.Size())
	}
}

// buildAll submits a deck for every person, which begins the rounds.
func buildAll(t *testing.T, d *Draft, people []uuid.UUID) {
	t.Helper()
	for _, id := range people {
		if err := d.SubmitDeck(id, legalDeck(d, id), t0); err != nil {
			t.Fatal(err)
		}
	}
	if d.Phase() != PhaseRounds {
		t.Fatalf("phase = %s after every deck", d.Phase())
	}
}

// playRound reports every table match of the round in play as won by
// its A seat, opening a table for each first, and returns how many
// there were.
func playRound(t *testing.T, d *Draft) int {
	t.Helper()
	needs := d.TablesNeeded()
	for _, n := range needs {
		if err := d.SetTable(n.Round, n.Index, uuid.New()); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range needs {
		if err := d.ReportMatch(n.Round, n.Index, n.A, 1, 0, t0); err != nil {
			t.Fatal(err)
		}
	}
	return len(needs)
}

// Three rounds, and no seat meets a seat twice, for every size of draft
// that can avoid it and for 2, 3, 5 and 8 people among bots.
func TestRoundsPairWithoutRepeats(t *testing.T) {
	type size struct{ seats, humans int }
	sizes := []size{{8, 2}, {8, 3}, {8, 5}, {8, 8}, {8, 1}, {3, 3}, {5, 5}, {4, 2}, {6, 3}, {7, 5}, {5, 2}}
	for _, sz := range sizes {
		for seed := uint64(1); seed <= 12; seed++ {
			t.Run(fmt.Sprintf("%d seats %d people seed %d", sz.seats, sz.humans, seed), func(t *testing.T) {
				d, people := limitedDraft(t, sz.seats, sz.humans, seed)
				buildAll(t, d, people)
				met := map[[2]int]bool{}
				satOut := map[int]int{}
				for round := 1; round <= MatchRounds; round++ {
					if d.MatchRound() != round {
						t.Fatalf("round in play = %d, want %d", d.MatchRound(), round)
					}
					pairs := d.Pairings(round)
					seen := map[int]bool{}
					humanVsHuman, humanVsBot, humanBye, botBye := 0, 0, 0, 0
					for _, p := range pairs {
						for _, s := range []int{p.A, p.B} {
							if s == NoSeat {
								continue
							}
							if seen[s] {
								t.Fatalf("round %d seats %d twice", round, s)
							}
							seen[s] = true
						}
						if p.Bye() {
							satOut[p.A]++
							if p.BotA {
								botBye++
							} else {
								humanBye++
							}
							continue
						}
						key := pairKey(p.A, p.B)
						if met[key] {
							t.Fatalf("round %d pairs seats %d and %d, who have already played", round, p.A, p.B)
						}
						met[key] = true
						switch {
						case !p.BotA && !p.BotB:
							humanVsHuman++
						case p.BotA != p.BotB:
							humanVsBot++
						default:
							// Two bots: tossed for as the round began.
							if !p.Done || !p.Coin || p.Table != uuid.Nil {
								t.Fatalf("round %d: a bot match was not tossed for: %+v", round, p)
							}
						}
					}
					if len(seen) != sz.seats {
						t.Fatalf("round %d seats %d of %d", round, len(seen), sz.seats)
					}
					// With an odd number of seats a bot sits out when there is one.
					if sz.seats%2 == 1 && sz.humans < sz.seats && humanBye > 0 {
						t.Fatalf("round %d: a person sits out with a bot at the table", round)
					}
					// People meet people first: in round one nobody has
					// played, so at most one person is left for a bot.
					if round == 1 && humanVsHuman != sz.humans/2 && sz.seats%2 == 0 {
						t.Fatalf("round 1: %d matches between people, want %d", humanVsHuman, sz.humans/2)
					}
					if got := playRound(t, d); got != humanVsHuman+humanVsBot {
						t.Fatalf("round %d: %d tables, want %d", round, got, humanVsHuman+humanVsBot)
					}
				}
				// A bot sits out before a person does, so a bot may sit
				// out twice; no person does while another could.
				for s, n := range satOut {
					if n > 1 && s < sz.humans && sz.seats > 3 {
						t.Fatalf("person in seat %d sat out %d rounds", s, n)
					}
				}
				if d.Phase() != PhaseFinished {
					t.Fatalf("phase = %s after three rounds", d.Phase())
				}
				// Every seat's record adds up to the matches it played.
				for i, st := range d.Standings() {
					if got, want := st.Wins+st.Losses+st.Draws, MatchRounds-satOut[i]; got != want {
						t.Fatalf("seat %d has %d results, want %d", i, got, want)
					}
				}
			})
		}
	}
}

// A draft of two seats has one possible pairing: its later rounds are
// rematches, because there is nobody else to play.
func TestATwoSeatDraftReplaysItsOnePairing(t *testing.T) {
	d, people := limitedDraft(t, 2, 2, 3)
	buildAll(t, d, people)
	for round := 1; round <= MatchRounds; round++ {
		if got := playRound(t, d); got != 1 {
			t.Fatalf("round %d: %d tables", round, got)
		}
	}
	if d.Phase() != PhaseFinished {
		t.Fatalf("phase = %s", d.Phase())
	}
	st := d.Standings()
	if st[0].Wins+st[0].Losses != 3 || st[1].Wins+st[1].Losses != 3 {
		t.Fatalf("standings = %+v", st)
	}
}

// The same seed pairs the same rounds and tosses the same coins.
func TestRoundsReplayFromTheSeed(t *testing.T) {
	run := func() string {
		d, people := limitedDraft(t, 8, 1, 77)
		buildAll(t, d, people)
		out := ""
		for round := 1; round <= MatchRounds; round++ {
			for _, p := range d.Pairings(round) {
				out += fmt.Sprintf("%d:%d-%d w%d c%v|", round, p.A, p.B, p.Winner, p.Coin)
			}
			playRound(t, d)
		}
		return out
	}
	if a, b := run(), run(); a != b {
		t.Fatalf("two runs of one seed differ:\n%s\n%s", a, b)
	}
}

func TestResultsAndStandings(t *testing.T) {
	d, people := limitedDraft(t, 4, 2, 21)
	buildAll(t, d, people)
	n := d.TablesNeeded()
	if len(n) != 1 {
		t.Fatalf("round 1 of two people among four seats needs %d tables, want 1", len(n))
	}
	table := uuid.New()
	if err := d.SetTable(n[0].Round, n[0].Index, table); err != nil {
		t.Fatal(err)
	}
	// Both seats of the match are told its table; the view of a seat
	// carries no other.
	for i, id := range people {
		v, _ := d.ViewFor(id, t0, nil)
		if v.Play == nil || v.Play.Table == nil || *v.Play.Table != table || v.Play.Yours == nil {
			t.Fatalf("seat %d: play = %+v", i, v.Play)
		}
	}
	// A winner who is not in the match is refused.
	other := 0
	for other == n[0].A || other == n[0].B {
		other++
	}
	if err := d.ReportMatch(n[0].Round, n[0].Index, other, 1, 0, t0); !errors.Is(err, ErrNoSuchMatch) {
		t.Fatalf("a winner outside the match: %v", err)
	}
	if err := d.ReportMatch(n[0].Round, n[0].Index, n[0].B, 1, 2, t0); err != nil {
		t.Fatal(err)
	}
	// A second report changes nothing.
	if err := d.ReportMatch(n[0].Round, n[0].Index, n[0].A, 2, 0, t0); err != nil {
		t.Fatal(err)
	}
	st := d.Standings()
	if st[n[0].B].Wins != 1 || st[n[0].A].Losses != 1 {
		t.Fatalf("standings = %+v", st)
	}
	res := d.Results(n[0].A)
	if len(res) != 1 || res[0].Result != "loss" || res[0].GamesWon != 1 || res[0].GamesLost != 2 || res[0].OpponentBot {
		t.Fatalf("results = %+v", res)
	}
	if d.MatchRound() != 2 {
		t.Fatalf("round = %d after round 1's last result", d.MatchRound())
	}
}

// A person may leave after any round, or during one: the seat is a bot
// from then on, and the draft ends when nobody is left.
func TestLeavingAfterThePicks(t *testing.T) {
	t.Run("during the build", func(t *testing.T) {
		d, people := limitedDraft(t, 4, 2, 31)
		if err := d.SubmitDeck(people[0], legalDeck(d, people[0]), t0); err != nil {
			t.Fatal(err)
		}
		if err := d.Leave(people[1], t0); err != nil {
			t.Fatal(err)
		}
		// The seat is built for, and the rounds begin.
		if _, ok := d.Deck(1); !ok || d.Phase() != PhaseRounds {
			t.Fatalf("phase %s", d.Phase())
		}
	})
	t.Run("between rounds", func(t *testing.T) {
		d, people := limitedDraft(t, 4, 2, 32)
		buildAll(t, d, people)
		playRound(t, d)
		if err := d.Leave(people[0], t0); err != nil {
			t.Fatal(err)
		}
		if d.Phase() != PhaseRounds {
			t.Fatalf("phase %s with a person still in", d.Phase())
		}
		// The seat that left plays on as a bot.
		for _, p := range d.Pairings(d.MatchRound()) {
			if p.has(0) && p.has(1) && !p.Done && len(d.TablesNeeded()) != 1 {
				t.Fatal("a match between a person and the seat that left has no table to open")
			}
		}
		if err := d.Leave(people[1], t0); err != nil {
			t.Fatal(err)
		}
		if d.Phase() != PhaseFinished {
			t.Fatalf("phase %s with nobody left", d.Phase())
		}
	})
	t.Run("a match nobody is left to play is tossed for", func(t *testing.T) {
		// A third person keeps the draft alive.
		d, people := limitedDraft(t, 4, 3, 33)
		buildAll(t, d, people)
		before := d.MatchRound()
		// Find a match between two people and have both leave it before
		// its table opens.
		var a, b = -1, -1
		for _, p := range d.Pairings(before) {
			if !p.BotA && !p.BotB && !p.Bye() {
				a, b = p.A, p.B
			}
		}
		if a < 0 {
			t.Skip("this seed paired no two people")
		}
		if err := d.Leave(people[a], t0); err != nil {
			t.Fatal(err)
		}
		if err := d.Leave(people[b], t0); err != nil {
			t.Fatal(err)
		}
		for _, p := range d.Pairings(before) {
			if p.has(a) && p.has(b) && (!p.Done || !p.Coin) {
				t.Fatalf("the abandoned match was not tossed for: %+v", p)
			}
		}
	})
}

// No view of the rounds carries another seat's deck or table.
func TestTheRoundsViewHidesOtherSeats(t *testing.T) {
	d, people := limitedDraft(t, 8, 4, 41)
	buildAll(t, d, people)
	for _, n := range d.TablesNeeded() {
		if err := d.SetTable(n.Round, n.Index, uuid.New()); err != nil {
			t.Fatal(err)
		}
	}
	pairs := d.Pairings(1)
	for i, id := range people {
		v, _ := d.ViewFor(id, t0, nil)
		var mine uuid.UUID
		for _, p := range pairs {
			if p.has(i) {
				mine = p.Table
			}
		}
		if v.Play.Table == nil || *v.Play.Table != mine {
			t.Fatalf("seat %d is told table %v, its own is %v", i, v.Play.Table, mine)
		}
		own, _ := d.Deck(i)
		if v.Deck == nil || len(v.Deck.Main) != len(own.Main) {
			t.Fatalf("seat %d does not see its own deck", i)
		}
	}
}

// The room's timer ends the build: a person who never submitted is
// built for, their socket is told, and the lobby's watcher is woken to
// open the round's tables.
func TestTheRoomsBuildTimer(t *testing.T) {
	r, clock, people := newRoom(t, PickOne, 2, 1, 51)
	if err := r.Start(people[0]); err != nil {
		t.Fatal(err)
	}
	s := &sink{}
	detach, err := r.Attach(people[0], s)
	if err != nil {
		t.Fatal(err)
	}
	defer detach()
	for guard := 0; s.last(t).Phase != PhaseBuilding; guard++ {
		if guard > 1000 {
			t.Fatal("the picks never ended")
		}
		v := s.last(t)
		ids := make([]string, 0, v.Take)
		for _, c := range v.Pack[:v.Take] {
			ids = append(ids, c.ID)
		}
		if err := r.Pick(people[0], ids); err != nil {
			t.Fatal(err)
		}
	}
	// An illegal deck is refused and the build stays open.
	if err := r.SubmitDeck(people[0], LimitedDeck{}); !errors.Is(err, ErrDeckIllegal) {
		t.Fatalf("an empty deck: %v", err)
	}
	clock.Advance(DefaultBuildTimeout - time.Second)
	if got := s.last(t).Phase; got != PhaseBuilding {
		t.Fatalf("phase = %s a second before the deadline", got)
	}
	select {
	case <-r.Wake():
	default:
	}
	clock.Advance(2 * time.Second)
	v := s.last(t)
	if v.Phase != PhaseRounds || v.Deck == nil || v.Deck.Size() < 40 || v.Play == nil || v.Play.Yours == nil {
		t.Fatalf("after the deadline: phase %s, deck %+v, play %+v", v.Phase, v.Deck, v.Play)
	}
	select {
	case <-r.Wake():
	default:
		t.Fatal("the watcher was not woken when the rounds began")
	}
	var needs []TableNeed
	r.With(func(d *Draft) { needs = d.TablesNeeded() })
	if len(needs) != 1 {
		t.Fatalf("the round needs %d tables, want 1", len(needs))
	}
	table := uuid.New()
	if err := r.SetTable(needs[0].Round, needs[0].Index, table); err != nil {
		t.Fatal(err)
	}
	if got := s.last(t).Play.Table; got == nil || *got != table {
		t.Fatalf("the seat was not told its table: %v", got)
	}
	if err := r.ReportMatch(needs[0].Round, needs[0].Index, needs[0].A, 1, 0); err != nil {
		t.Fatal(err)
	}
	if got := s.last(t).Play.Round; got != 2 {
		t.Fatalf("round = %d after the result", got)
	}
}
