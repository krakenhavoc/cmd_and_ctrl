package gamestats

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

func ip(n int) *int   { return &n }
func bp(b bool) *bool { return &b }
func fv(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}

// framed is a version-2 seat with frames.
func framed(result Result, place, players int) SeatRow {
	return SeatRow{Result: result, Place: ip(place), Players: players, StatsVersion: 2, FramesComplete: bp(true),
		Turns: 10, Rounds: 3, ScrewGame: bp(false), FloodGame: bp(false), ColourScrewGame: bp(false),
		CreaturesEligible: ip(0), CreaturesAttacked: ip(0), CommanderCasts: ip(0), LandDropsHit: ip(0), LandDropsMissed: ip(0)}
}

func TestSummariseRecordPlacesAndKinds(t *testing.T) {
	a := framed(ResultWin, 1, 4)
	a.WinKind, a.ComboRule, a.TurnPosition = KindCombo, ComboRepeats, ip(1)
	b := framed(ResultLoss, 3, 4)
	b.WinKind, b.OutKind, b.TurnPosition = KindCombat, KindCombat, ip(1) // someone else's combat win
	c := framed(ResultLoss, 2, 2)
	c.OutKind = KindPoison
	v1 := SeatRow{Result: ResultWin, Place: ip(1), Players: 4, StatsVersion: 1, Turns: 6, Rounds: 2, Mulligans: 2}
	s := Summarise([]SeatRow{a, b, c, v1})

	if s.Coverage != (SummaryCoverage{Games: 4, V2: 3, Frames: 3}) {
		t.Errorf("coverage %+v", s.Coverage)
	}
	if s.Record.Games != 4 || s.Record.Wins != 2 || s.WinRate.Rate == nil || *s.WinRate.Rate != 0.5 {
		t.Errorf("record %+v rate %+v", s.Record, s.WinRate)
	}
	if len(s.Places) != 2 || s.Places[0].Players != 2 || s.Places[1].Players != 4 {
		t.Fatalf("places by size %+v", s.Places)
	}
	if got := s.Places[1].Places; got[0] != 2 || got[2] != 1 || len(got) != 4 {
		t.Errorf("four-seat places %v", got)
	}
	// Only the person's own wins have a win kind; how they lose is OutKinds.
	if len(s.WinKinds) != 1 || s.WinKinds[0] != (KindCount{Kind: "combo", Games: 1}) {
		t.Errorf("win kinds %v", s.WinKinds)
	}
	if len(s.ComboRules) != 1 || s.ComboRules[0].Kind != "repeats" {
		t.Errorf("combo rules %v", s.ComboRules)
	}
	if len(s.OutKinds) != 2 || s.OutKinds[0].Kind != "poison" || s.OutKinds[1].Kind != "combat" {
		t.Errorf("out kinds %v (in ADR order)", s.OutKinds)
	}
	// Two four-seat games at position 1: below MinGames, so no rate.
	if len(s.TurnOrder) != 1 || s.TurnOrder[0].Wins.Of != 2 || s.TurnOrder[0].Wins.Rate != nil {
		t.Errorf("turn order %+v", s.TurnOrder)
	}
	if s.Mulligans.Rows[2].Wins.N != 1 || s.Mulligans.Rows[0].Wins.Of != 3 {
		t.Errorf("mulligans %+v", s.Mulligans.Rows)
	}
	if s.Length.Turns.Median == nil || *s.Length.Turns.Median != 10 {
		t.Errorf("median turns %v", fv(s.Length.Turns.Median))
	}
	if s.Length.Minutes.Median != nil {
		t.Error("minutes without start times")
	}
}

