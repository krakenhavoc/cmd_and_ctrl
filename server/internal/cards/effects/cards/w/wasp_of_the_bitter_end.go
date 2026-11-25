package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wasp of the Bitter End — {1}{B} Creature — Insect Horror 2/1 (Hour of Devastation):
//
//	"Flying
//	 Whenever you cast a Bolas planeswalker spell, you may sacrifice this
//	 creature. If you do, destroy target creature."
//
// The target is chosen as the trigger goes on the stack; the "may" is asked as
// it resolves, and the destruction follows only a completed sacrifice.
func init() {
	bolasWalker := func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsPlaneswalker() && c.HasSubtype("Bolas")
	}
	effects.Register(effects.Spec{
		OracleID:        "9331cfd7-a818-45ed-9e4a-4d3c381052e0",
		Name:            "Wasp of the Bitter End",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.WheneverYouCast(bolasWalker, "Wasp of the Bitter End — sacrifice it, then destroy target creature", waspOfTheBitterEndSacrifice),
					"Sacrifice Wasp of the Bitter End to destroy target creature?"),
				effects.TargetCreature("target creature"),
			),
		},
	})
}

var waspOfTheBitterEndSacrifice = effects.SacrificeThisThen("then/wasp-of-the-bitter-end-destroy", func(ctx *effects.Context) error {
	return effects.DestroyFirstLegalCardTarget(ctx.Game, ctx.Item)
})
