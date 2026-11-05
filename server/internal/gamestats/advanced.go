package gamestats

// advanced.go folds a person's SeatRows into ADR 0217 §3 and §4's
// numbers (slice 5 of #4880). Pure: AdvancedSeats reads the rows and
// the lobby adds what comes from deck lists.
//
// Denominators. Every stat counts only the seats that have its
// columns: a version-1 record has none of the version-2 ones, and a
// game whose replay was not read has none of the frame ones. Coverage
// says how many seats each kind of stat stands on.

import (
	"math"
	"sort"
	"strings"
	"time"
)

// MinGames is ADR 0217 §3's small-sample rule: no rate is given on
// fewer games than this.
const MinGames = 3

// MakeupFilter is ADR 0217 §3's table makeup filter, read from the seats.
type MakeupFilter string

const (
	// MakeupAll is every game (the default).
	MakeupAll MakeupFilter = "all"
	// MakeupPeople: every other seat is a person (guests count).
	MakeupPeople MakeupFilter = "people"
	// MakeupWithBots: at least one other seat is a bot and one a person.
	MakeupWithBots MakeupFilter = "with_bots"
	// MakeupBots: every other seat is a bot.
	MakeupBots MakeupFilter = "bots"
)

// ParseMakeup reads the makeup query value; "" is MakeupAll.
func ParseMakeup(s string) (MakeupFilter, bool) {
	switch MakeupFilter(s) {
	case "", MakeupAll:
		return MakeupAll, true
	case MakeupPeople, MakeupWithBots, MakeupBots:
		return MakeupFilter(s), true
	}
	return "", false
}

// Matches reports whether a seat's game has this makeup.
func (m MakeupFilter) Matches(r *SeatRow) bool {
	switch m {
	case MakeupPeople:
		return r.OtherBots == 0
	case MakeupWithBots:
		return r.OtherBots > 0 && r.OtherPeople > 0
	case MakeupBots:
		return r.OtherBots > 0 && r.OtherPeople == 0
	}
	return true
}

// Range is ADR 0217 §3's date range.
type Range string

const (
	RangeAll Range = "all"
	Range90  Range = "90d"
	Range30  Range = "30d"
)

// ParseRange reads the range query value; "" is RangeAll.
func ParseRange(s string) (Range, bool) {
	switch Range(s) {
	case "", RangeAll:
		return RangeAll, true
	case Range90, Range30:
		return Range(s), true
	}
	return "", false
}

// Since is the earliest end time (Unix ms) the range keeps, 0 for all.
func (r Range) Since(now time.Time) int64 {
	switch r {
	case Range90:
		return now.Add(-90 * 24 * time.Hour).UnixMilli()
	case Range30:
		return now.Add(-30 * 24 * time.Hour).UnixMilli()
	}
	return 0
}

// Rate is a count out of a total. Rate is nil when Of is below
// MinGames, so a page cannot show a rate on too few games.
type Rate struct {
	N    int      `json:"n"`
	Of   int      `json:"of"`
	Rate *float64 `json:"rate"`
}

// RateOf is n out of of, with the rate left out below MinGames.
func RateOf(n, of int) Rate {
	r := Rate{N: n, Of: of}
	if of >= MinGames {
		v := float64(n) / float64(of)
		r.Rate = &v
	}
	return r
}

// SummaryCoverage counts the seats each kind of stat stands on: Games every
// counted seat, V2 those with a version-2 record, Frames those whose
// replay was read whole. StandIn is the games a stand-in bot played
// part of (ADR 0177).
type SummaryCoverage struct {
	Games   int `json:"games"`
	V2      int `json:"v2"`
	Frames  int `json:"frames"`
	StandIn int `json:"stand_in"`
}

// PlaceRow is the finishing places at one table size: Places[0] is the
// games finished 1st. Sizes are never mixed (ADR 0217 §3).
type PlaceRow struct {
	Players int   `json:"players"`
	Games   int   `json:"games"`
	Places  []int `json:"places"`
}

// TurnOrderRow is the first-turn effect at one turn position, for
// four-seat games.
type TurnOrderRow struct {
	Position int      `json:"position"`
	Wins     Rate     `json:"wins"`
	AvgPlace *float64 `json:"avg_place"`
}

// Spread is a median and the quartiles around it.
type Spread struct {
	Median *float64 `json:"median"`
	P25    *float64 `json:"p25"`
	P75    *float64 `json:"p75"`
}

