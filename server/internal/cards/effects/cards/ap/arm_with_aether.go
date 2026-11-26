package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arm with Aether — Sorcery {2}{U} (New Phyrexia, #5575):
//
//	"Until end of turn, creatures you control gain "Whenever this creature deals
//	 damage to an opponent, you may return target creature that player controls
//	 to its owner's hand.""
//
// Showstopper's mass grant: the creatures are fixed as the spell resolves
// (CR 611.2c) and each carries the trigger for the turn. The trigger fires for
// damage of any kind to an opponent, aims as it goes on the stack at a creature
// THAT player controls (read off the damage event) and asks "you may" after the
// pick (CR 603.5); with no creature to name it is not put on the stack.
//
// No simplification.
func init() {
	const grant = "arm-with-aether/bounce-on-damage"
	effects.Register(effects.Spec{
		OracleID:     "e6479195-cbdf-4d2b-9f45-5fd0a6ab4a77",
		Name:         "Arm with Aether",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Triggered: []game.TriggeredAbility{
				effects.Optional(game.TriggeredAbility{
					Watches:     []game.EventKind{game.EventDealDamage},
					AppliesTo:   effects.ThisDealtDamageToAnOpponent,
					TargetsFrom: effects.S82OldB458TargetCreatureThatPlayerControls,
					Key:         "Arm with Aether — return target creature that player controls to its owner's hand",
					Effect:      effects.BounceTheTargetToHand,
				}, "Arm with Aether — return target creature that player controls to its owner's hand?"),
			},
			Text: "Whenever this creature deals damage to an opponent, you may return target creature that player controls to its owner's hand.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantAbilitiesFor{
				Match: effects.And(effects.Creature(), effects.YouControl()),
				Keys:  []string{grant},
				Label: "Arm with Aether — creatures you control gain \"Whenever this creature deals damage to an opponent, you may return target creature that player controls to its owner's hand.\"",
			}.Apply(ctx)
		},
	})
}
