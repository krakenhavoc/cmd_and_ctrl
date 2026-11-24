package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap2_b04_b2_helpers.go — the card-facing halves of #6087's two seams.

// CreaturesWithPowerAboveYourHandCantAttack is "Creatures with power greater
// than the number of cards in your hand can't attack" (Ensnaring Bridge,
// CR 508.1c). Every creature, any controller's, the printing permanent's
// controller's own included, carries the restriction with this permanent as
// its source; "your" is that permanent's controller, read when the layer
// pass runs. The creature's power and the hand size are read live at
// declaration (game.AttackTargetRestriction.PowerAtMostHandOf), so a card
// drawn or a pump cast before attackers are declared counts.
func CreaturesWithPowerAboveYourHandCantAttack(label string) game.StaticAbility {
	return game.StaticAbility{
		Label: label,
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return source != nil && target.IsCreature()
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			if source == nil {
				return
			}
			c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, game.AttackTargetRestriction{
				Source:            source.InstanceID,
				SourceName:        source.Name,
				PowerAtMostHandOf: source.Controller,
			})
		},
	}
}

// MustBeBlockedBySubtypeWhere gives every creature `appliesTo` matches "must
// be blocked by a <subtype> if able" (CR 509.1c; Ace's Baseball Bat's
// "must be blocked by a Dalek if able"): one requirement on the attacker,
// obeyed only by a blocker with that subtype (changeling included), while
// any other creature may still block it. The subtype is a word of the
// source's text (CR 612.2), so a text-changing effect on the source moves it.
func MustBeBlockedBySubtypeWhere(subtype string, appliesTo func(target *game.Card, g *game.Game, source *game.Card) bool) game.StaticAbility {
	return game.StaticAbility{
		Label: "Creatures it affects must be blocked by " + game.BlockerFilterNoun(game.BlockerFilterSubtype(subtype)) + " if able.",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && appliesTo(target, g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			r := game.BlockRequirement{Kind: game.BlockRequirementMustBeBlocked, Filter: game.BlockerFilterSubtype(subtype)}
			if source != nil {
				r.Filter = game.BlockerFilterSubtype(source.TextSubtype(subtype))
				r.Source, r.SourceName = source.InstanceID, source.Name
			}
			c.BlockRequirements = append(c.BlockRequirements, r)
		},
	}
}
