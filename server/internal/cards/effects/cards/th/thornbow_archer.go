package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thornbow Archer — Creature — Elf Archer {B}, 1/2 (Magic Origins, S82 older
// sets):
//
//	"Whenever this creature attacks, each opponent who doesn't control an Elf
//	 loses 1 life."
//
// "An Elf" is any permanent with the subtype. Each opponent is judged as the trigger resolves (CR 608.2), by the Elf
// subtype read through the layered view (a changeling counts, CR 702.73).
// Life loss, not damage, so no prevention applies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9b35457-0f1c-495c-a508-6f6693835051",
		Name:         "Thornbow Archer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Thornbow Archer — each opponent who doesn't control an Elf loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, opp := range ctx.Opponents() {
						if effects.CountControlled(g, opp, func(c game.Card) bool { return effects.CardHasSubtype(c, ctx.TextSubtype("Elf")) }) > 0 {
							continue
						}
						if err := g.ChangePlayerLifeForEffect(ctx.Source(), opp, -1); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
