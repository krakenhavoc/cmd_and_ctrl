// Package gamestats derives a statistics record from a finished game
// (ADR 0165): per game, the turn and round counts; per seat, the result,
// place, elimination, life and damage totals, spells, draws, attacks,
// tokens, mulligans and commanders; per seat and card, how often it was
// cast.
//
// A record is computed, never entered: Compute reads the game's final
// omniscient view and its whole log (protocol.FullLogOf at game end, or
// a merge of the replay log's frames for an older game) and nothing
// else. Version names the definitions; when one changes, Version moves
// and old records are computed again. The record holds no names of
// people and no user ids: a seat's person is seats.user_id, through the
// join on (game_id, seat).
package gamestats

import (
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// Version is the version of the definitions below (ADR 0165 §1). Bump it
// when any of them changes, so the backfill computes older records again.
const Version = 1

// Source says what a record was derived from.
type Source string

const (
	// SourceLive: the engine's complete event history, when the game ended.
	SourceLive Source = "live"
	// SourceReplay: a merge of the replay log's frames (cmd/statsbackfill).
	SourceReplay Source = "replay"
)

// Result is a seat's result.
type Result string

const (
	ResultWin  Result = "win"
	ResultLoss Result = "loss"
	ResultDraw Result = "draw"
	// ResultNone: the game ended without an outcome (closed by its host
	// or an admin, or before it started).
	ResultNone Result = "none"
)

// Game is one game's record.
type Game struct {
	GameID      uuid.UUID
	Version     int
	Source      Source
	LogComplete bool
	// Turns is the last turn number in the log; 0 for a game that never
	// reached its first turn.
	Turns  int
	Rounds int
	// StartingSeat is the seat that took the first turn; nil when there
	// was none.
	StartingSeat *int
	ComputedAt   time.Time
	// Seats is one entry per seat, in seat order.
	Seats []Seat
}

// Seat is one seat's line of a game's record.
type Seat struct {
	Seat   int
	Result Result
	// Place is 1 for the winner, then by reverse elimination order;
	// seats still standing in a draw share 1. Nil when the game has no
	// outcome.
	Place          *int
	EliminatedTurn *int
	Conceded       bool

	FinalLife  int
	LifeGained int
	// LifeLost is positive: the sum of every loss, damage included.
	LifeLost int
	// DamageToPlayers counts damage this seat's sources dealt to other
	// players; CombatDamageToPlayers is the part dealt in combat.
	DamageToPlayers       int
	CombatDamageToPlayers int
	CommanderDamageDealt  int
	CommanderDamageTaken  int

	SpellsCast int
	CardsDrawn int
	// Attacks counts attacking creatures, one per declaration.
	Attacks       int
	TokensCreated int
	Mulligans     int
	// Commanders are the names of the commander cards the seat owns.
	Commanders []string
	// Cards is one entry per card name the seat cast, by name.
	Cards []Card
}

// Card is how often a seat cast one card.
type Card struct {
	Name string
	// ScryfallID is one printing's id, for a picture; "" when unknown.
	ScryfallID    string
	Casts         int
	FirstCastTurn int
}

// Input is what Compute reads.
type Input struct {
	GameID uuid.UUID
	// View is the game's final omniscient view (protocol.ViewOfGame).
	View protocol.GameView
	// Log is the game's whole log, oldest first.
	Log []protocol.LogEvent
	// Outcome is games.outcome: game.OutcomeWin, game.OutcomeDraw or "".
	Outcome string
	// WinnerSeat is games.winner_seat; nil for no single winner.
	WinnerSeat  *int
	Source      Source
	LogComplete bool
	Now         time.Time
}

// Compute derives a game's record. It never fails: a fact the inputs do
// not carry is left at its zero value.
func Compute(in Input) Game {
	v := &in.View
	out := Game{
		GameID:      in.GameID,
		Version:     Version,
		Source:      in.Source,
		LogComplete: in.LogComplete,
		ComputedAt:  in.Now.UTC(),
	}

	// Seats in seat order, and the lookups every rule below needs.
	players := append([]protocol.PlayerView(nil), v.Seats...)
	sort.Slice(players, func(i, j int) bool { return players[i].Seat < players[j].Seat })
	out.Seats = make([]Seat, len(players))
	at := make(map[int]*Seat, len(players))            // seat number → its line
	seatOfPlayer := make(map[string]int, len(players)) // player id → seat number
	for i, p := range players {
		out.Seats[i] = Seat{Seat: p.Seat, FinalLife: p.Life, Mulligans: p.MulligansTaken}
		at[p.Seat] = &out.Seats[i]
		seatOfPlayer[p.ID] = p.Seat
	}
	cards := cardsOf(v)
	// seatFor is the seat responsible for an entry: the one the log
	// names, or else the controller of the card it is about.
	seatFor := func(e protocol.LogEvent) (*Seat, int) {
		if s, ok := at[e.Seat]; ok {
			return s, e.Seat
		}
		if c, ok := cards[e.CardID]; ok {
			if n, ok := seatOfPlayer[c.Controller]; ok {
				return at[n], n
			}
		}
		return nil, protocol.NoSeat
	}

	var eliminated []int // seat numbers, in the order they left
	cast := make(map[int]map[string]*Card)
	for _, e := range in.Log {
		out.Turns = max(out.Turns, e.Turn)
		out.Rounds = max(out.Rounds, e.Round)
		switch e.Kind {
		case protocol.LogCast:
			s, n := seatFor(e)
			if s == nil {
				continue
			}
			s.SpellsCast++
			c, ok := cards[e.CardID]
			if !ok || c.Name == "" {
				continue
			}
			if cast[n] == nil {
				cast[n] = make(map[string]*Card)
			}
			if got := cast[n][c.Name]; got != nil {
				got.Casts++
				if got.ScryfallID == "" {
					got.ScryfallID = c.ScryfallID
				}
			} else {
				cast[n][c.Name] = &Card{Name: c.Name, ScryfallID: c.ScryfallID, Casts: 1, FirstCastTurn: e.Turn}
			}
		case protocol.LogDraw:
			if s, ok := at[e.Seat]; ok {
				s.CardsDrawn += max(e.Amount, 1)
			}
		case protocol.LogLife:
			if s, ok := at[e.Seat]; ok {
				if e.Amount > 0 {
					s.LifeGained += e.Amount
				} else {
					s.LifeLost -= e.Amount
				}
			}
		case protocol.LogDamage:
			if e.TargetSeat == nil {
				continue // damage to a permanent
			}
			s, n := seatFor(e)
			if s == nil || n == *e.TargetSeat {
				continue // damage to yourself is not damage to players
			}
			s.DamageToPlayers += e.Amount
			if e.Combat {
				s.CombatDamageToPlayers += e.Amount
			}
		case protocol.LogAttack:
			if s, _ := seatFor(e); s != nil {
				s.Attacks++
			}
		case protocol.LogToken:
			if s, _ := seatFor(e); s != nil {
				s.TokensCreated++
			}
		case protocol.LogEliminated:
			if s, ok := at[e.Seat]; ok && s.EliminatedTurn == nil {
				turn := e.Turn
				s.EliminatedTurn = &turn
				s.Conceded = e.Amount == 1
				eliminated = append(eliminated, e.Seat)
			}
		}
	}
	// A seat the view says is out but whose line the log lost (a replay
	// merge with a gap) still left, at an unknown turn, before anyone the
	// log saw leave.
	for _, p := range players {
		if p.Eliminated && at[p.Seat].EliminatedTurn == nil && !contains(eliminated, p.Seat) {
			eliminated = append([]int{p.Seat}, eliminated...)
		}
	}
	if out.Turns > 0 {
		start := v.StartingSeat
		out.StartingSeat = &start
	}

	// Commander damage: the view keeps what each commander dealt to each
	// player; the commander's owner dealt it.
	for _, p := range players {
		for id, dmg := range p.CommanderDamage {
			at[p.Seat].CommanderDamageTaken += dmg
			if c, ok := cards[id]; ok {
				if n, ok := seatOfPlayer[c.Owner]; ok && n != p.Seat {
					at[n].CommanderDamageDealt += dmg
				}
			}
		}
	}
	for _, c := range cards {
		if !c.IsCommander {
			continue
		}
		if n, ok := seatOfPlayer[c.Owner]; ok && !containsString(at[n].Commanders, c.Name) {
			at[n].Commanders = append(at[n].Commanders, c.Name)
		}
	}
	for i := range out.Seats {
		s := &out.Seats[i]
		sort.Strings(s.Commanders)
		if s.Commanders == nil {
			s.Commanders = []string{}
		}
		for _, c := range cast[s.Seat] {
			s.Cards = append(s.Cards, *c)
		}
		sort.Slice(s.Cards, func(a, b int) bool { return s.Cards[a].Name < s.Cards[b].Name })
	}

	placeAndResult(out.Seats, at, eliminated, in.Outcome, in.WinnerSeat)
	return out
}

// placeAndResult fills Result and Place (ADR 0165 §3).
func placeAndResult(seats []Seat, at map[int]*Seat, eliminated []int, outcome string, winner *int) {
	n := len(seats)
	out := make(map[int]bool, len(eliminated))
	for _, s := range eliminated {
		out[s] = true
	}
	switch {
	case outcome == "win" && winner != nil && at[*winner] != nil:
		for i := range seats {
			seats[i].Result = ResultLoss
		}
		at[*winner].Result = ResultWin
		at[*winner].Place = ptr(1)
		// Standing but not the winner (an alternate win): second, shared.
		for i := range seats {
			if s := &seats[i]; s.Seat != *winner && !out[s.Seat] {
				s.Place = ptr(2)
			}
		}
	case outcome == "draw":
		for i := range seats {
			if s := &seats[i]; !out[s.Seat] {
				s.Result = ResultDraw
				s.Place = ptr(1)
			} else {
				s.Result = ResultLoss
			}
		}
	default:
		for i := range seats {
			seats[i].Result = ResultNone
		}
		return
	}
	// The first seat out places last, the next one place above it.
	for i, s := range eliminated {
		if at[s] != nil && at[s].Result != ResultWin {
			at[s].Place = ptr(n - i)
		}
	}
}

// cardsOf indexes every card the view holds by instance id.
func cardsOf(v *protocol.GameView) map[string]protocol.CardView {
	out := make(map[string]protocol.CardView)
	add := func(z protocol.ZoneView) {
		for _, c := range z.Cards {
			out[c.InstanceID] = c
		}
	}
	add(v.Battlefield)
	add(v.Stack)
	add(v.Exile)
	add(v.PhasedOut)
	for _, p := range v.Seats {
		add(p.Library)
		add(p.Hand)
		add(p.Graveyard)
		add(p.Command)
	}
	return out
}

func ptr(n int) *int { return &n }

func contains(xs []int, x int) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

func containsString(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}
