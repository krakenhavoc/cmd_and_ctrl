package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Rend — Instant {1}{B} (Mirage):
//
//	"Destroy target creature if it's white. A creature destroyed this way
//	 can't be regenerated.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// Any creature may be targeted; the colour is read as the spell resolves
// (CR 608.2b), so a creature that stopped being white is not destroyed. The
// draw is a delayed trigger set up as the spell resolves (CR 603.7) and
// happens whether or not the creature was white.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "80d233cc-7748-41a2-bfd6-2189d0136900",
		Name:         "Soul Rend",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok {
				if card, found := ctx.Game.LookupCardForEffect(t.ID); found && card.HasColorWord("W") {
					if err := effects.DestroyTheTargetPermanentNoRegen(item, ctx); err != nil {
						return err
					}
				}
			}
			return effects.DrawACardAtTheNextUpkeep("Soul Rend — draw a card").Apply(ctx)
		},
	})
}
