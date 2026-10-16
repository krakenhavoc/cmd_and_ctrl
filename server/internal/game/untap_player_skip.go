package game

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// untap_player_skip.go — "<permanents> don't untap during <player>'s
// next untap step" (CR 502.3, #3995; ADR 0058's 2026-10-16 amendment).
//
// ADR 0058 Decision 2's marker sits on each permanent, so it freezes a
// SET: "those creatures don't untap" (Sleep, Frost Breath) names the
// creatures the spell tapped. The player-wide form names none: "Lands
// target player controls don't untap during their next untap step"
// (Mana Vapors), "Creatures don't untap during target player's next
// untap step" (Blinding Beam). It changes no characteristic, so CR
// 611.2c does not lock its set: it reaches a land played, or a creature
// whose control changed, after the spell resolved. So it is a RULE read
// at the step, not a marker written on each permanent.
//
// A skip is a ScopedEffect record (ScopeGame, IndefiniteDuration)
// carrying one ModSkipNextUntapStep mod: Player is whose untap step,
// Types the card types it covers (any of them), and AnyController
// whether it covers every permanent that would untap in that step
// (Blinding Beam's "Creatures") rather than only the ones Player
// controls ("creatures target player controls"). Records, not a field
// on the player, for the reason turn_skip.go gives: a binary that does
// not know the kind refuses the file instead of untapping what the
// card said would stay tapped.
//
// Read in untapStepSetLocked beside the per-permanent markers, and used
// up by Player's next untap step in performUntapStepLocked, at the same
// place and for the same reasons (ADR 0058 Decision 2): a skipped step
// (Stasis) never reaches it, so the skip waits for the next untap step
// that does happen (CR 614.10a), and an extra turn's untap step is that
// player's next one.

// ModSkipNextUntapStep is "<Types> [Player controls] don't untap during
// Player's next untap step" (#3995, CR 502.3). Scope ScopeGame,
// IndefiniteDuration; used up by Player's next untap step. Reads Player,
// Types (lowercase card types, any of them) and AnyController.
const ModSkipNextUntapStep ModKind = "skipNextUntapStep"

// untapSkipTypes is the card types a player-wide untap skip may name:
// the permanent types (CR 110.4).
var untapSkipTypes = []string{"artifact", "battle", "creature", "enchantment", "land", "planeswalker"}

// skipNextUntapStepModProblem is the registration and restore check for
// ModSkipNextUntapStep. "" when the mod is sound, or is another kind
// that does not carry AnyController.
func skipNextUntapStepModProblem(m Mod) string {
	if m.Kind != ModSkipNextUntapStep {
		if m.AnyController {
			return fmt.Sprintf("a %s mod carries anyController, which only skipNextUntapStep reads", m.Kind)
		}
		return ""
	}
	if m.Player == uuid.Nil {
		return "a skipNextUntapStep mod names no player"
	}
	if len(m.Types) == 0 {
		return "a skipNextUntapStep mod names no card type"
	}
	for _, t := range m.Types {
		if !slices.Contains(untapSkipTypes, t) {
			return fmt.Sprintf("a skipNextUntapStep mod names %q, which is not a permanent type", t)
		}
	}
	return ""
}

// SkipNextUntapStepForPlayerForEffect records that permanents of any of
// `types` don't untap during `player`'s next untap step (CR 502.3).
// With anyController false the skip covers only the permanents `player`
// controls as that step begins ("lands target player controls", "lands
// you control"); with it true, every permanent of those types that would
// untap in that step (Blinding Beam's "Creatures"), a Seedborn Muse's
// included. `types` are lowercase permanent types. `label` is the
// printed clause, for the record.
//
// Returns false, writing nothing, for a player who is not seated or has
// left the game (CR 800.4a), or for no types.
//
// Safe inside a resolution: it only writes a record.
//
// Caller must hold g.mu (write).
func (g *Game) SkipNextUntapStepForPlayerForEffect(player, source uuid.UUID, types []string, anyController bool, label string) bool {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated || len(types) == 0 {
		return false
	}
	m := Mod{Kind: ModSkipNextUntapStep, Player: player, Types: copyStrings(types), AnyController: anyController}
	if problem := skipNextUntapStepModProblem(m); problem != "" {
		panic(fmt.Sprintf("game: %s: %s", label, problem))
	}
	return g.RegisterScopedRuleEffectForEffect(source, ScopeGame, uuid.Nil, []Mod{m}, IndefiniteDuration(), label)
}

// playerUntapSkipCovers reports whether the mod m keeps `c` from
// untapping during `player`'s untap step.
func playerUntapSkipCovers(m Mod, c *Card, player uuid.UUID) bool {
	if m.Kind != ModSkipNextUntapStep || m.Player != player {
		return false
	}
	if !m.AnyController && c.Controller != player {
		return false
	}
	for _, t := range m.Types {
		if hasCardType(c, t) {
			return true
		}
	}
	return false
}

// playerUntapSkippedLocked reports whether a player-wide skip keeps `c`
// from untapping during `player`'s untap step. Read on fresh layers, so
// "creatures" means the permanents that are creatures as the step asks.
//
// Caller holds g.mu.
func (g *Game) playerUntapSkippedLocked(c *Card, player uuid.UUID) bool {
	for i := range g.ScopedEffects {
		for _, m := range g.ScopedEffects[i].Mods {
			if playerUntapSkipCovers(m, c, player) {
				return true
			}
		}
	}
	return false
}

// PlayerUntapSkipsCoveringLocked is the players whose next untap step a
// player-wide skip keeps `c` from untapping in, as the board stands: the
// view's "won't untap" badge (NoUntapView.Next) and the bots read it.
// Each player at most once, in record order.
//
// Only a player who controls `c` is named, AnyController or not: a
// permanent untaps in its own controller's untap step, and badging every
// creature at the table for Blinding Beam would tell its other players
// about a step their creatures sit out anyway. The step itself reads the
// whole rule (playerUntapSkippedLocked).
//
// Caller holds g.mu (read is enough).
func (g *Game) PlayerUntapSkipsCoveringLocked(c *Card) []uuid.UUID {
	if c == nil {
		return nil
	}
	var out []uuid.UUID
	for i := range g.ScopedEffects {
		for _, m := range g.ScopedEffects[i].Mods {
			if m.Kind != ModSkipNextUntapStep || m.Player != c.Controller || slices.Contains(out, m.Player) {
				continue
			}
			if playerUntapSkipCovers(m, c, m.Player) {
				out = append(out, m.Player)
			}
		}
	}
	return out
}

// consumePlayerUntapSkipsLocked drops every player-wide skip keyed to
// `activePlayer`'s untap step: the step that is happening is the "next"
// one they named. Builds a fresh slice, as every registry sweep does,
// so an undo snapshot taken before keeps them.
//
// Caller must hold g.mu (write).
func (g *Game) consumePlayerUntapSkipsLocked(activePlayer uuid.UUID) {
	spent := func(e ScopedEffect) bool {
		if len(e.Mods) == 0 {
			return false
		}
		for _, m := range e.Mods {
			if m.Kind != ModSkipNextUntapStep || m.Player != activePlayer {
				return false
			}
		}
		return true
	}
	if !slices.ContainsFunc(g.ScopedEffects, spent) {
		return
	}
	next := make([]ScopedEffect, 0, len(g.ScopedEffects))
	for _, e := range g.ScopedEffects {
		if !spent(e) {
			next = append(next, e)
		}
	}
	if len(next) == 0 {
		next = nil
	}
	g.ScopedEffects = next
}
