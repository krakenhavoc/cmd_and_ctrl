package protocol

import (
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// match_view.go is GameView.match (ADR 0222 §3): the score of a table's
// match, and the sideboarding window between two of its games.
//
// A table with no match (every Commander table) has no `match` key at
// all. A best-of-one table has one, with best_of 1, so a client has one
// thing to read; it shows a score only for a longer match.
//
// WHAT IS HIDDEN. A player's deck and sideboard between games are
// theirs alone: FilterViewFor keeps each pool for its own seat and
// drops the rest. Who plays first is the chooser's to decide until the
// window closes, and the opponent sideboards without knowing it, so the
// choice as it stands reaches the chooser alone.

// MatchView is GameView.match.
type MatchView struct {
	// BestOf is 1 or 3.
	BestOf int `json:"best_of"`
	// Game is the number of the game in play, or of the last game
	// played while the table is between games or the match is over.
	Game int `json:"game"`
	// Wins is each seat's games won, by seat index.
	Wins []int `json:"wins"`
	// Games is every finished game, in order.
	Games []MatchGameView `json:"games,omitempty"`
	// Over says the match is decided, drawn or conceded; WinnerSeat is
	// the seat that won it (absent for a draw) and ConcededSeat the
	// seat that conceded it, if one did.
	Over         bool `json:"over,omitempty"`
	WinnerSeat   *int `json:"winner_seat,omitempty"`
	ConcededSeat *int `json:"conceded_seat,omitempty"`
	// Sideboarding is the open window between two games. While it is
	// present the board is the finished game's and no game action is
	// legal.
	Sideboarding *SideboardingView `json:"sideboarding,omitempty"`
}

// MatchGameView is one finished game of a match.
type MatchGameView struct {
	// WinnerSeat is absent for a drawn game.
	WinnerSeat *int `json:"winner_seat,omitempty"`
	// Cause is the game outcome's cause (GameView.outcome.cause).
	Cause string `json:"cause,omitempty"`
	// StartingSeat took the first turn, and Turns is the turn number
	// the game ended on.
	StartingSeat int `json:"starting_seat"`
	Turns        int `json:"turns"`
}

// SideboardingView is the window between two games of a match.
type SideboardingView struct {
	// Deadline is when the next game begins whoever is ready, RFC 3339
	// UTC. SecondsLeft is the same measured by the server's clock when
	// the view was built, for a client whose own clock is off.
	Deadline    string `json:"deadline"`
	SecondsLeft int    `json:"seconds_left"`
	// Ready is the seats that are ready, a bot seat always among them.
	Ready []int `json:"ready"`
	// ChooserSeat chooses who plays first (CR 103.1: the loser of the
	// previous game).
	ChooserSeat int `json:"chooser_seat"`
	// PlayFirst is the chooser's choice as it stands. The chooser's
	// frame and the admin's only.
	PlayFirst *bool `json:"play_first,omitempty"`
	// DeckMin and SideboardMax are the format's deck check, for the
	// screen's counts: a deck of fewer cards, or a sideboard of more,
	// is refused.
	DeckMin      int `json:"deck_min"`
	SideboardMax int `json:"sideboard_max"`
	// Pools is each seat's deck and sideboard as they stand. The full
	// view carries every seat's; a player's frame carries their own,
	// and a spectator's none.
	Pools []MatchPoolView `json:"pools,omitempty"`
}

// MatchPoolView is one seat's deck and sideboard between games. A
// card's id is its pool id, which set_match_deck takes.
type MatchPoolView struct {
	Seat      int        `json:"seat"`
	PlayerID  string     `json:"player_id"`
	Deck      []CardView `json:"deck"`
	Sideboard []CardView `json:"sideboard"`
	// Registered is the pool ids of the deck as it was registered, so
	// the screen can mark what changed and offer the way back.
	Registered []string `json:"registered"`
}

// viewOfMatch projects Game.Match. Caller holds the game's read lock.
func viewOfMatch(g *game.Game, ids *idStrings) *MatchView {
	m := g.Match
	if m == nil {
		return nil
	}
	seatOf := func(id [16]byte) *int {
		for _, p := range g.Seats {
			if p != nil && p.ID == id {
				seat := p.Seat
				return &seat
			}
		}
		return nil
	}
	out := &MatchView{
		BestOf: m.BestOf,
		Game:   m.GameNumber(),
		Wins:   make([]int, len(g.Seats)),
		Over:   m.Over,
	}
	for i, p := range g.Seats {
		if p != nil {
			out.Wins[i] = m.Wins(p.ID)
		}
	}
	for _, mg := range m.Games {
		gv := MatchGameView{Cause: mg.Cause, StartingSeat: mg.StartingSeat, Turns: mg.Turns}
		if mg.Kind == game.OutcomeWin {
			gv.WinnerSeat = seatOf(mg.Winner)
		}
		out.Games = append(out.Games, gv)
	}
	if m.Over {
		out.WinnerSeat = seatOf(m.Winner)
	}
	out.ConcededSeat = seatOf(m.ConcededBy)

	b := m.Between
	if b == nil || g.State != game.StateBetweenGames {
		return out
	}
	rules := g.Settings.FormatSpec().Deck
	sb := &SideboardingView{
		Deadline:     b.Deadline.UTC().Format(time.RFC3339),
		SecondsLeft:  max(int(time.Until(b.Deadline).Round(time.Second)/time.Second), 0),
		Ready:        []int{},
		DeckMin:      rules.MinCards,
		SideboardMax: rules.SideboardSize,
	}
	if seat := seatOf(b.Chooser); seat != nil {
		sb.ChooserSeat = *seat
	}
	playFirst := b.PlayFirst
	sb.PlayFirst = &playFirst
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		ready := p.IsBot || p.Agent
		for _, id := range b.Ready {
			ready = ready || id == p.ID
		}
		if ready {
			sb.Ready = append(sb.Ready, p.Seat)
		}
		main, side, registered := g.MatchPoolLocked(p.ID)
		if main == nil && side == nil {
			continue
		}
		pool := MatchPoolView{
			Seat:       p.Seat,
			PlayerID:   ids.of(p.ID),
			Deck:       make([]CardView, len(main)),
			Sideboard:  make([]CardView, len(side)),
			Registered: make([]string, len(registered)),
		}
		for i := range main {
			pool.Deck[i] = viewOfCardPtrIDs(&main[i], ids)
		}
		for i := range side {
			pool.Sideboard[i] = viewOfCardPtrIDs(&side[i], ids)
		}
		for i, id := range registered {
			pool.Registered[i] = id.String()
		}
		sb.Pools = append(sb.Pools, pool)
	}
	out.Sideboarding = sb
	return out
}

// filterMatchFor is the per-viewer copy: every pool but the viewer's
// own is dropped, and the play-or-draw choice stays with the chooser.
// The admin's full view ("") keeps everything; a spectator sees the
// score, the timer and who is ready.
func filterMatchFor(m *MatchView, viewerID string) *MatchView {
	if m == nil {
		return nil
	}
	out := *m
	out.Wins = append([]int(nil), m.Wins...)
	out.Games = append([]MatchGameView(nil), m.Games...)
	if m.Sideboarding == nil {
		return &out
	}
	sb := *m.Sideboarding
	sb.Ready = append([]int{}, m.Sideboarding.Ready...)
	if viewerID != "" {
		sb.Pools = nil
		chooser := false
		for _, p := range m.Sideboarding.Pools {
			if p.PlayerID == viewerID && viewerID != SpectatorViewerID {
				sb.Pools = []MatchPoolView{p}
				chooser = p.Seat == sb.ChooserSeat
			}
		}
		if !chooser {
			sb.PlayFirst = nil
		}
	}
	out.Sideboarding = &sb
	return &out
}