// Bucket is one bar of a histogram.
type Bucket struct {
	Value int `json:"value"`
	Games int `json:"games"`
}

// Length is game length: turns, rounds and minutes (ADR 0217 §3), as
// medians and spreads, and a histogram of turns.
type Length struct {
	Turns   Spread   `json:"turns"`
	Rounds  Spread   `json:"rounds"`
	Minutes Spread   `json:"minutes"`
	ByTurns []Bucket `json:"by_turns"`
}

// MulliganRow is the win rate after a number of mulligans; "2+" pools
// two and more.
type MulliganRow struct {
	Mulligans string `json:"mulligans"`
	Wins      Rate   `json:"wins"`
}

// Mulligans is mulligans per game and the win rate after each.
type Mulligans struct {
	PerGame *float64      `json:"per_game"`
	Rows    []MulliganRow `json:"rows"`
}

// CommanderRow is one commander (pair) of the person's: its record and
// how its wins came.
type CommanderRow struct {
	Name     string      `json:"name"`
	Tally    Tally       `json:"record"`
	WinKinds []KindCount `json:"win_kinds"`
}

// CommanderStats is ADR 0217 §3's commander numbers.
type CommanderStats struct {
	// CastsPerGame is over the version-2 seats.
	CastsPerGame *float64 `json:"casts_per_game"`
	// FirstCastTurn is the game turn of the first commander cast, over
	// the seats that cast one.
	FirstCastTurn Spread `json:"first_cast_turn"`
	// ByTurn5 is the win rate when the commander was first cast by
	// turn 5, Later when later or never.
	ByTurn5 Rate `json:"by_turn_5"`
	Later   Rate `json:"later"`
	// ByCommander is filled on the overview only.
	ByCommander []CommanderRow `json:"by_commander,omitempty"`
}

// OverTurn is attacks and damage at one own-turn number; OverTurnTop
// pools that number and later.
type OverTurn struct {
	OwnTurn        int      `json:"own_turn"`
	Turns          int      `json:"turns"`
	AttacksPerTurn *float64 `json:"attacks_per_turn"`
	DamagePerTurn  *float64 `json:"damage_per_turn"`
}

// OverTurnTop is the last own-turn bucket: 12 and later.
const OverTurnTop = 12

// Aggression is ADR 0217 §3's aggression numbers. The per-turn ones
// are over own turns with frames.
type Aggression struct {
	OwnTurns       int      `json:"own_turns"`
	AttacksPerTurn *float64 `json:"attacks_per_turn"`
	DamagePerTurn  *float64 `json:"damage_per_turn"`
	// AttackShare is creatures that attacked over creatures that could
	// have, summed over the seats with frames.
	AttackShare *float64 `json:"attack_share"`
	// MostHitShare is the mean, over games with three or more seats
	// where the seat dealt damage, of the share the most-hit opponent
	// took; Evenness the mean of 0 (all at one) to 1 (even). Spread
	// games is how many games they stand on. Opponents are never named.
	MostHitShare *float64   `json:"most_hit_share"`
	Evenness     *float64   `json:"evenness"`
	SpreadGames  int        `json:"spread_games"`
	OverTurns    []OverTurn `json:"over_turns"`
}

// LandMiss is the missed drops for one land number; LandMissTop pools
// that number and up.
type LandMiss struct {
	Land   int `json:"land"`
	Missed int `json:"missed"`
}

// LandMissTop is the last land-number bucket: 6 and up.
const LandMissTop = 6

// LandCurvePoint is ADR 0217 §4's land-drop reliability at one N: of
// the games that reached own turn N, the share with N land drops by
// then (Hit) and with N lands on the battlefield by then (Controlled,
// which counts ramp). LandCurveTop pools N = 7 and up, counting each
// game once per N it reached.
type LandCurvePoint struct {
	N          int  `json:"n"`
	Hit        Rate `json:"hit"`
	Controlled Rate `json:"controlled"`
}

// LandCurveTop is the last N of the land-drop curve: 7 and up.
const LandCurveTop = 7

// ManaGame is one recent game's mana troubles, for the trend.
type ManaGame struct {
	Screw       bool `json:"screw"`
	Flood       bool `json:"flood"`
	ColourScrew bool `json:"colour_screw"`
}

// TrendMax is how many recent games the mana trend lists.
const TrendMax = 20

