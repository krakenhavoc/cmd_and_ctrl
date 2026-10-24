package game

import "github.com/google/uuid"

// last_turn_actions.go — who PUT a permanent onto the battlefield, and
// whether a player cast a spell or put a nontoken permanent onto the
// battlefield during their last turn (#1997, ADR 0109 amendment
// 2026-10-23, Arboria).
//
// Arboria: "Creatures can't attack a player unless that player cast a
// spell or put a nontoken permanent onto the battlefield during their
// last turn." Its 2022 ruling: "It's unusual for an ability to care who
// put a permanent onto the battlefield, as opposed to under whose
// control it entered." So the record is of the PUTTER, which is not
// always the player the permanent enters under:
//
//   - a land play is its player's (CR 305.1: "the player simply puts the
//     land onto the battlefield");
//   - a resolving permanent spell is its controller's (CR 608.3a, CR
//     608.3e: "its controller can't put it onto the battlefield");
//   - an effect's put is made by the controller of the resolving spell or
//     ability, who follows its instructions (CR 608.2c). That is who puts
//     a flickered creature back "under its owner's control", or an
//     opponent's card with Open the Vaults;
//   - except where the instruction names the player who puts: "each
//     player puts" (Exhume, Living Death), "its controller may search
//     their library … put it onto the battlefield" (Path to Exile), "that
//     player manifests" (Reality Shift). The player named puts it, and it
//     enters under them (CR 110.2a). A search and a manifest say so
//     themselves; a card whose own words name the putter runs its put
//     inside PutByEnteringPlayerForEffect.
//
// The record is a per-turn count (PlayerTurnTally.NontokenPermanentsPut)
// taken as each entry is announced, so a permanent that leaves before
// the turn ends still counts ("Arboria's effect cares about the actions
// taken by players, not their results"). As a turn ends its active
// player's spells cast and permanents put are folded into
// Player.LastTurnCastOrPut, beside ADR 0108 §6's Player.LastTurnAttacks
// and with the same meaning of "your last turn": an extra turn replaces
// it, a turn not taken writes nothing, and a player who has taken no
// turn yet has not done either.

// PutByEnteringPlayerForEffect runs fn with every permanent it puts onto
// the battlefield recorded as put there by the player it enters under,
// rather than by the controller of the resolving spell or ability: the
// put of an instruction that names the player who puts ("each player
// puts a creature card from their graveyard onto the battlefield", "they
// put it onto the battlefield"; CR 110.2a). A put that pauses for a
// prompt keeps the answer: the entry carries it.
//
// Caller must hold g.mu.
func (g *Game) PutByEnteringPlayerForEffect(fn func() error) error {
	g.putByEntering++
	defer func() { g.putByEntering-- }()
	return fn()
}

// stampPutByEnteringLocked marks an entry made inside
// PutByEnteringPlayerForEffect, as it begins, so a resume after a pause
// still knows. Caller must hold g.mu.
func (g *Game) stampPutByEnteringLocked(ev *ReplacementEvent) {
	if ev != nil && g.putByEntering > 0 {
		ev.putByEntering = true
	}
}

// entryPutterLocked is the player who put the landed permanent onto the
// battlefield (see the file comment), or uuid.Nil when nobody did.
//
// Caller must hold g.mu.
func (g *Game) entryPutterLocked(l entryLanding) uuid.UUID {
	ev := l.ev
	c := findBattlefieldCard(g, l.entered)
	if c == nil || ev == nil {
		return uuid.Nil
	}
	switch {
	case ev.landPlay:
		if ev.landPlayer != uuid.Nil {
			return ev.landPlayer
		}
		if ev.Actor != uuid.Nil {
			return ev.Actor
		}
		return c.Controller
	case ev.putByEntering || g.putByEntering > 0:
		return c.Controller
	case ev.stackItem != nil:
		return ev.stackItem.Controller
	case g.resolutionOpen && g.resolving != nil && g.resolving.item != nil:
		return g.resolving.item.Controller
	}
	// No spell or ability is resolving: a sandbox move or a replacement
	// with no instruction behind it. The player it enters under (CR
	// 110.2a's default).
	return c.Controller
}

// notePermanentPutLocked counts a landed NONTOKEN permanent for the
// player who put it there. Run as each entry is announced. Caller must
// hold g.mu.
func (g *Game) notePermanentPutLocked(l entryLanding) {
	c := findBattlefieldCard(g, l.entered)
	if c == nil || c.IsToken() {
		return
	}
	if who := g.entryPutterLocked(l); who != uuid.Nil {
		g.bumpPlayerTally(who, func(p *PlayerTurnTally) { p.NontokenPermanentsPut++ })
	}
}

// recordLastTurnCastOrPutLocked folds the ending turn's spells cast and
// nontoken permanents put by its active player into that player's
// LastTurnCastOrPut. It runs as the turn ends, beside
// recordLastTurnAttacksLocked, before the tally is replaced. Copies are
// not cast (CR 707.10), so CastTally.Total is the number the words
// count. Caller must hold g.mu.
func (g *Game) recordLastTurnCastOrPutLocked() {
	seat := g.Turn.ActiveSeat
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil {
		return
	}
	p := g.Seats[seat]
	p.LastTurnCastOrPut = g.CastTallyFor(p.ID).Total > 0 || g.TurnTallyFor(p.ID).NontokenPermanentsPut > 0
}

// CastOrPutDuringLastTurnForEffect reports whether `player` cast a spell
// or put a nontoken permanent onto the battlefield during their last
// turn: Arboria's "unless that player …". False for a player who has
// taken no turn yet, and for one who is not at the table.
//
// Caller must hold g.mu.
func (g *Game) CastOrPutDuringLastTurnForEffect(player uuid.UUID) bool {
	p := g.playerByIDLocked(player)
	return p != nil && p.LastTurnCastOrPut
}
