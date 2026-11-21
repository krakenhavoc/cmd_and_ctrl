package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yavimaya Iconoclast — Creature — Elf {1}{G}, 3/2 (Dominaria United,
// #5487):
//
//	"Kicker {R} (You may pay an additional {R} as you cast this spell.)
//	 Trample
//	 When this creature enters, if it was kicked, it gets +1/+1 and gains
//	 haste until end of turn."
//
// "It" is the Iconoclast itself, so both halves are pinned to it; a
// Iconoclast that has left the battlefield by resolution gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9900c771-ab81-415e-90a7-123c1bacfa83",
		Name:            "Yavimaya Iconoclast",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{R}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Yavimaya Iconoclast — kicked: it gets +1/+1 and gains haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.ThisGetsUntilEndOfTurn(1, 1, "Yavimaya Iconoclast — +1/+1")(g, item); err != nil {
						return err
					}
					return effects.ThisGainsKeywordUntilEndOfTurn("haste", "Yavimaya Iconoclast — haste")(g, item)
				}),
		},
	})
}
