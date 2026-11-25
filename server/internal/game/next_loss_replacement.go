package game

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// next_loss_replacement.go — "The next time you would lose the game this
// turn, instead draw seven cards and your life total becomes 1"
// (Stunning Reversal, #6244; CR 614.1a, 614.5, 614.6, 704.7, 119.5).
//
// The resolution-made sibling of The Golden Throne's static
// (game_action_replacements.go): a ScopedEffect record (ScopeGame,
// UntilEndOfTurn) with one ModNextLossInstead mod naming the player,
// read at the same door, replaceLossLocked, ahead of any permanent's
// replacement. The first loss it meets is replaced and the record is
// used up (CR 614.5: one opportunity), so a second loss that turn is not
// replaced; at the cleanup step it ends unused (UntilEndOfTurn). A
// concession is never replaced (CR 104.3a). With CR 704.7 one record
// replaces every loss a single state-based pass would deal the player.
// An older binary refuses a file carrying one (the unknown kind).

// ModNextLossInstead is "The next time <Player> would lose the game this
// turn, instead draw <Amount> cards and <Player>'s life total becomes
// <Power>" (#6244). Scope ScopeGame, UntilEndOfTurn. Reads Player,
// Amount (cards drawn, 0 or more) and Power (the life total set).
const ModNextLossInstead ModKind = "nextLossInstead"

// nextLossInsteadModProblem is the registration and restore check. ""
// when the mod is sound or is another kind.
func nextLossInsteadModProblem(m Mod) string {
	if m.Kind != ModNextLossInstead {
		return ""
	}
	if m.Player == uuid.Nil {
		return "a nextLossInstead mod names no player"
	}
	if m.Amount < 0 {
		return fmt.Sprintf("a nextLossInstead mod draws %d cards", m.Amount)
	}
	return ""
}

// ReplaceNextLossThisTurnForEffect records that the next time `player`
// would lose the game this turn, they instead draw `draw` cards and
// their life total becomes `life` (Stunning Reversal: 7 and 1). Returns
// false, writing nothing, for a player who is not seated or has left.
//
// Safe inside a resolution: it only writes a record.
//
// Caller must hold g.mu (write).
func (g *Game) ReplaceNextLossThisTurnForEffect(player, source uuid.UUID, draw, life int, label string) bool {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return false
	}
	m := Mod{Kind: ModNextLossInstead, Player: player, Amount: draw, Power: life}
	if problem := nextLossInsteadModProblem(m); problem != "" {
		panic(fmt.Sprintf("game: %s: %s", label, problem))
	}
	return g.RegisterScopedRuleEffectForEffect(source, ScopeGame, uuid.Nil, []Mod{m}, Duration{Kind: UntilEndOfTurn}, label)
}

func nextLossInsteadRecord(e ScopedEffect) (Mod, bool) {
	if len(e.Mods) != 1 || e.Mods[0].Kind != ModNextLossInstead {
		return Mod{}, false
	}
	return e.Mods[0], true
}

// replaceLossByRecordLocked applies p's oldest waiting record, if any,
// in place of a loss for `cause`, and reports whether it did. The record
// is removed before the draw, so a loss the draw itself would cause is
// not replaced by it.
//
// Caller must hold g.mu (write).
func (g *Game) replaceLossByRecordLocked(p *Player, cause LossCause) bool {
	idx := slices.IndexFunc(g.ScopedEffects, func(e ScopedEffect) bool {
		m, ok := nextLossInsteadRecord(e)
		return ok && m.Player == p.ID
	})
	if idx < 0 {
		return false
	}
	rec := g.ScopedEffects[idx]
	m := rec.Mods[0]
	source := rec.Source.ID
	g.dropScopedEffectsLocked(func(i int, _ ScopedEffect) bool { return i == idx })
	g.EmitEvent(Event{
		Kind:   EventGameLossReplaced,
		Actor:  p.ID,
		Source: source,
		CardID: source,
		Label:  string(cause),
	})
	if m.Amount > 0 {
		_ = g.DrawNForEffect(p.ID, m.Amount)
	}
	// "your life total becomes N" (CR 119.5).
	if delta := m.Power - p.Life; delta != 0 {
		_ = g.ChangePlayerLifeForEffect(source, p.ID, delta)
	}
	return true
}

// NextLossReplacementsForLocked is how many waiting records will replace
// `player`'s next loss this turn, for the view and the bots.
//
// Caller must hold g.mu (read or write).
func (g *Game) NextLossReplacementsForLocked(player uuid.UUID) int {
	n := 0
	for _, e := range g.ScopedEffects {
		if m, ok := nextLossInsteadRecord(e); ok && m.Player == player {
			n++
		}
	}
	return n
}