// Mana is ADR 0217 §4's numbers.
type Mana struct {
	LandDropsHit    int        `json:"land_drops_hit"`
	LandDropsMissed int        `json:"land_drops_missed"`
	MissedByLand    []LandMiss `json:"missed_by_land"`
	// Screw, Flood and ColourScrew are the screwed, flooded and
	// colour-screwed games.
	Screw       Rate `json:"screw"`
	Flood       Rate `json:"flood"`
	ColourScrew Rate `json:"colour_screw"`
	// Trend is the last TrendMax games with frames, newest first.
	Trend     []ManaGame       `json:"trend"`
	LandCurve []LandCurvePoint `json:"land_curve"`
}

// CastCard is a card the person cast: in how many games, and how often.
type CastCard struct {
	Name  string `json:"name"`
	Games int    `json:"games"`
	Casts int    `json:"casts"`
	// ScryfallID is a printing to draw the card with, filled in by the
	// read route from the card index; empty when the name is unknown.
	ScryfallID string `json:"scryfall_id,omitempty"`
}

// Summary is what both read routes show for a set of seats.
type Summary struct {
	Coverage   SummaryCoverage `json:"coverage"`
	Record     Tally           `json:"record"`
	WinRate    Rate            `json:"win_rate"`
	Places     []PlaceRow      `json:"places"`
	WinKinds   []KindCount     `json:"win_kinds"`
	ComboRules []KindCount     `json:"combo_rules"`
	// OutKinds is how the person's seats left: how they lose.
	OutKinds   []KindCount    `json:"out_kinds"`
	TurnOrder  []TurnOrderRow `json:"turn_order"`
	Length     Length         `json:"length"`
	Mulligans  Mulligans      `json:"mulligans"`
	Commander  CommanderStats `json:"commander"`
	Aggression Aggression     `json:"aggression"`
	Mana       Mana           `json:"mana"`
	MostCast   []CastCard     `json:"most_cast"`
}

