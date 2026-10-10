package game

import "github.com/google/uuid"

// hand_play_ban.go — #2559, ADR 0066's amendment of 2026-10-10: "they
// can't play cards from their hand" for a duration (Memory Vessel's
// "until your next turn").
//
// "Play" is both halves of CR 305.1 and CR 601: casting a spell and
// playing a land. Each half already has its one gate, so the ban is
// one stored record read by both, never a third gate:
//
//   - CastGateLocked refuses a cast whose source zone is the hand;
//   - LandPlayGateLocked refuses a land play out of the hand.
//
// Every caller of either gate — the cast path, a land play during a
// resolution, the bot enumerator and the view's `cant_cast` stamps —
// sees the ban for free. Nothing else out of a hand is a play: cycling
// and channel are activated abilities, foretell and ninjutsu are a
// special action and an ability, and "put a land card from your hand
// onto the battlefield" is not a land play (CR 305.4). None of them is
// gated, which is what the printed text says.
//
// The record is a ModCantPlayFromHand ScopedEffect, for the reason
// ModCantPlayLands is one: a binary from before this kind refuses a
// restore point naming it, where an unknown CastBanRule kind would be
// read as no ban at all.

// handPlayBanReason is the clause a refused play reports, before the
// source's name.
const handPlayBanReason = "You can't play cards from your hand"

// CantPlayFromHandForEffect registers "<player> can't play cards from
// their hand" until `d` runs out. `player` zero is every player, which
// is Memory Vessel's "players … they can't". `label` names the effect
// for the log and the record. Reports whether a record was written.
//
// Caller must hold g.mu (write).
func (g *Game) CantPlayFromHandForEffect(sourceID, player uuid.UUID, d Duration, label string) bool {
	if player != uuid.Nil && g.playerByIDLocked(player) == nil {
		return false
	}
	if d.IsZero() {
		d = g.UntilEndOfTurnDuration()
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		[]Mod{{Kind: ModCantPlayFromHand, Player: player}}, d, label)
}

// handPlayBanLocked is THE reader: does a live record stop `player`
// playing cards from their hand right now? It returns the clause, with
// the source's name, and the source.
//
// The duration is asked here rather than left to the sweep, for the
// reason castBanForbidsLocked gives: "until your next turn" ends as that
// turn begins, and the read has to be right between the sweeps.
//
// Caller must hold g.mu (read or write).
func (g *Game) handPlayBanLocked(player uuid.UUID) (reason string, source uuid.UUID, banned bool) {
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind != ModCantPlayFromHand {
				continue
			}
			if m.Player != uuid.Nil && m.Player != player {
				continue
			}
			if g.durationExpiredLocked(e.Duration, false) {
				continue
			}
			reason = handPlayBanReason
			if e.SourceName != "" {
				reason += " — " + e.SourceName
			}
			return reason, e.Source.ID, true
		}
	}
	return "", uuid.Nil, false
}

// HandPlayBanFor is the clause that stops `player` playing any card from
// their hand right now, or "" — the seat's banner
// (PlayerView.cant_play_from_hand).
//
// Caller must hold g.mu (read or write).
func (g *Game) HandPlayBanFor(player uuid.UUID) string {
	reason, _, _ := g.handPlayBanLocked(player)
	return reason
}

// The windows a permission's `exile_play.until` names (#2559): when the
// grant ENDS, so a tooltip can say "until Ana's next turn" instead of
// assuming "until end of turn", and a bystander can read another
// player's grant.
const (
	PermissionUntilEndOfTurn     = "end_of_turn"
	PermissionUntilEndOfNextTurn = "end_of_next_turn"
	PermissionUntilNextTurn      = "next_turn"
	PermissionUntilNextEndStep   = "next_end_step"
	PermissionWhileExiled        = "while_exiled"
	PermissionUntilAnother       = "until_another"
)

// PermissionWindowLocked names when `perm` ends and the player its
// window is counted against ("your next turn"), for the wire. "" for a
// kind no permission carries. "end_of_next_turn" is an UntilEndOfTurn
// stamped against a turn of its player that has not begun yet
// (UntilEndOfYourNextTurnDuration).
//
// Caller must hold g.mu (read or write).
func (g *Game) PermissionWindowLocked(perm *CastPermission) (until string, player uuid.UUID) {
	if perm == nil {
		return "", uuid.Nil
	}
	d := perm.Duration
	switch d.Kind {
	case UntilEndOfTurn:
		if g.turnsBegunForLocked(d.Player) < d.ExpiresAfterTurnsBegun {
			return PermissionUntilEndOfNextTurn, d.Player
		}
		return PermissionUntilEndOfTurn, uuid.Nil
	case UntilYourNextTurn:
		return PermissionUntilNextTurn, d.Player
	case UntilYourNextEndStep:
		return PermissionUntilNextEndStep, d.Player
	case WhileInZone:
		return PermissionWhileExiled, uuid.Nil
	case UntilSourceExilesAnother:
		return PermissionUntilAnother, uuid.Nil
	}
	return "", uuid.Nil
}
