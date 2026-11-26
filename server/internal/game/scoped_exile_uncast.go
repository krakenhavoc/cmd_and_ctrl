package game

import "github.com/google/uuid"

// scoped_exile_uncast.go — "until end of turn, if a nontoken creature would
// enter and it wasn't cast, exile it instead" (Mistcaller, Core Set 2019,
// S82 older sets), the timed form of Containment Priest's static
// replacement (CR 614.1, 614.6): created by a resolving ability (CR 611.2)
// and held as a scoped record, so it outlives the ability's source and
// ends with its duration.
//
// "It was cast" is the move coming off the stack (OldZone stack); a creature
// put onto the battlefield from a hand, graveyard, library or exile, or
// blinked back, has some other OldZone and is exiled instead. A token is
// created from no zone and is excluded by name ("nontoken"). The mod kind
// is an on-disk identity: never renamed, never reused.

// ModExileUncastCreatures is "if a nontoken creature would enter and it
// wasn't cast, exile it instead" for as long as the record lasts. It reads
// nothing but the entering card. A replacement, read by the replacement
// gather rather than the layer pass.
const ModExileUncastCreatures ModKind = "exileUncastCreatures"

// ExileUncastCreaturesForEffect registers the replacement with duration `d`
// (Mistcaller: g.UntilEndOfTurnDuration()). The record's Controller is
// `controller`, the player who made it. Returns false, registering nothing,
// for a player who is not in the game.
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) ExileUncastCreaturesForEffect(sourceID, controller uuid.UUID, d Duration, label string) bool {
	if g.playerByIDLocked(controller) == nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, controller,
		[]Mod{{Kind: ModExileUncastCreatures}}, d, label)
}

// exileUncastCreaturesAppliesLocked is the kind's test: a nontoken creature
// card moving onto the battlefield from anywhere but the stack. Caller must
// hold g.mu.
func exileUncastCreaturesAppliesLocked(g *Game, ev *ReplacementEvent) bool {
	if ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || ev.OldZone == ZoneStack || ev.Canceled {
		return false
	}
	entering, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !entering.IsToken() && entering.IsCreature()
}
