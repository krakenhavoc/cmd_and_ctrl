package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Breathe Your Last — Instant {1}{B}{B} (Modern Horizons 3, #5467):
//
//	"Destroy target creature or planeswalker. You gain 1 life for each
//	 of its colors."
//
// Hero's Downfall with a life rider. The target's colours are read
// through the layers just before the destroy, which is the permanent as
// it last existed on the battlefield (CR 608.2h): a colour-changing
// effect counts, and a colorless target gives no life. The life is
// gained whether or not the destroy happened (an indestructible target
// is still a legal one). A target gone by resolution makes the spell do
// nothing at all (CR 608.2b), so no life is gained either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc3af7cb-03b9-445f-82eb-18798ec92934",
		Name:         "Breathe Your Last",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				card, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				colors := len(card.EffectiveColors())
				if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				return GainLife{Amount: colors}.Apply(ctx)
			}
			return nil
		},
	})
}
