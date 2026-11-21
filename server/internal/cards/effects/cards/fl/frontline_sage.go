package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frontline Sage — Creature — Human Wizard {2}{U}, 0/1 (Conflux, #5587):
//
//	"Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)
//	 {U}, {T}: Draw a card, then discard a card."
//
// Exalted is the engine's keyword (game/exalted.go, CR 702.83a). The
// loot is Erratic Visionary's body at {U}: the draw, then a discard the
// controller picks (CR 701.9a). The {T} cost waits out summoning
// sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b0f6130c-58f7-4693-894c-598cd9a38ce7",
		Name:            "Frontline Sage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, Discards: 1},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.LootOne(g, item, 1)
			},
		}},
	})
}
