package game

import (
	"errors"
	"slices"

	"github.com/google/uuid"
)

// draw_limit.go — "can't draw more than one card each turn" and "can't
// draw cards", read at one gate (#3927).
//
// THE RULE. CR 121.2b: "Some effects say that a player can't draw more
// than one card each turn. Such an effect applies to individual card
// draws. Instructions to draw multiple cards may still be partially
// carried out. However, if an effect offers the player a choice to draw
// multiple cards, the affected player can't choose to do so. Similarly,
// the player can't pay a cost that includes drawing multiple cards." CR
// 121.3: a player who can't draw can't choose to draw. CR 614.17: a
// "can't" is not a replacement effect; CR 101.3: the part of an
// instruction that is impossible is ignored.
//
// A RESTRICTION, NOT A REPLACEMENT. It is asked for each individual draw
// BEFORE that draw's CR 614 window opens (drawEventLocked). A draw the
// player can't make is not performed: no replacement sees it (the
// rulings: dredge, Laboratory Maniac and Underrealm Lich can't replace a
// draw Narset disallows), no card moves, no EventDrawCard is emitted, so
// no "whenever a player draws" triggers and the log names no draw, and
// an empty library is not drawn from (no CR 704.5b loss). The rest of the
// instruction goes on: "draw three cards" with one draw allowed draws one
// and skips two, and the "then" of "draw two, then discard a card" still
// runs.
//
// THE COUNT is the cards the player has actually drawn this turn,
// Game.DrawnThisTurn, which only actuallyDrawCardLocked appends to. So a
// draw a replacement turned into something else was not a draw and does
// not count (the Narset and Leovold rulings), and draws made before the
// permanent arrived DO count (CR 614.17a: the effect cannot undo them,
// but the second draw of the turn after Narset enters is still refused).
// "Each turn" is every player's turn: the tally is cleared as each turn
// begins.
//
// THE SOURCES. Battlefield statics only, read live through
// CatalogAbilityKey, so a permanent that has lost its abilities (CR
// 613.1f) or has left gives nothing. Nothing is stored, so nothing is
// snapshotted. "Can't draw cards" (Maralen of the Mornsong) is the same
// static with a cap of zero; the lowest cap that reaches a player wins.

// DrawLimitWhose is which players a printed draw limit covers, relative
// to the controller of the permanent that has it.
type DrawLimitWhose uint8

const (
	// DrawLimitEveryone is "Each player can't draw more than …" and
	// "Players can't draw cards".
	DrawLimitEveryone DrawLimitWhose = iota
	// DrawLimitOpponents is "Each opponent can't draw more than …"
	// (Narset, Parter of Veils; Leovold, Emissary of Trest).
	DrawLimitOpponents
	// DrawLimitYou is "You can't draw more than …" / "You can't draw
	// cards".
	DrawLimitYou
)

// DrawLimitStatic is one printed draw limit on a permanent. Catalog
// data, never stored.
type DrawLimitStatic struct {
	// Label is the clause as printed.
	Label string
	// Whose is which players it covers.
	Whose DrawLimitWhose
	// Max is how many cards a covered player may draw each turn: one
	// for "can't draw more than one card each turn", zero for "can't
	// draw cards".
	Max int
}

// CatalogDrawLimits returns the draw limits a permanent with the given
// catalog key has. carddef.go sets it from CardDef.DrawLimits; a
// game-package test may stub it.
var CatalogDrawLimits func(key string) []DrawLimitStatic

// ErrDrawRestricted is the manual draw's refusal when a draw limit
// forbids the player another draw this turn (CR 121.2b).
var ErrDrawRestricted = errors.New("game: that player can't draw another card this turn")

// drawLimitCovers reports whether a limit on `src` covers `player`.
func drawLimitCovers(whose DrawLimitWhose, src *Card, player uuid.UUID) bool {
	switch whose {
	case DrawLimitEveryone:
		return true
	case DrawLimitOpponents:
		return src.Controller != uuid.Nil && player != src.Controller
	case DrawLimitYou:
		return player == src.Controller
	}
	return false
}

