package game

import "github.com/google/uuid"

// hexproof_bypass.go — two statics that reach INTO another object's
// protection without touching it (#1560, ADR 0038's amendment of
// 2026-09-27):
//
//   - "<these> can be the targets of spells and abilities [you
//     control] as though they didn't have hexproof" (CR 702.11) —
//     Nowhere to Run, Glaring Spotlight, Kaya, Bane of the Dead.
//   - "Ward abilities of those creatures don't trigger" (CR 702.21) —
//     Nowhere to Run.
//
// NEITHER IS A LAYER EFFECT. "As though it didn't have hexproof" is
// not "loses hexproof": the creature still has the ability (Glaring
// Spotlight's own ruling says so), a second waiver composes with the
// first, and one leaving the battlefield cannot revoke the other's.
// Removing "hexproof" from Characteristic.Abilities in layer 6 would
// get all three wrong, and would also strip it from the view the
// creature's own controller reads. Shadowspear's "lose hexproof" is
// the layer-6 shape and it stays one (RemoveKeywordsMod).
//
// So both are read LIVE off the battlefield on every query, the way
// block_rules.go reads a BlockRule and player_statics.go reads a
// PlayerKeywords token: nothing is written anywhere, so there is no
// state for undo, the snapshot or a restore point to carry. The
// source is keyed by CatalogAbilityKey, so a Nowhere to Run that has
// lost all its abilities (CR 613.1f) waives nothing.
//
// WHERE THEY ARE READ.
//
//   - The waiver is consulted by canBeTargetedBy (keywords.go) and by
//     canPlayerBeTargetedByLocked (player_statics.go), and ONLY once
//     hexproof is the thing that would refuse — so a board without
//     hexproof on it never walks for a waiver. Both halves of targeting
//     legality (the CR 601.2c announce gate and the CR 608.2b
//     resolution re-check) go through those two functions, so a spell
//     aimed at a hexproof creature while Nowhere to Run was out becomes
//     illegal if the enchantment leaves before it resolves — the card's
//     2024-09-20 ruling — with nothing in the resolution path touched.
//   - The ward half is consulted by effects.WardGranted's AppliesTo,
//     which is the one door every ward trigger in the catalog comes
//     through (printed Ward, granted ward, an emblem's, a face-down
//     permanent's). It is judged at the moment the ability would
//     trigger and at no other, so removing the enchantment later does
//     not make ward trigger retroactively (the second ruling), and a
//     ward trigger already on the stack still resolves.
//
// A WAIVER WITH A DURATION (#1651, ADR 0038's amendment of 2026-09-28,
// B1). Detection Tower's "{1}, {T}: Until end of turn, your opponents
// and creatures your opponents control with hexproof can be the targets
// of spells and abilities you control …" is the same waiver created by a
// resolving ability. It is an ADR 0041 data record (ModWaiveHexproof),
// read by the same two functions below, after the battlefield statics:
// the beneficiary is the record's Controller, and its set is the
// record's scope read live (a waiver changes no characteristic, so CR
// 611.2c does not lock it). Being data, it rides undo and the snapshot
// with the rest of the registry and ends at cleanup with it.
//
// WHAT IS NOT WAIVED. Only hexproof. Shroud (CR 702.18) and protection
// (CR 702.16b) are separate refusals and a hexproof waiver says
// nothing about them; canBeTargetedBy tests them before and after the
// waiver exactly as it did before it.

// HexproofBypass is one printed "as though it didn't have hexproof"
// static, as the engine reads it. Build it with the constructors in
// cards/effects/hexproof_bypass.go rather than by hand.
//
// `source` in both predicates is the permanent the static is printed
// on, so "your opponents" is `target.Controller != source.Controller`
// read live — a waiver on a permanent that changes control follows
// its new controller.
type HexproofBypass struct {
	// Permanent reports whether this static waives `target`'s
	// hexproof. `target` is a battlefield permanent that has hexproof.
	// Nil waives no permanent's.
	Permanent func(g *Game, target *Card, source *Card) bool

	// Player reports whether this static waives a PLAYER's hexproof
	// (CR 702.11d) — Kaya, Bane of the Dead's "your opponents … can be
	// the targets …". Nil waives no player's.
	Player func(g *Game, target *Player, source *Card) bool

	// YoursOnly is the printed "spells and abilities YOU control"
	// (Glaring Spotlight, Kaya): only a spell or ability the static's
	// controller controls may ignore the hexproof. False is Nowhere to
	// Run's unrestricted "spells and abilities" — any player's, the
	// creature's other opponents included.
	YoursOnly bool
}

// WardSuppression is one printed "ward abilities of <these> don't
// trigger" static (CR 702.21). Build it with effects.WardDoesNotTrigger.
type WardSuppression struct {
	// Warded reports whether the ward abilities of `warded`, a
	// battlefield permanent, don't trigger. `source` is the permanent
	// the static is printed on.
	Warded func(g *Game, warded *Card, source *Card) bool
}

// CatalogHexproofBypasses returns the hexproof waivers a battlefield
// permanent with the given catalog key imposes, or nil. carddef.go
// sets it from CardDef.HexproofBypasses. Keyed by CatalogAbilityKey:
// the waiver is a static ability of the permanent that prints it.
//
// Tests stub this directly, as keywords_test.go stubs
// CatalogPrintedKeywords.
var CatalogHexproofBypasses func(key string) []HexproofBypass