func TestSummariseManaAndAggression(t *testing.T) {
	var seats []SeatRow
	for i := range 3 {
		r := framed(ResultLoss, 2, 4)
		r.ScrewGame = bp(i == 0)
		r.FloodGame = bp(i < 2)
		r.CreaturesEligible, r.CreaturesAttacked = ip(4), ip(2)
		// Own turns 1–3: land 1 and 2 hit, the 3rd missed; one more land
		// on the battlefield by turn 3 from ramp.
		r.TurnRows = []TurnRow{
			{OwnTurn: 1, LandPlayed: 1, LandDrops: 1, LandsControlled: 0, Attackers: 0, DamageToPlayers: 0},
			{OwnTurn: 2, LandPlayed: 1, LandDrops: 2, LandsControlled: 1, Attackers: 1, DamageToPlayers: 2},
			{OwnTurn: 3, LandPlayed: 0, LandDrops: 2, LandsControlled: 3, Missed: true, Attackers: 2, DamageToPlayers: 4},
		}
		r.Targets = []TargetRow{{TargetSeat: 1, Damage: 6}, {TargetSeat: 2, Damage: 3}, {TargetSeat: 3, Damage: 3}}
		r.Cast = []CastRow{{Name: "Sol Ring", Casts: 1}, {Name: "Bolt", Casts: 3}}
		seats = append(seats, r)
	}
	// A game whose replay was not read adds nothing frame-derived.
	noFrames := SeatRow{Result: ResultWin, Place: ip(1), Players: 4, StatsVersion: 2, FramesComplete: bp(false)}
	noFrames.Cast = []CastRow{{Name: "Bolt", Casts: 1}}
	seats = append(seats, noFrames)
	s := Summarise(seats)

	m := s.Mana
	if m.Screw.N != 1 || m.Screw.Of != 3 || m.Flood.N != 2 || m.ColourScrew.N != 0 || m.ColourScrew.Rate == nil {
		t.Errorf("screw %+v flood %+v colour %+v", m.Screw, m.Flood, m.ColourScrew)
	}
	if len(m.Trend) != 3 || !m.Trend[0].Screw || m.Trend[2].Flood {
		t.Errorf("trend %+v", m.Trend)
	}
	if m.MissedByLand[2] != (LandMiss{Land: 3, Missed: 3}) || m.MissedByLand[0].Missed != 0 || len(m.MissedByLand) != LandMissTop {
		t.Errorf("missed by land %+v", m.MissedByLand)
	}
	curve := m.LandCurve
	if len(curve) != LandCurveTop || curve[1].Hit.N != 3 || curve[2].Hit.N != 0 || curve[2].Hit.Of != 3 {
		t.Errorf("land curve hit %+v", curve)
	}
	// By own turn 3: 3 lands at the start of main, none played: ramp
	// counts in Controlled, not in Hit.
	if curve[2].Controlled.N != 3 || curve[3].Hit.Of != 0 || curve[3].Hit.Rate != nil {
		t.Errorf("land curve controlled %+v", curve)
	}

	ag := s.Aggression
	if ag.OwnTurns != 9 || ag.AttacksPerTurn == nil || *ag.AttacksPerTurn != 1 || *ag.DamagePerTurn != 2 {
		t.Errorf("aggression %+v", ag)
	}
	if ag.AttackShare == nil || *ag.AttackShare != 0.5 {
		t.Errorf("attack share %v", fv(ag.AttackShare))
	}
	if ag.SpreadGames != 3 || ag.MostHitShare == nil || *ag.MostHitShare != 0.5 {
		t.Errorf("most hit %v over %d", fv(ag.MostHitShare), ag.SpreadGames)
	}
	h := -(0.5*math.Log(0.5) + 2*0.25*math.Log(0.25))
	if ag.Evenness == nil || math.Abs(*ag.Evenness-h/math.Log(3)) > 1e-9 {
		t.Errorf("evenness %v, want %v", fv(ag.Evenness), h/math.Log(3))
	}
	if len(ag.OverTurns) != 3 || ag.OverTurns[2].OwnTurn != 3 || *ag.OverTurns[2].AttacksPerTurn != 2 {
		t.Errorf("over turns %+v", ag.OverTurns)
	}
	if len(s.MostCast) != 2 || s.MostCast[0] != (CastCard{Name: "Bolt", Games: 4, Casts: 10}) {
		t.Errorf("most cast %+v", s.MostCast)
	}
}

func TestSpreadNeedsTwoOpponentsAndDamage(t *testing.T) {
	if _, _, ok := spreadOf([]TargetRow{{Damage: 5}}, 1); ok {
		t.Error("a two-seat game has a spread")
	}
	if _, _, ok := spreadOf([]TargetRow{{Damage: 0}, {Damage: 0}}, 3); ok {
		t.Error("no damage has a spread")
	}
	share, even, ok := spreadOf([]TargetRow{{Damage: 9}}, 3)
	if !ok || share != 1 || even != 0 {
		t.Errorf("all at one: %v %v %v", share, even, ok)
	}
	share, even, _ = spreadOf([]TargetRow{{Damage: 2}, {Damage: 2}, {Damage: 2}}, 3)
	if math.Abs(share-1.0/3) > 1e-9 || math.Abs(even-1) > 1e-9 {
		t.Errorf("even: %v %v", share, even)
	}
}

