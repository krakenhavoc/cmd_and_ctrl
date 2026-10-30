package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// First Family — Instant {2}{G}{U} (Marvel Super Heroes Commander,
// #4455):
//
//	"You draw X cards and gain X life, where X is the number of colors
//	 among permanents you control and spells you've cast this turn."
//
// X is the union of the colors of your permanents (Faeburrow Elder's
// read) and of every spell you cast this turn (CastTally.Colors, read off
// each spell as it became cast, CR 601.2i), so First Family itself, being
// green and blue, always counts. The count is a single number (CR 107.3),
// read when the spell resolves; the life is gained after the draw, each
// step waiting for the prompt the one before asks (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e119c709-bf9d-451e-827b-9787d455029a",
		Name:         "First Family",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return DrawCards{Player: ctx.Item.Controller, N: msc17FirstFamilyX(ctx.Game, ctx.Item.Controller)}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				return GainLife{Player: ctx.Item.Controller, Amount: msc17FirstFamilyX(ctx.Game, ctx.Item.Controller)}.Apply(ctx)
			}),
		),
	})
}

// msc17FirstFamilyX is the number of distinct colors among the player's
// permanents and the spells they have cast this turn.
func msc17FirstFamilyX(g *game.Game, you uuid.UUID) int {
	seen := map[string]bool{}
	for _, col := range colorsAmongPermanentsYouControl(g, you) {
		seen[col] = true
	}
	for _, col := range g.CastTallyFor(you).SpellColorsCast() {
		seen[col] = true
	}
	return len(seen)
}
