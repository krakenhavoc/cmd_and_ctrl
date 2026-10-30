package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harmonized Crescendo — Instant {4}{U}{U} (Lorwyn Eclipsed, #4458):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Choose a creature type. Draw a card for each permanent you control
//	 of that type."
//
// Distant Melody at instant speed, with convoke. The type is chosen as
// the spell resolves (#2382) and the count read from the board when the
// answer arrives: PERMANENTS of the type, so a changeling and a Kindred
// permanent count, and so do the creatures tapped to convoke it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b93cd38a-1d0c-40e6-aaf1-621a2c88bf46",
		Name:         "Harmonized Crescendo",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "Harmonized Crescendo — choose a creature type",
				ctx.ItemTail(harmonizedCrescendoTypeChosen))
			return nil
		},
	})
}

// harmonizedCrescendoTypeChosen draws a card for each permanent of the
// chosen type the caster controls. An on-disk key.
var harmonizedCrescendoTypeChosen = creatureTypeChosen("then/harmonized-crescendo-type-chosen", func(ctx *Context, t string, _ game.TailOutcome) error {
	controller := ctx.Controller()
	return DrawCards{Player: controller, N: permanentsOfTypeYouControl(ctx.Game, controller, t)}.Apply(ctx)
})
