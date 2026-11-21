package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyre of the World Tree — Enchantment, the back face of Invasion of
// Kaldheim (March of the Machine, S82 batch 23):
//
//	"Discard a land card: This enchantment deals 2 damage to any target.
//	 Whenever you discard a land card, exile the top card of your library.
//	 You may play that card this turn."
//
// The discard is the cost, so it happens on activation and itself triggers
// the second ability, whose exile lands on the stack above the damage. Both
// read the discarded card as a land card in the graveyard, where it has
// its printed characteristics. The impulse exile is an ordinary play
// permission to the end of the turn (a land may be played).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     effects.InvasionOfKaldheimOracleID + "#1",
		Name:         "Pyre of the World Tree",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a land card: This enchantment deals 2 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.DiscardCardsMatching(1, "a land card", effects.MatchLand),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, effects.YouDiscardedALandCard, "Pyre of the World Tree — exile the top card of your library; you may play it this turn", effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.ExileTopWithPermission{From: ctx.Controller(), GrantTo: ctx.Controller(), N: 1}.Apply(ctx)
				}),
			)),
		},
	})
}
