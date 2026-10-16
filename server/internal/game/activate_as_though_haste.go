package game

import "github.com/google/uuid"

// activate_as_though_haste.go — "You may activate abilities of creatures
// you control as though those creatures had haste" (Thousand-Year Elixir,
// Tyvar, Jubilant Brawler; #3928).
//
// THE RULES.
//
//   - CR 302.6: a creature's activated ability with {T} or {Q} in its cost
//     can't be activated unless the creature has been under its
//     controller's control continuously since their most recent turn
//     began. The same rule stops it attacking.
//   - CR 702.10b and 702.10c: haste lifts both halves, the attack and
//     the activation. CR 602.5a restates the activation half.
//   - CR 609.4: an "as though" effect applies only to what it states. So
//     the static lifts the activation half for its controller and nothing
//     else: the creature gains no haste, still can't attack, and a
//     "creatures with haste" reader does not count it.
//
// WHERE IT IS READ. Every reader of the activation half asks
// ActivationSummoningSickLocked: ActivateCatalogAbility and
// ActivateManaAbility (the engine's refusals), the auto-tapper's planner
// and executor, internal/legal's enumerator (the bot's moves and the
// client's legal-action digest), and the view's
// `activates_as_though_haste`, which the client reads before it greys a
// {T} row for sickness. The attack half keeps reading HasSummoningSickness
// directly, so it is untouched.
//
// WHOSE PERMISSION. "You may activate abilities of creatures you
// control": the static's controller must be the activator, and the
// creature must be theirs too. An opponent activating one of those
// creatures' "any player may activate" abilities is not "you", so a row
// across the table stays sick.

// CatalogActivateAsThoughHaste reports whether a battlefield permanent
// with this catalog key has the printed static. Populated from
// effects.Spec.ActivateAsThoughHaste. A nil hook grants nothing.
var CatalogActivateAsThoughHaste func(key string) bool

// activatesAsThoughHaste answers the static for one permanent through
// CatalogAbilityKey: a permanent that has lost its abilities grants
// nothing (CR 613.1f).
func (c *Card) activatesAsThoughHaste() bool {
	if CatalogActivateAsThoughHaste == nil {
		return false
	}
	key := catalogAbilityKeyOf(c)
	return key != "" && CatalogActivateAsThoughHaste(key)
}

// activateAsThoughHasteLocked reports whether a permanent on the
// battlefield lets `activator` activate the abilities of creature `c` as
// though it had haste. A phased-out permanent is not in the battlefield
// slice (CR 702.26b).
//
// Caller must hold g.mu. Reads only.
func (g *Game) activateAsThoughHasteLocked(c *Card, activator uuid.UUID) bool {
	if g == nil || g.Battlefield == nil || c == nil || activator == uuid.Nil ||
		c.Controller != activator || CatalogActivateAsThoughHaste == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		if src.Controller == activator && src.activatesAsThoughHaste() {
			return true
		}
	}
	return false
}

// ActivationSummoningSickLocked reports whether CR 302.6 stops
// `activator` paying a {T} or {Q} cost of permanent `c`: it is a
// summoning-sick creature (HasSummoningSickness, haste included) and no
// "as though those creatures had haste" static of the activator's reaches
// it. The question every activation reader asks; the attack reader asks
// HasSummoningSickness instead.
//
// Layers must be fresh, as for HasSummoningSickness. Caller must hold
// g.mu (read or write).
func (g *Game) ActivationSummoningSickLocked(c *Card, activator uuid.UUID) bool {
	if c == nil || !c.IsCreature() || !HasSummoningSickness(c) {
		return false
	}
	return !g.activateAsThoughHasteLocked(c, activator)
}

// ActivatesAsThoughHasteLocked reports whether creature `c`'s controller
// may activate its abilities as though it had haste right now. The view
// stamps it beside summoning_sick so the client's {T} rows read the same
// permission the engine does. Caller must hold g.mu.
func (g *Game) ActivatesAsThoughHasteLocked(c *Card) bool {
	return c != nil && g.activateAsThoughHasteLocked(c, c.Controller)
}
