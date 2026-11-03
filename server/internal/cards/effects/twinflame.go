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
// spell resolves is not copied (CR 608.2b). Each copy takes the creature's
// copiable values (CR 707.2) with haste added to its own keywords, and the
// exile is a CR 603.7 delayed trigger over exactly the tokens this spell made.
// Kindred Charge's copy-and-exile body, over the chosen creatures instead of a
// type.
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
			return socB14HastyCopiesExiledAtEnd(ctx, ctx.Controller(), originals, "Twinflame — exile the tokens")
		},
	})
}