// drawLimitForLocked is the lowest per-turn draw cap reaching `p`, and
// whether any does. Caller must hold g.mu. Reads only.
func (g *Game) drawLimitForLocked(p *Player) (limit int, limited bool) {
	if p == nil || CatalogDrawLimits == nil || g.Battlefield == nil {
		return 0, false
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		key := catalogAbilityKeyOf(src)
		if key == "" {
			continue
		}
		for _, s := range CatalogDrawLimits(key) {
			if !drawLimitCovers(s.Whose, src, p.ID) {
				continue
			}
			if !limited || s.Max < limit {
				limit, limited = s.Max, true
			}
		}
	}
	return limit, limited
}

// canDrawLocked is THE reader: may this player draw one more card right
// now? A player no limit reaches always may (an empty library is the
// draw's own business, CR 121.4). Caller must hold g.mu. Reads only.
func (g *Game) canDrawLocked(p *Player) bool {
	limit, limited := g.drawLimitForLocked(p)
	if !limited {
		return true
	}
	return len(g.DrawnThisTurn[p.ID]) < limit
}

// CanDrawForEffect reports whether `player` may draw another card this
// turn, for a card that offers a CHOICE to draw (CR 121.3: a player who
// can't draw can't choose to) and for the view. Caller must hold g.mu.
func (g *Game) CanDrawForEffect(player uuid.UUID) bool {
	return g.canDrawLocked(g.playerByIDLocked(player))
}

// DrawsLeftThisTurnForEffect is how many more cards `player` may draw
// this turn under the draw limits that reach them, and whether any
// does. Caller must hold g.mu.
func (g *Game) DrawsLeftThisTurnForEffect(player uuid.UUID) (left int, limited bool) {
	p := g.playerByIDLocked(player)
	limit, limited := g.drawLimitForLocked(p)
	if !limited {
		return 0, false
	}
	left = limit - len(g.DrawnThisTurn[p.ID])
	if left < 0 {
		left = 0
	}
	return left, true
}

// drawForbiddenLocked is the gate drawEventLocked asks before the CR 614
// window: is this a draw its player can't make? Caller must hold g.mu.
func (g *Game) drawForbiddenLocked(ev *ReplacementEvent) bool {
	if ev == nil || ev.Kind != RepEventDraw {
		return false
	}
	return !g.canDrawLocked(g.playerByIDLocked(ev.DrawPlayer))
}

// CanChooseToDrawForEffect reports whether an effect may offer
// `player` a CHOICE to draw n cards (#4016). CR 121.3: a player who
// can't draw can't choose to. CR 121.2b: under "can't draw more than
// one card each turn" a choice to draw more cards than the limit still
// allows can't be taken either. CR 121.3a: the same holds when another
// player makes the choice for the drawer. An empty library is no bar
// (CR 121.3), so only the draw limits are read. Caller must hold g.mu.
func (g *Game) CanChooseToDrawForEffect(player uuid.UUID, n int) bool {
	left, limited := g.DrawsLeftThisTurnForEffect(player)
	return !limited || n <= left
}

// DrawChoiceCeilingForEffect is the most cards a "draw up to n" choice
// may name for `player`: n, or the draws a limit still allows this turn
// when that is fewer (CR 121.2b, #4016). Caller must hold g.mu.
func (g *Game) DrawChoiceCeilingForEffect(player uuid.UUID, n int) int {
	if left, limited := g.DrawsLeftThisTurnForEffect(player); limited && left < n {
		return left
	}
	return n
}

// DrawLimitStatus is a seat's draw limit as its badge shows it (#4016):
// the cap, how many more cards it allows this turn, and the permanents
// that impose it, by name, in battlefield order.
type DrawLimitStatus struct {
	Max     int
	Left    int
	Sources []string
}

// DrawLimitStatusLocked is the draw limit reaching p, and whether any
// does. Sources are the permanents whose cap is the one that applies
// (the lowest), each named once. Caller must hold g.mu. Reads only.
func (g *Game) DrawLimitStatusLocked(p *Player) (DrawLimitStatus, bool) {
	limit, limited := g.drawLimitForLocked(p)
	if !limited {
		return DrawLimitStatus{}, false
	}
	st := DrawLimitStatus{Max: limit, Left: max(limit-len(g.DrawnThisTurn[p.ID]), 0)}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		key := catalogAbilityKeyOf(src)
		if key == "" {
			continue
		}
		for _, s := range CatalogDrawLimits(key) {
			if s.Max != limit || !drawLimitCovers(s.Whose, src, p.ID) {
				continue
			}
			if !slices.Contains(st.Sources, src.Name) {
				st.Sources = append(st.Sources, src.Name)
			}
		}
	}
	return st, true
}