// Summarise folds seats (newest first) into a Summary.
func Summarise(seats []SeatRow) Summary {
	out := Summary{Places: []PlaceRow{}, WinKinds: []KindCount{}, ComboRules: []KindCount{},
		OutKinds: []KindCount{}, TurnOrder: []TurnOrderRow{}, MostCast: []CastCard{}}
	var turns, rounds, minutes, firstCast []float64
	places := map[int]*PlaceRow{}
	winKinds, outKinds, combo := map[string]int{}, map[string]int{}, map[string]int{}
	type posAcc struct{ games, wins, placeSum, placed int }
	pos := map[int]*posAcc{}
	byTurns := map[int]int{}
	mullSum := 0
	mull := [3][2]int{} // [0, 1, 2+] × [games, wins]
	var cmdCasts, cmdSeats int
	var by5, by5Wins, later, laterWins int
	var ag struct{ attackers, eligible, damage, attacked, couldAttack, ownTurns int }
	var spreadShare, spreadEven float64
	over := map[int]*[3]int{} // own turn → turns, attackers, damage
	missedBy := map[int]int{}
	var screw, flood, colour [2]int // n, of
	type curveAcc struct{ of, hit, ctl int }
	curve := map[int]*curveAcc{}
	type castAcc struct{ games, casts int }
	cast := map[string]*castAcc{}
	trend := []ManaGame{}

	for i := range seats {
		r := &seats[i]
		out.Coverage.Games++
		if r.StatsVersion >= 2 {
			out.Coverage.V2++
		}
		frames := r.FramesComplete != nil && *r.FramesComplete
		if frames {
			out.Coverage.Frames++
		}
		if r.StandIn {
			out.Coverage.StandIn++
		}
		out.Record.add(r.Result, r.Place)
		won := r.Result == ResultWin

		if r.Place != nil && r.Players > 0 && *r.Place >= 1 && *r.Place <= r.Players {
			pr := places[r.Players]
			if pr == nil {
				pr = &PlaceRow{Players: r.Players, Places: make([]int, r.Players)}
				places[r.Players] = pr
			}
			pr.Games++
			pr.Places[*r.Place-1]++
		}
		if won && r.WinKind != "" {
			winKinds[string(r.WinKind)]++
			if r.WinKind == KindCombo && r.ComboRule != "" {
				combo[string(r.ComboRule)]++
			}
		}
		if r.OutKind != "" {
			outKinds[string(r.OutKind)]++
		}
		if r.Players == 4 && r.TurnPosition != nil && *r.TurnPosition >= 1 && *r.TurnPosition <= 4 {
			a := pos[*r.TurnPosition]
			if a == nil {
				a = &posAcc{}
				pos[*r.TurnPosition] = a
			}
			a.games++
			if won {
				a.wins++
			}
			if r.Place != nil {
				a.placeSum += *r.Place
				a.placed++
			}
		}
		turns = append(turns, float64(r.Turns))
		rounds = append(rounds, float64(r.Rounds))
		byTurns[r.Turns]++
		if r.StartedAt > 0 && r.EndedAt > r.StartedAt {
			minutes = append(minutes, float64(r.EndedAt-r.StartedAt)/60000)
		}
		mullSum += r.Mulligans
		m := min(r.Mulligans, 2)
		mull[m][0]++
		if won {
			mull[m][1]++
		}
		if r.CommanderCasts != nil {
			cmdSeats++
			cmdCasts += *r.CommanderCasts
			if r.CommanderFirstCastTurn != nil && *r.CommanderFirstCastTurn <= 5 {
				by5++
				if won {
					by5Wins++
				}
			} else {
				later++
				if won {
					laterWins++
				}
			}
		}
		if r.CommanderFirstCastTurn != nil {
			firstCast = append(firstCast, float64(*r.CommanderFirstCastTurn))
		}
		for _, c := range r.Cast {
			a := cast[c.Name]
			if a == nil {
				a = &castAcc{}
				cast[c.Name] = a
			}
			a.games++
			a.casts += c.Casts
		}
		if !frames {
			continue
		}

		// Frame-derived: aggression, land drops, screw and flood.
		if r.CreaturesEligible != nil && r.CreaturesAttacked != nil {
			ag.couldAttack += *r.CreaturesEligible
			ag.attacked += *r.CreaturesAttacked
		}
		if r.LandDropsHit != nil {
			out.Mana.LandDropsHit += *r.LandDropsHit
		}
		if r.LandDropsMissed != nil {
			out.Mana.LandDropsMissed += *r.LandDropsMissed
		}
		count := func(acc *[2]int, flag *bool) bool {
			if flag == nil {
				return false
			}
			acc[1]++
			if *flag {
				acc[0]++
			}
			return *flag
		}
		g := ManaGame{Screw: count(&screw, r.ScrewGame), Flood: count(&flood, r.FloodGame),
			ColourScrew: count(&colour, r.ColourScrewGame)}
		if len(trend) < TrendMax {
			trend = append(trend, g)
		}
		reached := map[int]TurnRow{}
		for _, t := range r.TurnRows {
			ag.ownTurns++
			ag.attackers += t.Attackers
			ag.damage += t.DamageToPlayers
			o := over[min(t.OwnTurn, OverTurnTop)]
			if o == nil {
				o = &[3]int{}
				over[min(t.OwnTurn, OverTurnTop)] = o
			}
			o[0]++
			o[1] += t.Attackers
			o[2] += t.DamageToPlayers
			if t.Missed {
				missedBy[min(t.LandDrops+1, LandMissTop)]++
			}
			reached[t.OwnTurn] = t
		}
		for n, t := range reached {
			if n < 1 {
				continue
			}
			c := curve[min(n, LandCurveTop)]
			if c == nil {
				c = &curveAcc{}
				curve[min(n, LandCurveTop)] = c
			}
			c.of++
			if t.LandDrops >= n {
				c.hit++
			}
			// Lands at the start of main, plus the land played that
			// turn: the lands on the battlefield by the end of own turn n.
			if t.LandsControlled+t.LandPlayed >= n {
				c.ctl++
			}
		}
		if share, even, ok := spreadOf(r.Targets, r.Players-1); ok {
			out.Aggression.SpreadGames++
			spreadShare += share
			spreadEven += even
		}
	}
	out.Record.finish()
	out.WinRate = RateOf(out.Record.Wins, out.Record.Games)

	for _, n := range sortedKeys(places) {
		out.Places = append(out.Places, *places[n])
	}
	for _, k := range Kinds {
		if n := winKinds[string(k)]; n > 0 {
			out.WinKinds = append(out.WinKinds, KindCount{Kind: string(k), Games: n})
		}
		if n := outKinds[string(k)]; n > 0 {
			out.OutKinds = append(out.OutKinds, KindCount{Kind: string(k), Games: n})
		}
	}
	for _, k := range []ComboRule{ComboLoop, ComboRepeats, ComboActions, ComboCombats} {
		if n := combo[string(k)]; n > 0 {
			out.ComboRules = append(out.ComboRules, KindCount{Kind: string(k), Games: n})
		}
	}
	for p := 1; p <= 4; p++ {
		a := pos[p]
		if a == nil {
			continue
		}
		row := TurnOrderRow{Position: p, Wins: RateOf(a.wins, a.games)}
		if a.placed > 0 {
			v := float64(a.placeSum) / float64(a.placed)
			row.AvgPlace = &v
		}
		out.TurnOrder = append(out.TurnOrder, row)
	}
	out.Length = Length{Turns: spreadOf3(turns), Rounds: spreadOf3(rounds), Minutes: spreadOf3(minutes), ByTurns: []Bucket{}}
	for _, t := range sortedKeys(byTurns) {
		out.Length.ByTurns = append(out.Length.ByTurns, Bucket{Value: t, Games: byTurns[t]})
	}
	out.Mulligans.PerGame = mean(mullSum, out.Coverage.Games)
	for i, label := range []string{"0", "1", "2+"} {
		out.Mulligans.Rows = append(out.Mulligans.Rows, MulliganRow{Mulligans: label, Wins: RateOf(mull[i][1], mull[i][0])})
	}
	out.Commander = CommanderStats{CastsPerGame: mean(cmdCasts, cmdSeats), FirstCastTurn: spreadOf3(firstCast),
		ByTurn5: RateOf(by5Wins, by5), Later: RateOf(laterWins, later)}

	out.Aggression.OwnTurns = ag.ownTurns
	out.Aggression.AttacksPerTurn = mean(ag.attackers, ag.ownTurns)
	out.Aggression.DamagePerTurn = mean(ag.damage, ag.ownTurns)
	out.Aggression.AttackShare = mean(ag.attacked, ag.couldAttack)
	if n := out.Aggression.SpreadGames; n > 0 {
		s, e := spreadShare/float64(n), spreadEven/float64(n)
		out.Aggression.MostHitShare, out.Aggression.Evenness = &s, &e
	}
	out.Aggression.OverTurns = []OverTurn{}
	for t := 1; t <= OverTurnTop; t++ {
		o := over[t]
		if o == nil {
			continue
		}
		out.Aggression.OverTurns = append(out.Aggression.OverTurns, OverTurn{OwnTurn: t, Turns: o[0],
			AttacksPerTurn: mean(o[1], o[0]), DamagePerTurn: mean(o[2], o[0])})
	}

	out.Mana.MissedByLand = make([]LandMiss, LandMissTop)
	for l := 1; l <= LandMissTop; l++ {
		out.Mana.MissedByLand[l-1] = LandMiss{Land: l, Missed: missedBy[l]}
	}
	out.Mana.Screw, out.Mana.Flood, out.Mana.ColourScrew = RateOf(screw[0], screw[1]), RateOf(flood[0], flood[1]),
		RateOf(colour[0], colour[1])
	out.Mana.Trend = trend
	out.Mana.LandCurve = make([]LandCurvePoint, LandCurveTop)
	for n := 1; n <= LandCurveTop; n++ {
		c := curve[n]
		if c == nil {
			c = &curveAcc{}
		}
		out.Mana.LandCurve[n-1] = LandCurvePoint{N: n, Hit: RateOf(c.hit, c.of), Controlled: RateOf(c.ctl, c.of)}
	}

	for name, a := range cast {
		out.MostCast = append(out.MostCast, CastCard{Name: name, Games: a.games, Casts: a.casts})
	}
	sort.Slice(out.MostCast, func(i, j int) bool {
		a, b := out.MostCast[i], out.MostCast[j]
		if a.Games != b.Games {
			return a.Games > b.Games
		}
		if a.Casts != b.Casts {
			return a.Casts > b.Casts
		}
		return a.Name < b.Name
	})
	if len(out.MostCast) > TopCardsMax {
		out.MostCast = out.MostCast[:TopCardsMax]
	}
	return out
}

