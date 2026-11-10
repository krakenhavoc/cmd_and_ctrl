package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Molten Duplication — Sorcery {1}{R} (Murders at Karlov Manor, #4469):
//
//	"Create a token that's a copy of target artifact or creature you
//	 control, except it's an artifact in addition to its other types. It
//	 gains haste until end of turn. Sacrifice it at the beginning of the
//	 next end step."
//
// Saheeli, the Sun's Brilliance's body (artifact exception on the copy,
// CR 707.9b), but the haste is a gain "until end of turn", not part of the
// copy, and the sacrifice is one delayed trigger for the next end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92b91910-9b65-4421-8eef-a33793161089",
		Name:         "Molten Duplication",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or creature you control", Or(Artifact(), Creature()), YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			copyOf := FirstLegalBattlefieldTarget(ctx)
			if copyOf == uuid.Nil {
				return nil
			}
			return CreateTokenCopiesTogether{
				Controller: item.Controller,
				Copies:     []uuid.UUID{copyOf},
				Except:     tokenCopyIsAlsoAnArtifact,
				Then:       ctx.ItemTail(moltenDuplicationRest),
			}.Apply(ctx)
		},
	})
}

// moltenDuplicationRest is "It gains haste until end of turn. Sacrifice
// it at the beginning of the next end step." An on-disk key.
var moltenDuplicationRest = CardTail("molten-duplication/haste-and-sacrifice-copy", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"},
			Label: "Molten Duplication — the token gains haste until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Item.Controller,
		Label:      "Molten Duplication — sacrifice the token",
		Cards:      o.IDs,
		Body:       sacrificeListedCardsBody,
	}.Apply(ctx)
})
