package game

import "github.com/google/uuid"

// free_cast_command.go — #6242: "you may cast your commander from the
// command zone without paying its mana cost. (You still pay any
// additional costs.)" (Geode Golem).
//
// CR 903.8: a player may cast a commander they own from the command
// zone, and each previous cast from there that game adds {2} to the
// cost. CR 118.9: "without paying its mana cost" is an alternative cost,
// and the commander tax is an additional cost applied to it
// (printedCostLocked layers the tax over whichever cost was chosen), so
// a second free cast costs {2}. The cast is a cast from the command zone
// like any other: it bumps Player.CommanderCasts.
//
// The grant is a stored ScopeCards permission over the commander in the
// command zone, priced like a hand grant (AltCostKeyGrantedCost) and
// pass-closed: offered through the resolution_cast prompt
// (OfferResolutionCastForEffect) during the resolution of the effect,
// with flash timing (CR 608.2g), and gone if it is not taken. The
// command zone is otherwise closed to permissions (castPermissionLocked),
// so nothing else changes for an ordinary commander cast.

// LapseStaysInCommandZone is the pass-closed lapse of a command-zone
// grant: nothing moves; the grant is simply gone.
const LapseStaysInCommandZone PermissionLapse = "stays_in_command"

// FreeCastFromCommandPermission is the grant "you may cast your
// commander from the command zone without paying its mana cost" stamps
// over `player`'s commanders: free, any timing during the resolution,
// cast-only and pass-closed. Caller must hold g.mu.
func (g *Game) FreeCastFromCommandPermission(player, source uuid.UUID, sourceName string) CastPermission {
	return CastPermission{
		Player:      player,
		Zone:        ZoneCommand,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		AltCostKey:  AltCostKeyGrantedCost,
		Timing:      TimingFlash,
		CastOnly:    true,
		LapseOnPass: LapseStaysInCommandZone,
		Source:      source,
		SourceName:  sourceName,
		Label:       sourceName + " — cast your commander without paying its mana cost",
	}
}

// OwnCommandersInCommandZoneForEffect is the commanders `player` owns
// that are in their command zone now: the cards CR 903.8 lets them cast
// from there. Caller must hold g.mu.
func (g *Game) OwnCommandersInCommandZoneForEffect(player uuid.UUID) []uuid.UUID {
	p := g.playerByIDLocked(player)
	if p == nil || p.Command == nil {
		return nil
	}
	var out []uuid.UUID
	for i := range p.Command.Cards {
		c := &p.Command.Cards[i]
		if c.IsCommander && c.Owner == player {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// storedCommandPermissionLocked is the first live stored permission of
// playerID's that names `card` in the command zone and that `accept`
// takes, or nil. Only the card's owner's commander is ever opened (CR
// 903.8: "a commander they own"). Caller must hold g.mu.
func (g *Game) storedCommandPermissionLocked(playerID uuid.UUID, card Card, accept func(*CastPermission) bool) *CastPermission {
	p := g.playerByIDLocked(playerID)
	if p == nil || !card.IsCommander || card.Owner != playerID {
		return nil
	}
	for i := range p.CastPermissions {
		perm := &p.CastPermissions[i]
		if perm.Scope != ScopeCards || !perm.CoversCard(card, ZoneCommand) || !g.CastPermissionActiveForEffect(perm, playerID) {
			continue
		}
		out := *perm
		if acceptsPermission(accept, &out) {
			return &out
		}
	}
	return nil
}
