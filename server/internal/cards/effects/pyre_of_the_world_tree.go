package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     invasionOfKaldheimOracleID + "#1",
		Name:         "Pyre of the World Tree",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Discard a land card: This enchantment deals 2 damage to any target.",
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Cost:    DiscardCardsMatching(1, "a land card", MatchLand),
			Targets: TargetAny(),
			Effect:  DealDamageToTheTarget(2),
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, YouDiscardedALandCard, "Pyre of the World Tree — exile the top card of your library; you may play it this turn", Do(
				Instruction(func(ctx *Context) error {
					return ExileTopWithPermission{From: ctx.Controller(), GrantTo: ctx.Controller(), N: 1}.Apply(ctx)
				}),
			)),
		},
	})
}