// ByCommander groups seats by commander (pair) as ADR 0165 §7 does,
// most-played first, with each one's win kinds.
func ByCommander(seats []SeatRow) []CommanderRow {
	type acc struct {
		row   CommanderRow
		kinds map[string]int
	}
	m := map[string]*acc{}
	var order []string
	for i := range seats {
		r := &seats[i]
		if len(r.Commanders) == 0 {
			continue
		}
		name := strings.Join(r.Commanders, " / ")
		a := m[name]
		if a == nil {
			a = &acc{row: CommanderRow{Name: name}, kinds: map[string]int{}}
			m[name] = a
			order = append(order, name)
		}
		a.row.Tally.add(r.Result, r.Place)
		if r.Result == ResultWin && r.WinKind != "" {
			a.kinds[string(r.WinKind)]++
		}
	}
	out := make([]CommanderRow, 0, len(order))
	for _, name := range order {
		a := m[name]
		a.row.Tally.finish()
		a.row.WinKinds = []KindCount{}
		for _, k := range Kinds {
			if n := a.kinds[string(k)]; n > 0 {
				a.row.WinKinds = append(a.row.WinKinds, KindCount{Kind: string(k), Games: n})
			}
		}
		out = append(out, a.row)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Tally.Games > out[j].Tally.Games })
	return out
}

