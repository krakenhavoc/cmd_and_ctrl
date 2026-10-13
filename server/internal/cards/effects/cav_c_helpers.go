package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cav_c_helpers.go — shared pieces of the #3172 slice that made
// "tapped and attacking" tokens and "attacking creatures have …" statics
// honest.

// triggerAttackTarget is what the attacker whose EventAttack queued this
// trigger was attacking: a player, a planeswalker or a battle. It is
// uuid.Nil for a trigger with no attack event behind it. Read it as the
// Prefer of CreateTokensAttackingYourChoice, which asks about it first.
func triggerAttackTarget(item *game.StackItem) uuid.UUID {
	if item == nil || item.Trigger == nil {
		return uuid.Nil
	}
	return item.Trigger.Event.Target
}

// AttackingCreaturesHave is "attacking creatures you control have
// <keywords>" (Ohran Frostfang's shape): a layer 6 grant gated on
// Card.AttackingTarget, with DependsOnAttackingStatus so a declaration,
// a creature put onto the battlefield attacking, a removal from combat
// and the end of combat all rebuild the cached characteristics. Match
// narrows which of the controller's creatures qualify (nil = any).
func AttackingCreaturesHave(label string, match func(*game.Card) bool, keywords ...string) game.StaticAbility {
	return game.StaticAbility{
		Layer:                    game.Layer6Ability,
		Label:                    label,
		DependsOnAttackingStatus: true,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			if !target.IsCreature() || target.Controller != source.Controller || target.AttackingTarget == uuid.Nil {
				return false
			}
			return match == nil || match(target)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			appendKeywordsTo(c, keywords)
		},
	}
}
