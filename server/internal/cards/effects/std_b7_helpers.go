package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b7_helpers.go — shared helpers of Standard set batches 7 and 8
// (#4452).

// SacrificeAnotherWithManaValueX is "Sacrifice a <permanent> you control
// with mana value X other than <this>" (Sidisi, Regent of the Mire): the
// source is no legal pick, the activator announces X (DemandsX asks for
// it as for an {X}), and the permanent sacrificed must have mana value X
// (game.SacrificeManaValueX). Pair it with a target clause bound to the
// same X (TargetSpec.WithManaValueEqualsXPlus).
func SacrificeAnotherWithManaValueX(label string, preds ...CardPredicate) game.AbilityCost {
	spec := Another(sacrificeSpec(label, preds...)).WithCount(1, 1)
	spec.ManaValueEqualsX = true
	return game.AbilityCost{SacrificeOther: spec}
}

// GrantedAltCostFreeDragonSpells is Dracogenesis's offer key: "You may
// cast Dragon spells without paying their mana costs." Its own key, not
// Omniscience's "granted-free": a table with both has two free offers whose
// reach differs, and a claim is stored by key.
const GrantedAltCostFreeDragonSpells = grantedAltCostKeyPrefix + "free-dragon-spells"

// CastDragonSpellsWithoutPayingManaCost is Dracogenesis's static. Omniscience's
// free cast narrowed to spells with the Dragon creature type, from every zone
// a spell is cast from (CR 118.9a: wherever its mana cost could be paid). An
// {X} is 0 (CR 107.3b) and additional costs are still paid (CR 118.9d); the
// spell keeps its own timing.
func CastDragonSpellsWithoutPayingManaCost() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:   GrantedAltCostFreeDragonSpells,
			Label: "Cast it without paying its mana cost",
		},
		Spells: game.PermissionFilter{CreatureType: "Dragon"},
	}
}
