package game

// cleanup_damage.go — "damage isn't removed from this creature during
// cleanup steps" (Ancient Adamantoise, #2058).
//
// CR 514.2 removes all damage marked on permanents in the cleanup step,
// "including phased-out permanents". A printed static can exempt its own
// permanent from that sweep; nothing else about the damage changes. It is
// still marked, still counts toward lethal damage (CR 704.5g) on a later
// turn, and any other effect that removes damage (regeneration, CR 701.19)
// or any battlefield exit (clearBattlefieldDamage) still clears it.

// CatalogDamageStaysThroughCleanup reports whether a battlefield
// permanent with this catalog key keeps its marked damage through the
// CR 514.2 sweep. Populated by the cards/effects package from
// effects.Spec.DamageStaysThroughCleanup. A nil hook exempts nothing.
var CatalogDamageStaysThroughCleanup func(oracleID string) bool

// keepsDamageThroughCleanup answers the exemption for one permanent.
//
// It reads CatalogAbilityKey, so a permanent that has lost its
// abilities (CR 613.1f), or lost this one (#1859), is cleaned as usual.
// A phased-out permanent is treated as though it does not exist
// (CR 702.26b), so its static does nothing and it is always cleaned;
// callers sweeping Game.PhasedOut simply do not ask.
func (c *Card) keepsDamageThroughCleanup() bool {
	if CatalogDamageStaysThroughCleanup == nil {
		return false
	}
	key := catalogAbilityKeyOf(c)
	return key != "" && CatalogDamageStaysThroughCleanup(key)
}
