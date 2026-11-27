package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Root Greevil — Creature — Beast {3}{G}, 2/3 (Planeshift, S82 OLD b644, #5615):
//
//	"{2}{G}, {T}, Sacrifice this creature: Destroy all enchantments of the color of
//	 your choice."
//
// Wash Out's colour choice: the controller names a colour as the ability resolves
// (the sacrifice and the tap are costs, so the choice is made with the Greevil gone),
// then every enchantment of that colour, every player's, is destroyed together
// (CR 700.4). Colourless enchantments are never of a chosen colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e43b25be-fe67-4eaa-80df-85718f38f904",
		Name:         "Root Greevil",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, {T}, Sacrifice this creature: Destroy all enchantments of the color of your choice.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy, Partial: true}},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				effects.ChooseColorTail(game.ColorForHarm, g, item.Controller, item.SourceCardID, "Root Greevil — choose a color", ctx.ItemTail(rootGreevilColorChosen))
				return nil
			},
		}},
	})
}

// rootGreevilColorChosen destroys every enchantment of the chosen colour. An
// on-disk key.
var rootGreevilColorChosen = effects.CardTail("then/root-greevil-color-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.DestroyAllMatching{Match: effects.And(effects.Enchantment(), effects.OfColor(o.Tail.Word))}.Apply(ctx)
})