// CatalogWardSuppressions is CatalogHexproofBypasses's ward twin.
var CatalogWardSuppressions func(key string) []WardSuppression

// hexproofBypassedLocked reports whether some static on the
// battlefield lets a spell or ability `by` controls target the
// permanent `target` as though it didn't have hexproof.
//
// Nil-receiver safe: the exported CanBeTargetedBy has no game and asks
// through a nil one, which waives nothing.
//
// Caller must hold g.mu (read or write) with fresh layers. Reads only.
func (g *Game) hexproofBypassedLocked(target *Card, by uuid.UUID) bool {
	if g == nil || target == nil {
		return false
	}
	if g.anyHexproofBypassLocked(by, func(b HexproofBypass, source *Card) bool {
		return b.Permanent != nil && b.Permanent(g, target, source)
	}) {
		return true
	}
	return g.scopedHexproofWaiverLocked(by, target, nil)
}

// playerHexproofBypassedLocked is hexproofBypassedLocked for a
// player's hexproof (CR 702.11d). Caller must hold g.mu.
func (g *Game) playerHexproofBypassedLocked(target *Player, by uuid.UUID) bool {
	if g == nil || target == nil {
		return false
	}
	if g.anyHexproofBypassLocked(by, func(b HexproofBypass, source *Card) bool {
		return b.Player != nil && b.Player(g, target, source)
	}) {
		return true
	}
	return g.scopedHexproofWaiverLocked(by, nil, target)
}

// WaiveHexproofMod is "<affected> can be the targets of spells and
// abilities you control as though they didn't have hexproof" as an
// ADR 0041 data record (#1651, Detection Tower). Register it with
// RegisterScopedRuleEffectForEffect under ScopeOpponentsAndTheirCreatures;
// "you" is the record's Controller.
func WaiveHexproofMod() Mod { return Mod{Kind: ModWaiveHexproof} }

// scopedHexproofWaiverLocked reports whether a turn-scoped waiver
// record lets a spell or ability `by` controls target `card` (a
// battlefield permanent with hexproof) or `player` (a player with
// hexproof) as though it didn't have hexproof. Exactly one of the two
// is non-nil.
//
// Only a record whose Controller is `by` counts: every printed timed
// waiver says "spells and abilities YOU control" (A3's YoursOnly).
//
// Caller must hold g.mu. Reads only.
func (g *Game) scopedHexproofWaiverLocked(by uuid.UUID, card *Card, player *Player) bool {
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if e.Controller != by || !scopedEffectHasMod(e, ModWaiveHexproof) {
			continue
		}
		if player != nil {
			if scopeCoversPlayer(e.Scope, e.Controller, player.ID) {
				return true
			}
			continue
		}
		applies := affectedPredicate(e.Affected)
		if e.Scope != ScopeNone {
			applies = scopePredicate(e.Scope, e.Controller)
		}
		if applies(card, g, nil) {
			return true
		}
	}
	return false
}

// scopedEffectHasMod reports whether the record carries a mod of kind k.
func scopedEffectHasMod(e *ScopedEffect, k ModKind) bool {
	for _, m := range e.Mods {
		if m.Kind == k {
			return true
		}
	}
	return false
}

// scopeCoversPlayer reports whether a live scope names the PLAYER
// `player` (#1651). ScopeOpponentsAndTheirCreatures is the only scope
// that names players at all — "your opponents", the "you" being
// `controller` — so every other scope covers none, and no record
// written before it changes meaning.
func scopeCoversPlayer(scope AffectedScope, controller, player uuid.UUID) bool {
	return scope == ScopeOpponentsAndTheirCreatures && player != uuid.Nil && player != controller
}

// anyHexproofBypassLocked walks every waiver on the battlefield,
// skipping the "you control" ones whose controller is not `by`, and
// reports whether `waives` accepted any. Caller must hold g.mu.
func (g *Game) anyHexproofBypassLocked(by uuid.UUID, waives func(b HexproofBypass, source *Card) bool) bool {
	if CatalogHexproofBypasses == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		key := CatalogAbilityKey(*src)
		if key == "" {
			continue
		}
		for _, b := range CatalogHexproofBypasses(key) {
			if b.YoursOnly && by != src.Controller {
				continue
			}
			if waives(b, src) {
				return true
			}
		}
	}
	return false
}

// WardSuppressedForEffect reports whether a static on the battlefield
// says the ward abilities of `warded` don't trigger (CR 702.21, #1560).
// Read by effects.WardGranted's trigger condition, which every ward in
// the catalog is built on, at the moment the ability would trigger.
//
// Judged against the WARDED permanent, never the trigger's source: a
// granted ward's trigger belongs to the Equipment or emblem that
// grants it, but "ward abilities of those creatures" is about the
// creature that has it.
//
// Caller must hold g.mu (the trigger harvest does).
func (g *Game) WardSuppressedForEffect(warded *Card) bool {
	if g == nil || warded == nil || CatalogWardSuppressions == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		key := CatalogAbilityKey(*src)
		if key == "" {
			continue
		}
		for _, s := range CatalogWardSuppressions(key) {
			if s.Warded != nil && s.Warded(g, warded, src) {
				return true
			}
		}
	}
	return false
}