func TestMakeupAndRangeFilters(t *testing.T) {
	people := &SeatRow{OtherPeople: 3}
	mixed := &SeatRow{OtherPeople: 1, OtherBots: 2}
	bots := &SeatRow{OtherBots: 3}
	for _, c := range []struct {
		m    MakeupFilter
		want [3]bool
	}{
		{MakeupAll, [3]bool{true, true, true}},
		{MakeupPeople, [3]bool{true, false, false}},
		{MakeupWithBots, [3]bool{false, true, false}},
		{MakeupBots, [3]bool{false, false, true}},
	} {
		got := [3]bool{c.m.Matches(people), c.m.Matches(mixed), c.m.Matches(bots)}
		if got != c.want {
			t.Errorf("%s: %v, want %v", c.m, got, c.want)
		}
	}
	if m, ok := ParseMakeup(""); !ok || m != MakeupAll {
		t.Error(`"" is not all`)
	}
	if _, ok := ParseMakeup("humans"); ok {
		t.Error("an unknown makeup parsed")
	}
	now := time.UnixMilli(100 * 24 * 3600 * 1000)
	if RangeAll.Since(now) != 0 || Range30.Since(now) != now.Add(-30*24*time.Hour).UnixMilli() {
		t.Error("Since")
	}
	if _, ok := ParseRange("7d"); ok {
		t.Error("an unknown range parsed")
	}
}

func TestDeadCards(t *testing.T) {
	seen := func(name string, affordable int, cast bool) SeenRow {
		return SeenRow{Name: name, AffordableTurns: affordable, Cast: cast}
	}
	seats := []SeatRow{
		{Seen: []SeenRow{seen("Craterhoof", 0, false), seen("Bolt", 2, false), seen("Forest", 0, false), seen("Ramp", 1, true)}},
		{Seen: []SeenRow{seen("Craterhoof", 0, false), seen("Bolt", 0, true)}},
		{Seen: []SeenRow{seen("Craterhoof", 3, false), seen("Bolt", 1, true)}},
	}
	got := DeadCards(seats, func(name string) bool { return name == "Forest" })
	if len(got) != 2 {
		t.Fatalf("dead cards %+v (a land and an always-cast card are not dead)", got)
	}
	hoof, bolt := got[0], got[1]
	if hoof.Name != "Craterhoof" || hoof.Dead.N != 3 || hoof.Dead.Of != 3 || hoof.Dead.Rate == nil || hoof.CouldAfford {
		t.Errorf("craterhoof %+v: affordable in 1 of 3 dead games is couldn't afford", hoof)
	}
	if bolt.Name != "Bolt" || bolt.Dead.N != 1 || bolt.Dead.Of != 3 || !bolt.CouldAfford {
		t.Errorf("bolt %+v", bolt)
	}
}

