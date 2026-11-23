package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Netherese Puzzle-Ward — Enchantment {3}{U} (Forgotten Realms Commander,
// #5681):
//
//	"Focus Beam — At the beginning of your upkeep, roll a d4. Scry X,
//	 where X is the result.
//	 Perfect Illumination — Whenever you roll a die's highest natural
//	 result, draw a card."
//
// Focus Beam rolls through the dice door, so a replaced roll (Pixie
// Guide's extra die) settles before the scry, and X is the result of the
// kept die.
//
// Perfect Illumination reads the natural result, the number the die shows
// before any modifier (CR 706.2): a 4 on a d4, a 20 on a d20, whatever
// the instruction adds or subtracts. It triggers once for each kept die
// that shows its highest face (YouRolledADiesHighestNaturalResult), its
// own d4 included, and never for an ignored roll (CR 706.6) or a planar
// die, which has no number.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "561e8ecc-1f2b-4fd4-8f92-a0acea449816",
		Name:         "Netherese Puzzle-Ward",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Netherese Puzzle-Ward — roll a d4, then scry X, where X is the result",
				func(g *game.Game, item *game.StackItem) error {
					return effects.RollDice{Sides: 4, N: 1, Then: netheresePuzzleWardScry}.Apply(effects.NewContext(g, item))
				}),
			effects.On(game.EventRollDie, effects.YouRolledADiesHighestNaturalResult,
				"Netherese Puzzle-Ward — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

// netheresePuzzleWardScry is Focus Beam after its d4: scry the result.
var netheresePuzzleWardScry = effects.CardTail("dice/netherese-puzzle-ward-scry", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Rolls) == 0 {
		return nil
	}
	return effects.Scry{Player: ctx.Controller(), N: effects.DiceResult(o)}.Apply(ctx)
})
