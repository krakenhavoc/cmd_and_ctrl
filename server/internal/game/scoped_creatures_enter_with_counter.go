package game

import "github.com/google/uuid"

// scoped_creatures_enter_with_counter.go — "this turn, each creature you
// control enters with an additional +1/+1 counter on it" (Combine
// Guildmage, Ravnica Allegiance; #5521).
//
// A CR 614.1c replacement effect ("[This permanent] enters with
// [counters]") a resolving ability creates (CR 611.2), held as a scoped
// record so it outlives the ability's source and ends at cleanup
// (CR 514.2); the sibling of ModPermanentsEnterTapped. The set is read
// live as each creature enters: a replacement effect is not a
// characteristic, so CR 611.2c does not lock it to the objects there when
// it began. Two resolved abilities are two records, so the counters stack.
// The kind is an on-disk identity: never renamed, never reused.

// ModCreaturesEnterWithPlusOneCounter is "each creature <Player> controls
// enters with Amount additional +1/+1 counters on it" for as long as the
// record lasts. Reads Player (the creature's controller as it enters; the
// record's controller is "you") and Amount (at least 1; a zero Amount reads
// as 1). A replacement, read by the replacement gather rather than the
// layer pass.
const ModCreaturesEnterWithPlusOneCounter ModKind = "creaturesEnterWithPlusOneCounter"

// creaturesEnterWithPlusOneCounterModProblem is the registration and restore
// check: "" when the mod is sound or is another kind.
func creaturesEnterWithPlusOneCounterModProblem(m Mod) string {
	if m.Kind != ModCreaturesEnterWithPlusOneCounter {
		return ""
	}
	if m.Player == uuid.Nil {
		return "a creaturesEnterWithPlusOneCounter replacement names no player"
	}
	return ""
}

// CreaturesEnterWithAnAdditionalPlusOneCounterThisTurnForEffect registers
// "this turn, each creature you control enters with an additional +1/+1
// counter on it" for `player` until cleanup. Reports whether a record was
// written (false for a player who is not in the game).
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) CreaturesEnterWithAnAdditionalPlusOneCounterThisTurnForEffect(sourceID, player uuid.UUID, label string) bool {
	if player == uuid.Nil || g.playerByIDLocked(player) == nil {
		return false
	}
	// A second resolution the same turn adds to the record already there:
	// the two replacements would otherwise open a CR 616.1 ordering prompt
	// over a choice that changes nothing, since both add counters.
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if len(e.Mods) == 1 && e.Mods[0].Kind == ModCreaturesEnterWithPlusOneCounter &&
			e.Mods[0].Player == player && e.Controller == player {
			e.Mods[0].Amount++
			return true
		}
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, player,
		[]Mod{{Kind: ModCreaturesEnterWithPlusOneCounter, Player: player, Amount: 1}},
		g.UntilEndOfTurnDuration(), label)
}

// creaturesEnterWithPlusOneCounterAppliesLocked is the kind's test: a
// creature moving onto the battlefield under the mod's player's control.
// Caller must hold g.mu.
func creaturesEnterWithPlusOneCounterAppliesLocked(g *Game, m Mod, ev *ReplacementEvent) bool {
	if ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || ev.Canceled {
		return false
	}
	entering, ok := g.LookupCardForEffect(ev.CardID)
	return ok && entering.IsCreature() && entering.Controller == m.Player
}