// DeadCard is ADR 0217 §3's dead card: a non-land card that was in
// hand in a game and never cast in it. Dead is the games it was dead
// out of the games it was in hand. CouldAfford is true when it was
// affordable on some turn in at least half of its dead games.
type DeadCard struct {
	Name        string `json:"name"`
	Dead        Rate   `json:"dead"`
	CouldAfford bool   `json:"could_afford"`
}

// DeadCardsMax caps the dead-card list.
const DeadCardsMax = 24

// DeadCards lists the dead cards over seats, most games dead first.
// isLand says which names are lands, which are never dead cards; nil
// treats every name as a spell.
func DeadCards(seats []SeatRow, isLand func(name string) bool) []DeadCard {
	type acc struct{ held, dead, affordable int }
	m := map[string]*acc{}
	landCache := map[string]bool{}
	for i := range seats {
		for _, s := range seats[i].Seen {
			land, ok := landCache[s.Name]
			if !ok {
				land = isLand != nil && isLand(s.Name)
				landCache[s.Name] = land
			}
			if land {
				continue
			}
			a := m[s.Name]
			if a == nil {
				a = &acc{}
				m[s.Name] = a
			}
			a.held++
			if !s.Cast {
				a.dead++
				if s.AffordableTurns > 0 {
					a.affordable++
				}
			}
		}
	}
	out := []DeadCard{}
	for name, a := range m {
		if a.dead == 0 {
			continue
		}
		out = append(out, DeadCard{Name: name, Dead: RateOf(a.dead, a.held), CouldAfford: 2*a.affordable >= a.dead})
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Dead.N != b.Dead.N {
			return a.Dead.N > b.Dead.N
		}
		if a.Dead.Of != b.Dead.Of {
			return a.Dead.Of < b.Dead.Of
		}
		return a.Name < b.Name
	})
	if len(out) > DeadCardsMax {
		out = out[:DeadCardsMax]
	}
	return out
}

// spreadOf is one game's spread of damage over its opponents: the
// most-hit opponent's share and an evenness from 0 (all at one) to 1
// (even), the normalised entropy over every opponent. ok is false with
// fewer than two opponents or no damage.
func spreadOf(targets []TargetRow, opponents int) (share, evenness float64, ok bool) {
	if opponents < 2 {
		return 0, 0, false
	}
	total, top := 0, 0
	for _, t := range targets {
		total += t.Damage
		top = max(top, t.Damage)
	}
	if total <= 0 {
		return 0, 0, false
	}
	h := 0.0
	for _, t := range targets {
		if t.Damage > 0 {
			p := float64(t.Damage) / float64(total)
			h -= p * math.Log(p)
		}
	}
	return float64(top) / float64(total), h / math.Log(float64(opponents)), true
}

// spreadOf3 is the median and quartiles of xs (linear interpolation),
// all nil for an empty list.
func spreadOf3(xs []float64) Spread {
	if len(xs) == 0 {
		return Spread{}
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	q := func(p float64) *float64 {
		at := p * float64(len(s)-1)
		lo := int(math.Floor(at))
		hi := min(lo+1, len(s)-1)
		v := s[lo] + (s[hi]-s[lo])*(at-float64(lo))
		return &v
	}
	return Spread{Median: q(0.5), P25: q(0.25), P75: q(0.75)}
}

func mean(sum, n int) *float64 {
	if n <= 0 {
		return nil
	}
	v := float64(sum) / float64(n)
	return &v
}

func sortedKeys[V any](m map[int]V) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
