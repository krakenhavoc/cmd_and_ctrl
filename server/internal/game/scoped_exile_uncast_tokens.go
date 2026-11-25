package game

import "github.com/google/uuid"

// scoped_exile_uncast_tokens.go — "until end of turn, if a creature would
// enter and it wasn't cast, exile it instead" (Hallowed Moonlight, Magic
// Origins), Mistcaller's timed replacement without the "nontoken" (CR 614.1,
// 614.6). A creature token is created from no zone, was never cast, and is
// exiled instead of entering: it ceases to exist in exile (CR 111.7). The
// mod kind is an on-disk identity: never renamed, never reused.

// ModExileUncastCreaturesAndTokens is "if a creature would enter and it
// wasn't cast, exile it instead" for as long as the record lasts. It reads
// nothing but the entering object. A replacement, read by the replacement
// gather rather than the layer pass.
const ModExileUncastCreaturesAndTokens ModKind = "exileUncastCreaturesAndTokens"

// ExileUncastCreaturesAndTokensForEffect registers the replacement with
// duration `d` (Hallowed Moonlight: g.UntilEndOfTurnDuration()). The record's
// Controller is `controller`, the player who made it. Returns false,
// registering nothing, for a player who is not in the game.
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) ExileUncastCreaturesAndTokensForEffect(sourceID, controller uuid.UUID, d Duration, label string) bool {
	if g.playerByIDLocked(controller) == nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, controller,
		[]Mod{{Kind: ModExileUncastCreaturesAndTokens}}, d, label)
}

// exileUncastCreaturesAndTokensAppliesLocked is the kind's test: a creature
// card or creature token moving onto the battlefield from anywhere but the
// stack. Caller must hold g.mu.
func exileUncastCreaturesAndTokensAppliesLocked(g *Game, ev *ReplacementEvent) bool {
	if ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || ev.OldZone == ZoneStack || ev.Canceled {
		return false
	}
	entering, ok := g.LookupCardForEffect(ev.CardID)
	return ok && entering.IsCreature()
}
