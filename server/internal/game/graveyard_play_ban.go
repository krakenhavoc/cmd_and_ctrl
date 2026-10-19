package game

import "github.com/google/uuid"

// graveyard_play_ban.go — #1987, ADR 0066's amendment of 2026-10-16:
// "<player> can't play lands or cast spells from their graveyard this
// turn", the first sentence of Shaman's Trance ("Other players can't
// play lands or cast spells from their graveyards this turn").
//
// hand_play_ban.go's shape one zone over. "Play lands or cast spells"
// is both halves of CR 305.1 and CR 601, and each half already has its
// one gate, so the ban is one stored record read by both:
//
//   - CastGateLocked refuses a cast out of the player's OWN graveyard;
//   - LandPlayGateLocked refuses a land played out of it.
//
// "Their graveyards" is the banned player's own pile. A card in a
// graveyard is in its owner's graveyard (CR 404.1), so the pile is
// told by the card's owner and no zone scan is needed. A card the
// banned player may cast out of ANOTHER player's graveyard is not in
// "their graveyard" and is not touched, and nor is anything out of any
// other zone. Activated abilities of cards in the graveyard (unearth,
// embalm, scavenge) are not plays and stay open, which is what the
// printed text says.
//
// The record is a ModCantPlayFromOwnGraveyard ScopedEffect for the
// reason ModCantPlayLands is one: a binary from before this kind
// refuses a restore point naming it, where an unknown CastBanRule kind
// would be read as no ban at all.

// graveyardPlayBanReason is the clause a refused play reports, before
// the source's name.
const graveyardPlayBanReason = "You can't play lands or cast spells from your graveyard this turn"

// CantPlayFromOwnGraveyardThisTurnForEffect registers "<players> can't
// play lands or cast spells from their graveyards this turn": one record
// with one mod per player, swept at cleanup (CR 514.2). A player who is
// not seated is skipped; uuid.Nil names nobody, never everybody, because
// every printed form is "other players". `label` names the effect for
// the log and the refusal. Reports whether a record was written.
//
// Caller must hold g.mu (write).
func (g *Game) CantPlayFromOwnGraveyardThisTurnForEffect(sourceID uuid.UUID, players []uuid.UUID, label string) bool {
	var mods []Mod
	for _, id := range players {
		if id == uuid.Nil || g.playerByIDLocked(id) == nil {
			continue
		}
		mods = append(mods, Mod{Kind: ModCantPlayFromOwnGraveyard, Player: id})
	}
	if len(mods) == 0 {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		mods, g.UntilEndOfTurnDuration(), label)
}

// graveyardPermissionsReachAllLocked reports whether a live permission of
// `holder`'s makes their graveyard-casting permissions reach every
// graveyard: the last sentence of Shaman's Trance, "You may play lands
// and cast spells from other players' graveyards this turn as though
// those cards were in your graveyard" (#4177). The flag rides the
// Trance's own stored grants (CastPermission.GraveyardsAreYours), so the
// stored state of a resolved Trance gains a key and no record (snapshot
// fixture v9/shamans_trance.json stays a subset). Underworld Breach's
// escape, Yawgmoth's Will, Lurrus, Muldrotha and a flashback grant each
// speak of "your graveyard"; read by permissionReachesPileLocked, the
// one ownership rule they share.
//
// Caller must hold g.mu (read or write).
func (g *Game) graveyardPermissionsReachAllLocked(holder uuid.UUID) bool {
	p := g.playerByIDLocked(holder)
	if p == nil {
		return false
	}
	for i := range p.CastPermissions {
		perm := &p.CastPermissions[i]
		if perm.GraveyardsAreYours && g.CastPermissionActiveForEffect(perm, holder) {
			return true
		}
	}
	return false
}

// graveyardPlayBanLocked is THE reader: does a live record stop
// `player` playing `card` out of `zone` right now? Only a card in the
// player's own graveyard is ever refused. It returns the clause, with
// the source's name, and the source.
//
// The duration is asked here rather than left to the sweep, for the
// reason handPlayBanLocked gives.
//
// Caller must hold g.mu (read or write).
func (g *Game) graveyardPlayBanLocked(player uuid.UUID, card Card, zone ZoneKind) (reason string, source uuid.UUID, banned bool) {
	if zone != ZoneGraveyard || player == uuid.Nil || card.Owner != player {
		return "", uuid.Nil, false
	}
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind != ModCantPlayFromOwnGraveyard || m.Player != player {
				continue
			}
			if g.durationExpiredLocked(e.Duration, false) {
				continue
			}
			reason = graveyardPlayBanReason
			if e.SourceName != "" {
				reason += " — " + e.SourceName
			}
			return reason, e.Source.ID, true
		}
	}
	return "", uuid.Nil, false
}
