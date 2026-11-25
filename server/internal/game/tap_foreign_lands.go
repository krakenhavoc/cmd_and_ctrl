package game

import "github.com/google/uuid"

// tap_foreign_lands.go is "Until end of turn, you may tap lands you don't
// control for mana. Spend this mana only to cast spells." (Piracy, #6204,
// ADR 0106 §1 amendment 2026-11-25).
//
// CR 602.2: only an object's controller can activate its activated
// ability "unless the object specifically says otherwise". Piracy is the
// other case: an effect, not the land, says otherwise. The permission is
// the PLAYER's (TapForeignLandsThisTurn), read live, so a land that
// changes hands later this turn is covered and one that comes under the
// player's control is simply theirs.
//
// "Tap … for mana" is activating a mana ability with {T} in its cost (CR
// 106.12), so only those rows of a land are reachable: a land's mana
// ability without {T} stays its controller's. Everything else about the
// activation reads the activator, as an any-player row does (ADR 0106
// §1): the costs are the activator's, the mana goes to the activator's
// pool, and a rider's "you" is the activator, who controls the ability
// (CR 113.8).
//
// "Spend this mana only to cast spells" is ManaRestrictCast added to the
// mana's tags by restrictionsFor, the one place every mint (the click,
// the colour pick, the auto-tapper's executor) reads them from.

// AllowTapForeignLandsThisTurnForEffect is the resolution of "Until end of
// turn, you may tap lands you don't control for mana. Spend this mana only
// to cast spells." for `player`. A second resolution changes nothing.
//
// Caller must hold g.mu (write).
func (g *Game) AllowTapForeignLandsThisTurnForEffect(player uuid.UUID) {
	if player == uuid.Nil {
		return
	}
	if g.TapForeignLandsThisTurn == nil {
		g.TapForeignLandsThisTurn = make(map[uuid.UUID]bool)
	}
	g.TapForeignLandsThisTurn[player] = true
}

// MayTapForeignLandsLocked reports whether `player` may tap lands they
// don't control for mana this turn. Caller must hold g.mu.
func (g *Game) MayTapForeignLandsLocked(player uuid.UUID) bool {
	return g.TapForeignLandsThisTurn[player]
}

// tapsForeignLand reports whether `player` activating row `ab` of
// `source` is the Piracy case: a {T} mana ability of a land on the
// battlefield that someone else controls, with the permission held, and
// not a row any player may activate anyway (that one's mana is
// unrestricted, as its card prints). Caller must hold g.mu.
func (g *Game) tapsForeignLand(player uuid.UUID, source Card, zone ZoneKind, ab ManaAbilityShape) bool {
	return zone == ZoneBattlefield &&
		source.Controller != player &&
		!ab.AnyPlayer &&
		ab.TapCost &&
		// CR 701.21a: a player can sacrifice only a permanent they
		// control, so a row that sacrifices its land is not theirs.
		!ab.SacrificeCost &&
		source.IsLand() &&
		g.MayTapForeignLandsLocked(player)
}

// MayActivateManaLocked is MayActivateMana with the game in hand: the
// printed answer, widened by a Piracy permission for a land's {T} rows.
// The one gate ActivateManaAbility, the enumerator and the view share.
// Caller must hold g.mu.
func (g *Game) MayActivateManaLocked(player uuid.UUID, source Card, zone ZoneKind, ab ManaAbilityShape) bool {
	return MayActivateMana(player, source, zone, ab) || g.tapsForeignLand(player, source, zone, ab)
}

// MayReachManaAcrossLocked is the fast negative for a battlefield
// permanent `player` does not control: whether any of its mana rows could
// be theirs to activate (an any-player row, or a land under a Piracy
// permission). The row actually named is still held to
// MayActivateManaLocked. Caller must hold g.mu.
func (g *Game) MayReachManaAcrossLocked(player uuid.UUID, source Card) bool {
	return HasAnyPlayerManaAbility(source) || (source.IsLand() && g.MayTapForeignLandsLocked(player))
}

// foreignLandRestrictionLocked is the "Spend this mana only to cast
// spells" tag for mana `activator` makes with a land they don't control,
// or nil. Read by restrictionsFor. Caller must hold g.mu.
func (g *Game) foreignLandRestrictionLocked(ab *ManaAbilityShape, activator, source uuid.UUID) []string {
	if g == nil || g.TapForeignLandsThisTurn == nil || g.Battlefield == nil {
		return nil
	}
	c, ok := g.battlefieldCardLocked(source)
	if !ok || !g.tapsForeignLand(activator, *c, ZoneBattlefield, *ab) {
		return nil
	}
	return []string{ManaRestrictCast}
}
