package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_target_restrictions.go — the card-facing half of ADR 0106 §2
// (#1794): a creature that "can't attack its owner". The engine half —
// the predicate every declaration-time caller shares, the enumerator's
// per-attacker target list, the exemptions for CR 508.7b and 508.4c —
// is game/attack_target_restrictions.go. A card file only says which
// form it prints:
//
//	Static: []game.StaticAbility{AttacksEachCombat(), CantAttackItsOwnerOrItsOwnersPlaneswalkers()}, // Xantcha, Sleeper Agent
//	Static: []game.StaticAbility{AttacksEachCombat(), CantAttackItsOwner()},                         // Alexios, Deimos of Kosmos
//
// Both are the creature's own ability, so they go with its abilities:
// a creature that loses all abilities may attack its owner (CR 613.1f —
// the catalog stops being asked once CatalogAbilityKey answers empty).
// The owner is read live by the engine, so a Clone copying Xantcha may
// not attack the Clone's owner rather than Xantcha's.
//
// "This creature can't attack its owner" GRANTED by a resolved effect
// (Elrond of the White Council) is not this: it needs a ScopedEffect
// mod kind that writes the same restriction, which ADR 0106 leaves out
// of scope.

// CantAttackItsOwner is "~ can't attack its owner".
func CantAttackItsOwner() game.StaticAbility {
	return cantAttackOwnerStatic(false)
}

// CantAttackItsOwnerOrItsOwnersPlaneswalkers is "~ can't attack its
// owner or planeswalkers its owner controls". A battle the owner
// protects is still a legal target.
func CantAttackItsOwnerOrItsOwnersPlaneswalkers() game.StaticAbility {
	return cantAttackOwnerStatic(true)
}

func cantAttackOwnerStatic(planeswalkers bool) game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && selfOnly(target, g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			r := game.AttackTargetRestriction{NotOwner: true, NotOwnersPlaneswalkers: planeswalkers}
			if source != nil {
				r.Source, r.SourceName = source.InstanceID, source.Name
			}
			c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, r)
		},
	}
}
