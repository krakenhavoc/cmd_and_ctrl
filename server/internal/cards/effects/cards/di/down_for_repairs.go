package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Down for Repairs — Sorcery {2}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card. Destroy up to one target
//	 Attraction that player controls. (It's put into their junkyard.)"
//
// The revealed-hand pick (ADR 0116), then the destruction. The second
// clause is related to the first (ControlledByTargetsController, whose
// anchor is a player here, as Ravager of the Fells' is): the Attraction
// must be controlled by the targeted opponent as the spell is cast (CR
// 601.2c) and again as it resolves (CR 608.2b). "Up to one" lets the
// spell be cast with no Attraction at all.
//
// The destruction waits for the choice from the revealed hand (CR
// 608.2c, ResolveSteps). The Attraction goes to its owner's junkyard
// because the rules put a card with the Astrotorium back there (CR
// 717.6, ADR 0207); the parenthetical is reminder text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "455dc001-3f6d-4d44-afa7-0f83a2520baf",
		Name:         "Down for Repairs",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetPlayer("target opponent", effects.Opponent()),
			effects.ControlledByTargetsController(effects.TargetPermanent("up to one target Attraction that player controls", effects.Attraction()).WithCount(0, 1), 0),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.ChooseNonlandFromTargetedHand),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(1)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
			}),
		),
	})
}
