package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Insatiable Appetite — Instant {1}{G}:
//
//	"You may sacrifice a Food. If you do, target creature gets +5/+5 until
//	 end of turn. Otherwise, that creature gets +3/+3 until end of turn."
//
// The target is chosen as the spell is cast; the sacrifice is asked as it
// resolves (maySacrificeOne, CR 608.2c). The creature gets +3/+3 as it
// resolves; a Food that really went (CR 701.21a) adds another +2/+2 on top,
// so the creature ends the turn with exactly +5/+5, and a controller who
// declines, or has no Food, gets the +3/+3 alone. A target that is illegal
// on resolution means the spell does nothing, so no Food is asked for
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53570a4e-a750-4776-bc52-0025cd510a35",
		Name:         "Insatiable Appetite",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetCard {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: legal[0].ID, Power: 3, Toughness: 3, Label: "Insatiable Appetite — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			return effects.MaySacrificeOne(ctx.Game, item, "Insatiable Appetite — sacrifice a Food to give the creature +5/+5 instead?",
				effects.HasSubtype("Food"), false, insatiableAppetiteSacrificed)
		},
	})
}

// insatiableAppetiteSacrificed adds the remaining +2/+2 once a Food has
// really been sacrificed, keyed (#4290).
var insatiableAppetiteSacrificed = effects.RegisterSacrificedOne("prompt-run/insatiable-appetite-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	legal := ctx.LegalTargets()
	if len(legal) == 0 || legal[0].Kind != game.TargetCard {
		return nil
	}
	return effects.BoostUntilEOT{Target: legal[0].ID, Power: 2, Toughness: 2, Label: "Insatiable Appetite — +5/+5 (Food sacrificed)"}.Apply(ctx)
})
