package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// From Under the Floorboards — Sorcery {3}{B}{B} (Commander 2019):
//
//	"Madness {X}{B}{B} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)
//	 Create three tapped 2/2 black Zombie creature tokens and you gain 3 life.
//	 If this spell's madness cost was paid, instead create X of those tokens and
//	 you gain X life."
//
// Madness is the engine's machinery, declared once with Spec.Madness, with the
// X announced for the madness cast (Avacyn's Judgment's shape: the printed cost
// has no {X}, so CR 107.3b does not lock it). The branch reads which alternative
// cost the spell paid (Context.PaidAltCost, CR 702.35b): three Zombies and 3
// life, or X of each. The tokens enter tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1fc4269-f3e4-4a59-849d-aa1727eef23a",
		Name:         "From Under the Floorboards",
		Completeness: effects.CompletenessFull,
		Madness:      "{X}{B}{B}",
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerValue, LifeGain: 3},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 3
			if ctx.PaidAltCost(game.AltCostKeyMadness) {
				n = ctx.X()
			}
			if n > 0 {
				if _, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(effects.TokenCard("2/2 black Zombie")), n, game.TokenEntryOptions{Tapped: true}); err != nil {
					return err
				}
			}
			return effects.GainLife{Amount: n}.Apply(ctx)
		},
	})
}
