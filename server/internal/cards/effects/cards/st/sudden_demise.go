package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sudden Demise — Sorcery {X}{R} (Commander 2013, S82 OLD b386):
//
//	"Choose a color. Sudden Demise deals X damage to each creature of the
//	 chosen color."
//
// The colour is chosen as the spell resolves (Wash Out's prompt); the
// continuation then deals X damage, one damage event, to every creature of
// that colour. A multicoloured creature is "of" each of its colours; a
// colourless one never is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b34b5b3f-7f17-4292-814e-634408a5d7a5",
		Name:         "Sudden Demise",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.ChooseColorTail(game.ColorForHarm, ctx.Game, item.Controller, item.SourceCardID, "Sudden Demise — choose a color",
				ctx.ItemTail(suddenDemiseColorChosen))
			return nil
		},
	})
}

// suddenDemiseColorChosen deals X damage to each creature of the colour (the
// tail's Word). An on-disk key.
var suddenDemiseColorChosen = effects.CardTail("then/sudden-demise-color-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.OfColor(o.Tail.Word)), ctx.X())
})
