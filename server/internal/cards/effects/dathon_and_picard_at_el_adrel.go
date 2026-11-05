package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dathon and Picard at El-Adrel — Instant {2}{G} (Star Trek, #4973):
//
//	"Choose one —
//	 • Temba, His Arms Wide — Target creature gets +3/+3 and gains
//	   trample until end of turn.
//	 • Shaka, When the Walls Fell — Each opponent sacrifices a nontoken
//	   artifact or enchantment of their choice."
//
// Temba is Origin of the Hulk's chapter III on any creature, re-checked
// at resolution (CR 608.2b). Shaka is Lorehold Charm's untargeted edict
// over nontoken artifacts and enchantments, each opponent choosing their
// own (CR 701.21a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "76b7e1ee-43f0-4b56-822a-71b064c9201b",
		Name:         "Dathon and Picard at El-Adrel",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Temba, His Arms Wide — Target creature gets +3/+3 and gains trample until end of turn.",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.ID == uuid.Nil {
						return nil
					}
					return untilEndOfTurn(ctx, t.ID, nil, "Temba, His Arms Wide — +3/+3 and trample",
						game.ModifyPTMod(3, 3), game.AddKeywordsMod("trample"))
				}), ForTargets(BenefitUntilEOT(0, 3, 3, "trample"))),
			ModeDoing("Shaka, When the Walls Fell — Each opponent sacrifices a nontoken artifact or enchantment of their choice.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return EachPlayerSacrifices{
						ExceptController: true,
						Match:            And(Or(Artifact(), Enchantment()), Not(IsTokenPredicate())),
						Label:            "a nontoken artifact or enchantment",
					}.Apply(ctx)
				}),
		),
	})
}
