package game

// lose_abilities_except_mana.go — #6246: "All lands lose all abilities
// except mana abilities." (Blood Sun). CR 613.1f: an ability-removing
// effect, applied in layer 6; CR 605.1a says which abilities are mana
// abilities, and the catalog declares them as mana rows.
//
// It is a removal like any other (StaticAbility.RemovesAbilities, so
// AbilitiesRemoved is set and every reader of an own triggered,
// activated, static or replacement ability, and every keyword, sees it
// gone), narrowed by StaticAbility.KeepsManaAbilities:
//
//   - Characteristic.ManaAbilitiesKept lets the object's own declared
//     mana rows answer (untaxedManaAbilityRows). The intrinsic mana
//     ability of a basic land type (CR 305.6) is kept as it is under any
//     removal.
//   - a layer-6 grant whose bundle adds only mana abilities (Chromatic
//     Lantern's "{T}: Add one mana of any color") survives the removal
//     whatever its timestamp; any other grant is removed as usual, and a
//     grant sorted after the removal survives as usual (CR 613.6).
//   - two removals in one pass compose: one that keeps nothing, earlier
//     or later, leaves nothing; one that keeps mana never gives back what
//     an earlier full removal took.

// effectKeepsManaAbilities reports whether a removing effect is the
// "except mana abilities" kind.
func effectKeepsManaAbilities(eff ContinuousEffect) bool {
	s, ok := eff.(staticContinuousEffect)
	return ok && s.ability.KeepsManaAbilities
}

// keptGrantsUnderRemoval is what an ability removal leaves of the grants
// already applied in its bucket: none, or under a removal that keeps
// mana abilities, the grants whose bundle adds mana abilities and
// nothing else.
func keptGrantsUnderRemoval(grants []GrantedAbility, keepsMana bool) []GrantedAbility {
	if !keepsMana || len(grants) == 0 {
		return nil
	}
	var out []GrantedAbility
	for _, gr := range grants {
		if grantIsManaOnly(gr.Key) {
			out = append(out, gr)
		}
	}
	return out
}

// grantIsManaOnly reports whether the catalog bundle `key` adds at least
// one mana ability and no other ability.
func grantIsManaOnly(key string) bool {
	if CatalogManaAbilities == nil || len(CatalogManaAbilities(key)) == 0 {
		return false
	}
	if CatalogTriggers != nil && len(CatalogTriggers(key)) > 0 {
		return false
	}
	if CatalogActivatedAbilities != nil && len(CatalogActivatedAbilities(key)) > 0 {
		return false
	}
	if CatalogStaticAbilities != nil && len(CatalogStaticAbilities(key)) > 0 {
		return false
	}
	if CatalogReplacements != nil && len(CatalogReplacements(key)) > 0 {
		return false
	}
	return true
}
