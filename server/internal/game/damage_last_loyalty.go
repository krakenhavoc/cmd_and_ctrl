package game

import "github.com/google/uuid"

// damage_last_loyalty.go — "if damage dealt to a planeswalker you control
// … would result in all loyalty counters on it being removed, instead all
// but one of those counters are removed" (Deification, #5646).
//
// THE RULES.
//
//   - CR 120.3c: "Damage dealt to a planeswalker causes that many loyalty
//     counters to be removed from that planeswalker." CR 306.8 says the
//     same from the planeswalker's side.
//   - CR 614.1a: the sentence says "instead", so it is a replacement
//     effect. What it replaces is not the damage but the RESULT of the
//     damage. CR 120.4c: "damage that's been dealt is processed into its
//     results, as modified by replacement effects that interact with
//     those results (such as life loss or counters)."
//   - CR 611.3a: a static ability's effect "isn't 'locked in'", so the
//     condition and the planeswalker's type are read as the damage lands.
//
// So the damage is dealt in full. It is not prevented (CR 615 is never
// involved): lifelink gains the whole amount, "whenever … is dealt damage"
// triggers see the whole amount, a creature-planeswalker has all of it
// marked (CR 120.3e), and "damage can't be prevented" changes nothing.
// Only the number of loyalty counters that come off is capped, at one
// fewer than the planeswalker has. A planeswalker with exactly one
// loyalty counter keeps it: all but one of one counter is none.
//
// WHERE IT IS READ. In applyDamageToPermanentLocked, the one place damage
// becomes loyalty removal, after the CR 614 damage window has settled the
// amount (CR 120.4b before 120.4c). Read live off the battlefield, keyed
// by CatalogAbilityKey, and never stored (the DamageAsThoughStatic
// pattern): a source that left or lost its abilities caps nothing (CR
// 613.1f), and two of them say the same thing, so there is no order to
// choose (CR 616.1 has nothing to decide between identical results).
//
// Simultaneous combat damage from several sources lands one event at a
// time here. The first event that would empty the planeswalker leaves one
// counter and every later one finds one counter and removes none, which
// is the printed result for the whole instance.
//
// WHAT IT DOES NOT DO. It caps damage only. A loyalty cost, "remove N
// loyalty counters" and a counter-removing effect are not damage and take
// the last counter as they always did.

// DamageLastLoyaltyStatic is one printed "damage can't remove the last
// loyalty counter" static on a permanent. Catalog data, never stored: the
// gate reads it off the battlefield as damage lands on a planeswalker.
// It always covers planeswalkers its permanent's controller controls.
type DamageLastLoyaltyStatic struct {
	// Label is the clause as printed.
	Label string
	// OfChosenType narrows it to planeswalkers "of the chosen type": the
	// planeswalker type chosen as the static's permanent entered
	// (Card.ChosenOption, CR 614.12). Unanswered, it covers nothing.
	OfChosenType bool
	// WhileYouControlACreature is "as long as you control a creature",
	// about the static's controller, read as the damage lands.
	WhileYouControlACreature bool
}

// CatalogDamageLastLoyalty returns the statics a permanent with the given
// catalog key has. carddef.go sets it from CardDef.DamageLastLoyalty; a
// game-package test may stub it.
var CatalogDamageLastLoyalty func(key string) []DamageLastLoyaltyStatic

// loyaltyRemovedByDamageLocked is how many loyalty counters `amount`
// damage removes from `walker` (CR 120.3c): the amount, unless that would
// remove every counter it has and a DamageLastLoyaltyStatic covers it,
// when it is all but one of them. A planeswalker with no loyalty counters
// has none to keep, so the answer is the amount (which removes nothing).
//
// Caller must hold g.mu. Reads only.
func (g *Game) loyaltyRemovedByDamageLocked(walker *Card, amount int) int {
	have := walker.Counters[CounterLoyalty]
	if have <= 0 || amount < have || !g.damageKeepsLastLoyaltyLocked(walker) {
		return amount
	}
	return have - 1
}

// damageKeepsLastLoyaltyLocked reports whether a battlefield static stops
// damage removing `walker`'s last loyalty counter right now. A phased-out
// permanent is not in the battlefield slice (CR 702.26b).
//
// Caller must hold g.mu. Reads only.
func (g *Game) damageKeepsLastLoyaltyLocked(walker *Card) bool {
	if CatalogDamageLastLoyalty == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		if src.Controller != walker.Controller {
			continue
		}
		key := catalogAbilityKeyOf(src)
		if key == "" {
			continue
		}
		for _, s := range CatalogDamageLastLoyalty(key) {
			if s.OfChosenType && (src.ChosenOption == "" || !hasSubtype(walker, src.ChosenOption)) {
				continue
			}
			if s.WhileYouControlACreature && !g.controlsACreatureLocked(src.Controller) {
				continue
			}
			return true
		}
	}
	return false
}

// controlsACreatureLocked reports whether `player` controls a creature,
// by the layered view. Caller must hold g.mu. Reads only.
func (g *Game) controlsACreatureLocked(player uuid.UUID) bool {
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Controller == player && c.IsCreature() {
			return true
		}
	}
	return false
}

// DamageKeepsLastLoyaltyForEffect reports whether damage cannot remove
// `walker`'s last loyalty counter right now, for a caller outside the
// damage path (a test, a bot's "can this attack kill the walker" read).
// Caller must hold g.mu (read or write).
func (g *Game) DamageKeepsLastLoyaltyForEffect(walker Card) bool {
	return g.damageKeepsLastLoyaltyLocked(&walker)
}
