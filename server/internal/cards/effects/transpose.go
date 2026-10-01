package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Transpose — Instant {2}{B}:
//
//	"Draw a card, then discard a card. You lose 1 life. If this spell
//	 was cast from your hand, create a 0/1 black Wizard creature token
//	 with "Whenever you cast a noncreature spell, this token deals 1
//	 damage to each opponent."
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The life loss and the token wait for the discard (DiscardPrompt.Then),
// in printed order. "Cast from your hand" is the zone the spell was
// cast from (StackItem.CastFromZone), so the rebound cast from exile
// makes no Wizard. The Wizard is the catalog's printed token
// (NoncreatureCastWizardToken), ability and all.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "77d5f298-06b4-49d1-9b33-f1f176665ba2",
		Name:            "Transpose",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			me := ctx.Controller()
			if err := (DrawCards{Player: me, N: 1}).Apply(ctx); err != nil {
				return err
			}
			fromHand := item.CastFromZone == game.ZoneHand
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player: me,
				Source: item.SourceCardID,
				N:      1,
				Then: func(g *game.Game, _ uuid.UUID, _ []uuid.UUID) error {
					ctx := NewContext(g, item)
					if err := g.ChangePlayerLifeForEffect(ctx.Source(), me, -1); err != nil {
						return err
					}
					if !fromHand {
						return nil
					}
					return CreateToken{Controller: me, Template: NoncreatureCastWizardToken(), N: 1}.Apply(ctx)
				},
			})
			return nil
		},
	})
}
