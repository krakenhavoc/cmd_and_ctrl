package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Twinflame — Sorcery {1}{R} (Secrets of Strixhaven Commander, #4834):
//
//	"Strive — This spell costs {2}{R} more to cast for each target beyond
//	 the first.
//	 Choose any number of target creatures you control. For each of them,
//	 create a token that's a copy of that creature, except it has haste.
//	 Exile those tokens at the beginning of the next end step."
//
// Strive is the coloured per-target surcharge of Call the Coppercoats (CR
// 207.2c): the {R} of each surcharge joins the cost as a coloured symbol. Zero
// targets is legal and makes nothing. A target that is no longer legal as the
// spell resolves is not copied (CR 608.2b).
//
// Every copy is made by ONE creation, one group per target (#4594, #4945,
// CreateTokenCopiesTogether): the tokens are created at the same time (CR
// 608.2f) and enter together, so each newcomer sees the others enter (CR
// 603.6a). Each copy takes the creature's copiable values (CR 707.2), and
// "except it has haste" is part of the copying process, so haste is one of the
// copy's copiable values (CR 707.9a) and a copy of the token has haste too.
// That is unlike Kindred Charge's "those tokens gain haste", a grant. The exile
// is a CR 603.7 delayed trigger over exactly the tokens this spell made,
// scheduled by the creation's keyed tail, so a copy whose entry asks a question
// (Adaptive Automaton's creature type) is still exiled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83cf1169-5853-4332-b897-7b17d72d76ab",
		Name:         "Twinflame",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("any number of target creatures you control", YouControl()).WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			CostsMorePerTargetBeyondFirst("{2}{R}", "Strive — This spell costs {2}{R} more to cast for each target beyond the first."),
		},
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
				Except:     TokenCopyGainsHaste,
				Then:       ctx.ItemTail(twinflameExileCopies),
			}.Apply(ctx)
		},
	})
}

// twinflameExileCopies is "exile those tokens at the beginning of the next end
// step", told the tokens that landed. An on-disk key.
var twinflameExileCopies = CardTail("twinflame/exile-copies", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Twinflame — exile the tokens",
		Cards: o.IDs,
		Body:  exileListedCardsBody,
	}.Apply(ctx)
})
