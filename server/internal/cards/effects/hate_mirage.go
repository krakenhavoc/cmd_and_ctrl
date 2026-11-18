package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hate Mirage — Sorcery {3}{R}:
//
//	"Choose up to two target creatures you don't control. For each of those
//	 creatures, create a token that's a copy of that creature. Those tokens
//	 gain haste. Exile them at the beginning of the next end step."
//
// Both copies are made by ONE creation, one group per target
// (CreateTokenCopiesTogether, CR 608.2f), you control them, and each takes the
// target's copiable values (CR 707.2). "Those tokens gain haste" is a grant on
// each token, not an exception to the copy (CR 707.9a), as on Kindred Charge.
// The exile is a CR 603.7 delayed trigger over exactly the tokens made. A
// target that is illegal as the spell resolves is not copied (CR 608.2b); zero
// targets is legal and makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2db23ceb-448b-45d6-a931-d72e8ccdb541",
		Name:         "Hate Mirage",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to two target creatures you don't control", OpponentControls()).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var originals []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					originals = append(originals, t.ID)
				}
			}
			if len(originals) == 0 {
				return nil
			}
			return CreateTokenCopiesTogether{
				Controller: ctx.Controller(),
				Copies:     originals,
				Then:       ctx.ItemTail(hateMirageHasteAndExile),
			}.Apply(ctx)
		},
	})
}

// hateMirageHasteAndExile is "those tokens gain haste. Exile them at the
// beginning of the next end step", told the tokens that landed. An on-disk key.
var hateMirageHasteAndExile = CardTail("hate-mirage/haste-and-exile-copies", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		if err := (ScopedEffectFor{Target: id, Mods: []game.Mod{game.AddKeywordsMod("haste")},
			Duration: game.IndefiniteDuration(), Label: "Hate Mirage — haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return ScheduleDelayedTrigger{
		Label: "Hate Mirage — exile the tokens",
		Cards: o.IDs,
		Body:  exileListedCardsBody,
	}.Apply(ctx)
})
