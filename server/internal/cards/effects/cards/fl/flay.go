package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flay — Sorcery {3}{B} (Prophecy, #5617):
//
//	"Target player discards a card at random. Then that player discards
//	 another card at random unless they pay {1}."
//
// The first discard happens, then the player is asked for {1} and the table
// waits for the answer (#5793, CR 118.12); a decline is the second random
// discard. A player with an empty hand has nothing to discard, so is not
// asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "987b74f4-02cb-4d47-94e5-d58a3c8e92e4",
		Name:         "Flay",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: effects.ResolveSteps(
			effects.TargetPlayerDiscardsAtRandom{N: 1},
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					if p := ctx.Game.PlayerByIDForEffect(t.ID); p == nil || p.Hand.Size() == 0 {
						return nil
					}
					return effects.PayUnless{
						Chooser:  t.ID,
						Cost:     "{1}",
						Question: "Flay — pay {1}, or discard another card at random?",
						Else:     flayDiscard,
					}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}

// flayDiscard is the decline: the second random discard. Keyed, an on-disk
// identity.
var flayDiscard = effects.PayUnlessElse("pay-unless/flay-discard",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		for _, t := range ctx.Targets() {
			if t.Kind == game.TargetPlayer {
				return effects.DiscardCards{Player: t.ID, N: 1}.Apply(ctx)
			}
		}
		return nil
	})
