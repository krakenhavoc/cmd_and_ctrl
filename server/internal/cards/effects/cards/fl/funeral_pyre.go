package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Funeral Pyre — Instant {W} (Judgment):
//
//	"Exile target card from a graveyard. Its owner creates a 1/1 white Spirit
//	 creature token with flying."
//
// The exile, then the token for the card's OWNER, not the caster, read off the
// card once the exile is done (ExileTarget's tail, so a move a replacement
// paused or redirected is waited for). The tail runs whether or not the card
// reached exile, since the second sentence does not depend on it. With the one
// target illegal the spell does nothing at all (CR 608.2b), so the token comes
// only with a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02af53c1-83e6-447f-908b-2757c8aa445c",
		Name:         "Funeral Pyre",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1, CreatureTokens: 1},
		Targets:      effects.TargetCardInGraveyard("target card from a graveyard"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return effects.ExileTarget{Target: t.ID, ThenTail: ctx.ItemTail(funeralPyreOwnerGetsASpirit)}.Apply(ctx)
				}
			}
			return nil
		},
	})
}

// funeralPyreOwnerGetsASpirit is the second sentence: the exiled card's
// owner, read off the card wherever it now is, creates the token. The tail is
// told the outcome of the exile, and runs whether or not the card reached
// exile, since the sentence does not depend on it. An on-disk key: never
// renamed, never reused.
var funeralPyreOwnerGetsASpirit = effects.CardTail("tail/s82-old-b628-funeral-pyre-spirit", func(ctx *effects.Context, _ game.TailOutcome) error {
	if ctx.Item == nil || len(ctx.Item.Targets) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(ctx.Item.Targets[0].ID)
	if !ok || c.Owner == uuid.Nil {
		return nil
	}
	return effects.CreateToken{Template: effects.B28WhiteSpiritFlyingToken(), N: 1, Controller: c.Owner}.Apply(ctx)
})
