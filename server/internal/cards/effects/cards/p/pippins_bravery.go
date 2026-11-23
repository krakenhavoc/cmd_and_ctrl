package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pippin's Bravery — Instant {G} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"You may sacrifice a Food. If you do, target creature gets +4/+4 until
//	 end of turn. Otherwise, that creature gets +2/+2 until end of turn."
//
// Insatiable Appetite's shape. The creature gets +2/+2 as the spell resolves,
// then the controller may sacrifice a Food (no question with none to
// sacrifice), and a Food that really went adds another +2/+2: +4/+4 in all
// and +2/+2 otherwise. No player holds priority between the two, so the total
// is what "+4/+4 instead" reads as (CR 611.2c for the set of creature, which
// is the one target).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7ec52b1-295a-4daa-86fd-dcbe6fb412d7",
		Name:         "Pippin's Bravery",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetCard {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: legal[0].ID, Power: 2, Toughness: 2, Label: "Pippin's Bravery — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			return effects.MaySacrificeOne(ctx.Game, item, "Pippin's Bravery — sacrifice a Food to give the creature +4/+4 instead of +2/+2?",
				effects.HasSubtype("Food"), false, pippinsBraverySacrificed)
		},
	})
}

var pippinsBraverySacrificed = effects.RegisterSacrificedOne("prompt-run/pippins-bravery-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	legal := ctx.LegalTargets()
	if len(legal) == 0 || legal[0].Kind != game.TargetCard {
		return nil
	}
	return effects.BoostUntilEOT{Target: legal[0].ID, Power: 2, Toughness: 2, Label: "Pippin's Bravery — +4/+4 (Food sacrificed)"}.Apply(ctx)
})
