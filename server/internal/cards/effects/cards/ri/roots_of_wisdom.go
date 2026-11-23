package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roots of Wisdom — Sorcery {1}{G} (Jumpstart 2022, #5948):
//
//	"Mill three cards, then return a land card or Elf card from your
//	 graveyard to your hand. If you can't, draw a card."
//
// Corpse Churn's mill-then-return with a mandatory pick: when the
// graveyard holds a land card or an Elf card (read after the mill, the
// type read through this spell's text, CR 612.2) the controller must return
// one; when it holds neither they draw a card instead.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f6af612-a4a1-4165-ae4d-ff6353cbaf8d",
		Name:         "Roots of Wisdom",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.MillCards{N: 3}.Apply(ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				elf := ctx.TextSubtype("Elf")
				match := func(c game.Card) bool { return c.IsLand() || c.HasSubtype(elf) }
				if len(effects.GraveyardCardIDs(ctx, ctx.Controller(), match)) == 0 {
					return effects.DrawCards{N: 1}.Apply(ctx)
				}
				return effects.ReturnChosenFromGraveyard{
					Player:   ctx.Controller(),
					Question: "Roots of Wisdom — return a land card or Elf card from your graveyard to your hand",
					Match:    match,
					Min:      1,
					Max:      1,
					ToHand:   true,
				}.Apply(ctx)
			}),
		),
	})
}
