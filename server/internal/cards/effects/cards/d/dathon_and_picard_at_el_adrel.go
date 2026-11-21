package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "76b7e1ee-43f0-4b56-822a-71b064c9201b",
		Name:         "Dathon and Picard at El-Adrel",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Temba, His Arms Wide — Target creature gets +3/+3 and gains trample until end of turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.ID == uuid.Nil {
						return nil
					}
					return effects.UntilEndOfTurn(ctx, t.ID, nil, "Temba, His Arms Wide — +3/+3 and trample",
						game.ModifyPTMod(3, 3), game.AddKeywordsMod("trample"))
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "trample"))),
			effects.ModeDoing("Shaka, When the Walls Fell — Each opponent sacrifices a nontoken artifact or enchantment of their choice.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.EachPlayerSacrifices{
						ExceptController: true,
						Match:            effects.And(effects.Or(effects.Artifact(), effects.Enchantment()), effects.Not(effects.IsTokenPredicate())),
						Label:            "a nontoken artifact or enchantment",
					}.Apply(ctx)
				}),
		),
	})
}
