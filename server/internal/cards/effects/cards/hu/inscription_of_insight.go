package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inscription of Insight — Sorcery {3}{U} (Zendikar Rising, S82 OLD #5508):
//
//	"Kicker {2}{U}{U}
//	 Choose one. If this spell was kicked, choose any number instead.
//	 • Return up to two target creatures to their owners' hands.
//	 • Scry 2, then draw two cards.
//	 • Target player creates an X/X blue Illusion creature token, where X is
//	   the number of cards in their hand."
//
// A kicked cast chooses any number of the modes (CR 700.2d) and they
// resolve in printed order. X is counted as the mode resolves, in the hand
// of the player it targets, and the token is created under that player's
// control (CR 111.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "1551fe97-437e-4e89-a0d2-c398cb2155e4",
		Name:          "Inscription of Insight",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}{U}{U}")},
		Modes: effects.ChooseOne(
			effects.ModeDoing("Return up to two target creatures to their owners' hands.",
				effects.TargetCreature("up to two target creatures").WithCount(0, 2),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.BounceToHand{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Scry 2, then draw two cards.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.Scry{Player: ctx.Controller(), N: 2, ThenTail: ctx.TailN(effects.ThenDrawCards, 2)}.Apply(ctx)
				}), game.Purpose{Scry: 2, Draws: 2}),
			effects.ModeDoing("Target player creates an X/X blue Illusion creature token, where X is the number of cards in their hand.",
				effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if t.Kind != game.TargetPlayer {
							continue
						}
						p := ctx.Game.PlayerByIDForEffect(t.ID)
						if p == nil || p.Eliminated {
							return nil
						}
						x := p.Hand.Size()
						return effects.CreateToken{
							Controller: t.ID,
							Template: game.Card{
								Name: "Illusion", TypeLine: "Token Creature — Illusion",
								Colors: []string{"U"}, Power: x, Toughness: x,
							},
							N: 1,
						}.Apply(ctx)
					}
					return nil
				}),
		).AnyNumberIf(effects.WasKicked),
	})
}
