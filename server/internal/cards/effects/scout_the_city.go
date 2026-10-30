package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ecb8e283-5cdb-4e6a-bdff-546c1617363c",
		Name:         "Scout the City",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Look Around — Mill three cards. You may put a permanent card from among them into your hand. You gain 3 life.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					if err := (MillToZone{N: 3, ThenTail: ctx.ItemTail(scoutTheCityMilled)}).Apply(ctx); err != nil {
						return err
					}
					return GainLife{Amount: 3}.Apply(ctx)
				}),
			ModeDoing("Bring Down — Destroy target creature with flying.",
				TargetCreature("target creature with flying", HasKeyword("flying")),
				DestroyTheModesTarget),
		),
	})
}

var scoutTheCityMilled = CardTail("scout-the-city/after-mill", func(ctx *Context, o game.TailOutcome) error {
	return mayTakeOneFromAmongThem(ctx, o.IDs, Permanent(),
		"Scout the City — you may put a permanent card from among them into your hand")
})