// putV2 writes one finished four-seat game with a version-2 record:
// seat 0 is userID's with deckID, seat 1 a bot carrying the same user
// id (never theirs), seats 2 and 3 other people.
func putV2(t *testing.T, s *SQLStore, d *db.DB, at int64, userID, deckID, hash string, result Result) uuid.UUID {
	t.Helper()
	id := uuid.New()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := d.Exec(q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	exec(`INSERT INTO games (id, name, state, created_at, started_at, ended_at) VALUES (?, 'Table', 'ended', 1, ?, ?)`,
		id.String(), at-30*60000, at)
	exec(`INSERT OR IGNORE INTO decks (id, owner_id, name, source_format, source_text, commanders, card_count, created_at, updated_at)
	      VALUES (?, ?, ?, 'text', '', '[]', 100, 1, 1)`, deckID, userID, "Deck "+deckID)
	exec(`INSERT INTO seats (game_id, seat, player_id, user_id, deck_id, deck_name) VALUES (?, 0, 'p0', ?, ?, 'Deck')`, id.String(), userID, deckID)
	exec(`INSERT INTO seats (game_id, seat, player_id, user_id, bot_tier) VALUES (?, 1, 'p1', ?, 'heuristic')`, id.String(), userID)
	exec(`INSERT INTO seats (game_id, seat, player_id, user_id) VALUES (?, 2, 'p2', 'u-other')`, id.String())
	exec(`INSERT INTO seats (game_id, seat, player_id) VALUES (?, 3, 'p3')`, id.String())
	place := 1
	if result != ResultWin {
		place = 2
	}
	g := Game{GameID: id, Version: Version, Source: SourceLive, LogComplete: true, Turns: 12, Rounds: 3,
		StartingSeat: ip(0), ComputedAt: time.UnixMilli(at), FramesComplete: true, WinKind: KindCombat}
	for seat := range 4 {
		st := Seat{Seat: seat, Result: ResultLoss, Place: ip(seat + 2), Commanders: []string{"Cmdr"}, TurnPosition: ip(seat + 1)}
		switch seat {
		case 0:
			st.Result, st.Place, st.DeckHash = result, ip(place), hash
			st.Cards = []Card{{Name: "Sol Ring", Casts: 1, FirstCastTurn: 1}}
			st.Targets = []TargetStat{{TargetSeat: 1, Damage: 4}, {TargetSeat: 2, Damage: 4}, {TargetSeat: 3}}
			st.Play = &Play{OwnTurns: 1, Turns: []TurnStat{{OwnTurn: 1, Turn: 1, LandPlayed: 1, LandDrops: 1, Attackers: 1}},
				Seen: []SeenCard{{Name: "Bolt", Copies: 1}}}
		case 3:
			st.Cards = []Card{{Name: "Not Mine", Casts: 9}}
		}
		g.Seats = append(g.Seats, st)
	}
	if err := s.Put(context.Background(), g); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestAdvancedSeatsReadsOnlyTheCallersSeats(t *testing.T) {
	s, d := openStore(t)
	for _, u := range []string{"u-me", "u-other"} {
		if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, 1, 1)`, u, u); err != nil {
			t.Fatal(err)
		}
	}
	putV2(t, s, d, 1000, "u-me", "d1", "aaaaaaaaaaaaaaaa", ResultWin)
	newer := putV2(t, s, d, 2000, "u-me", "d2", "bbbbbbbbbbbbbbbb", ResultLoss)
	putV2(t, s, d, 3000, "u-other", "d3", "cccccccccccccccc", ResultWin)

	rows, err := s.AdvancedSeats(context.Background(), "u-me", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].GameID != newer.String() {
		t.Fatalf("rows %d, newest first: %+v", len(rows), rows)
	}
	r := rows[0]
	if r.Seat != 0 || r.OtherBots != 1 || r.OtherPeople != 2 || r.Players != 4 || r.DeckHash != "bbbbbbbbbbbbbbbb" ||
		r.TurnPosition == nil || *r.TurnPosition != 1 || r.StartedAt != 2000-30*60000 {
		t.Errorf("seat row %+v", r)
	}
	if len(r.TurnRows) != 1 || r.TurnRows[0].Attackers != 1 || len(r.Targets) != 3 || len(r.Seen) != 1 {
		t.Errorf("children turns %v targets %v seen %v", r.TurnRows, r.Targets, r.Seen)
	}
	if len(r.Cast) != 1 || r.Cast[0].Name != "Sol Ring" {
		t.Errorf("casts %v: only the caller's own seat's", r.Cast)
	}
	kept, err := s.AdvancedSeats(context.Background(), "u-me", func(r *SeatRow) bool { return r.DeckID == "d1" })
	if err != nil || len(kept) != 1 || kept[0].DeckID != "d1" || len(kept[0].TurnRows) != 1 {
		t.Errorf("keep d1: %v %+v", err, kept)
	}
	if none, err := s.AdvancedSeats(context.Background(), "", nil); err != nil || len(none) != 0 {
		t.Errorf("no user: %v %v", none, err)
	}
}

// The read routes run these per request: none may scan a whole stats
// table. The person's seats come through seats_user_id, and each child
// table through its (game_id, seat) primary key.
func TestAdvancedSeatsQueriesUseIndexes(t *testing.T) {
	_, d := openStore(t)
	queries := map[string]string{"seats": advancedSeatsQuery}
	for _, child := range []string{"seat_turn_stats", "seat_target_stats", "seat_card_stats", "seat_card_seen"} {
		queries[child] = mineCTE + `SELECT t.game_id FROM mine JOIN ` + child + ` t ON t.game_id = mine.game_id AND t.seat = mine.seat`
	}
	for child, q := range queries {
		rows, err := d.Query(`EXPLAIN QUERY PLAN `+q, "u-me")
		if err != nil {
			t.Fatal(err)
		}
		var plan []string
		for rows.Next() {
			var id, parent, notused int
			var detail string
			if err := rows.Scan(&id, &parent, &notused, &detail); err != nil {
				t.Fatal(err)
			}
			plan = append(plan, detail)
		}
		_ = rows.Close()
		joined := strings.Join(plan, "; ")
		t.Logf("%s: %s", child, joined)
		for _, line := range plan {
			if strings.HasPrefix(strings.TrimSpace(line), "SCAN") {
				t.Errorf("%s: a full scan in the plan: %s", child, joined)
			}
		}
		if !strings.Contains(joined, "seats_user_id") {
			t.Errorf("%s: the plan does not start from seats_user_id: %s", child, joined)
		}
	}
}
