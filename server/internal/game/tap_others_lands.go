package game

import "github.com/google/uuid"

// tap_others_lands.go — Piracy (#6204): "Until end of turn, you may tap
// lands you don't control for mana. Spend this mana only to cast
// spells."
//
//	602.2. … Only an object's controller (or its owner, if it doesn't
//	       have a controller) can activate its activated ability unless
//	       the object specifically says otherwise.
//	605.3a A player may activate an activated mana ability whenever they
//	       have priority, whenever they are casting a spell or activating
//	       an ability that requires a mana payment, or whenever a rule or
//	       effect asks for a mana payment …
//
// A ModTapLandsYouDontControl record (ScopeGame, Player the one player,
// until end of turn) is the effect that "specifically says otherwise":
// while it lasts, that player may activate the mana abilities of lands
// other players control. One predicate answers "may this player activate
// this mana row" (MayActivateManaForEffect) for the activation path, the
// legal-move enumerator and the view, so the bot, the client and the
// engine agree. The mana goes to the activator's pool (CR 106.4), and
// every mana such an activation makes carries ManaRestrictCast ("Spend
// this mana only to cast spells"), stamped where every restriction is
// read (restrictionsFor), so the click path, a pending colour pick and
// the auto-tapper's planned output are restricted alike. The auto-tapper
// gathers those lands as sources for the player, as restricted output it
// plans only for a cast (autoTapAdmitsRestricted).
//
// A row the land prints for any player (ADR 0106) is that row's own
// permission and is not restricted by this; a land's non-mana activated
// abilities are not mana abilities and stay its controller's.

// ModTapLandsYouDontControl is Piracy's permission. Reads Player. Scope
// ScopeGame. An on-disk identity: never renamed, never reused.
const ModTapLandsYouDontControl ModKind = "tapLandsYouDontControl"

// TapLandsYouDontControlThisTurnForEffect registers Piracy's permission
// for `player` until end of turn. Reports whether a record was written.
//
// Caller must hold g.mu (write).
func (g *Game) TapLandsYouDontControlThisTurnForEffect(sourceID, player uuid.UUID, label string) bool {
	if player == uuid.Nil || g.playerByIDLocked(player) == nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, player,
		[]Mod{{Kind: ModTapLandsYouDontControl, Player: player}}, g.UntilEndOfTurnDuration(), label)
}

// mayTapLandsYouDontControlLocked reports whether a live record lets
// `player` tap lands they don't control for mana.
//
// Caller must hold g.mu (read or write).
func (g *Game) mayTapLandsYouDontControlLocked(player uuid.UUID) bool {
	if player == uuid.Nil {
		return false
	}
	for i := range g.ScopedEffects {
		for _, m := range g.ScopedEffects[i].Mods {
			if m.Kind == ModTapLandsYouDontControl && m.Player == player {
				return true
			}
		}
	}
	return false
}

// MayTapLandsYouDontControlForEffect is the exported reader, for the
// view, the enumerator and card files. Caller must hold g.mu.
func (g *Game) MayTapLandsYouDontControlForEffect(player uuid.UUID) bool {
	return g.mayTapLandsYouDontControlLocked(player)
}

// tapsLandAcrossLocked reports whether `player` activating `ab` of
// `source` in `zone` is a Piracy activation: a land on the battlefield
// another player controls, a row not printed for any player, and a live
// permission.
//
// Caller must hold g.mu (read or write).
func (g *Game) tapsLandAcrossLocked(player uuid.UUID, source Card, zone ZoneKind, ab ManaAbilityShape) bool {
	return zone == ZoneBattlefield && source.IsLand() && source.Controller != player && !ab.AnyPlayer &&
		g.mayTapLandsYouDontControlLocked(player)
}

// MayActivateManaForEffect is MayActivateMana with Piracy's permission:
// the one answer to "may `player` activate this mana row" the activation
// path, the enumerator and the view ask. Caller must hold g.mu.
func (g *Game) MayActivateManaForEffect(player uuid.UUID, source Card, zone ZoneKind, ab ManaAbilityShape) bool {
	return MayActivateMana(player, source, zone, ab) || g.tapsLandAcrossLocked(player, source, zone, ab)
}

// MayVisitManaSourceForEffect is the cheap negative for a battlefield
// permanent `player` does not control: false when none of its mana rows
// can be theirs to activate. Caller must hold g.mu.
func (g *Game) MayVisitManaSourceForEffect(player uuid.UUID, source Card) bool {
	return source.Controller == player || HasAnyPlayerManaAbility(source) ||
		(source.IsLand() && g.mayTapLandsYouDontControlLocked(player))
}

// manaOutputRestrictedLocked is manaAbilityRestricted for a particular
// activator: the row's own restrictions, or Piracy's.
//
// Caller must hold g.mu (read or write).
func (g *Game) manaOutputRestrictedLocked(asker uuid.UUID, source Card, a ManaAbilityShape) bool {
	return manaAbilityRestricted(a) || g.tapsLandAcrossLocked(asker, source, ZoneBattlefield, a)
}

// withAcrossRestrictionLocked appends ManaRestrictCast to `rs` when
// `activator` making mana with `ab` of the permanent `source` is a
// Piracy activation.
//
// Caller must hold g.mu (read or write).
func (g *Game) withAcrossRestrictionLocked(rs []string, ab *ManaAbilityShape, activator, source uuid.UUID) []string {
	if ab == nil {
		return rs
	}
	c := findBattlefieldCard(g, source)
	if c == nil || !g.tapsLandAcrossLocked(activator, *c, ZoneBattlefield, *ab) {
		return rs
	}
	for _, r := range rs {
		if r == ManaRestrictCast {
			return rs
		}
	}
	return append(rs, ManaRestrictCast)
}
