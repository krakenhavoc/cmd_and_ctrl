package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scout the City — Sorcery {1}{G} (Marvel Super Heroes Commander, #4455):
//
//	"Choose one —
//	 • Look Around — Mill three cards. You may put a permanent card from
//	   among them into your hand. You gain 3 life.
//	 • Bring Down — Destroy target creature with flying."
//
// Look Around is Eerie Gravestone's mill-then-take (MillToZone with a
// tail and mayTakeOneFromAmongThem) over permanent cards, then the life.
// Bring Down is Crushing Canopy's flying-creature clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ecb8e283-5cdb-4e6a-bdff-546c1617363c",
		Name:         "Scout the City",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Look Around — Mill three cards. You may put a permanent card from among them into your hand. You gain 3 life.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					if err := (effects.MillToZone{N: 3, ThenTail: ctx.ItemTail(scoutTheCityMilled)}).Apply(ctx); err != nil {
						return err
					}
					return effects.GainLife{Amount: 3}.Apply(ctx)
				}),
			effects.ModeDoing("Bring Down — Destroy target creature with flying.",
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
				effects.DestroyTheModesTarget),
		),
	})
}

var scoutTheCityMilled = effects.CardTail("scout-the-city/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.MayTakeOneFromAmongThem(ctx, o.IDs, effects.Permanent(),
		"Scout the City — you may put a permanent card from among them into your hand")
})
