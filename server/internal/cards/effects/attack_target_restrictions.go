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
// ADR 0107 §2 (#1879) adds the third form, "This creature can't attack
// unless defending player controls <a permanent>", as data the engine
// asks of each target's defending player (CR 508.5):
//
//	Static: []game.StaticAbility{CantAttackUnlessDefendingPlayerControls(QuerySubtype("Island"))}, // Sea Serpent
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
	return selfAttackTargetRestriction(game.AttackTargetRestriction{NotOwner: true, NotOwnersPlaneswalkers: planeswalkers})
}

// CantAttackUnlessDefendingPlayerControls is "This creature can't attack
// unless defending player controls <a permanent>" (ADR 0107 §2, #1879,
// CR 508.1c). The queries are any-of: Godhunter Octopus's "an enchantment
// or an enchanted permanent" is two of them.
//
//	Static: []game.StaticAbility{CantAttackUnlessDefendingPlayerControls(QuerySubtype("Island"))}, // Sea Serpent
//
// The engine works out the defending player of each target (CR 508.5: the
// player, a planeswalker's controller, a battle's protector), so in
// Commander the creature may attack the opponents who control an Island
// and nobody else. Like the owner clauses it is the creature's own ability
// and goes with its abilities (CR 613.1f).
func CantAttackUnlessDefendingPlayerControls(queries ...game.PermanentQuery) game.StaticAbility {
	return selfAttackTargetRestriction(game.AttackTargetRestriction{DefenderMustControl: queries})
}

// QuerySubtype is "a permanent with this subtype": an Island, a Mountain.
func QuerySubtype(subtype string) game.PermanentQuery {
	return game.PermanentQuery{Subtypes: []string{subtype}}
}

// selfAttackTargetRestriction writes r onto the creature whose ability it
// is, naming that creature as the source.
func selfAttackTargetRestriction(r game.AttackTargetRestriction) game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && selfOnly(target, g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			r := r
			if source != nil {
				r.Source, r.SourceName = source.InstanceID, source.Name
			}
			c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, r)
		},
	}
}
