package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Atemsis, All-Seeing — Legendary Creature — Sphinx {3}{U}{U}{U}, 4/5 (Core Set
// 2020, #5518):
//
//	"Flying
//	 {2}{U}, {T}: Draw two cards, then discard a card.
//	 Whenever Atemsis deals damage to an opponent, you may reveal your hand. If
//	 cards with at least six different mana values are revealed this way, that
//	 player loses the game."
//
// The trigger is one per damage event dealt to an opponent (combat or not). The
// hand is revealed as it resolves and the different mana values among its cards
// are counted then; the damaged player loses only with six or more.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f557aa26-8b79-4116-9249-a067b21b88ef",
		Name:            "Atemsis, All-Seeing",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}: Draw two cards, then discard a card.",
			Purpose: game.Purpose{Draws: 2, Discards: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawThenDiscard(g, item.Controller, item.SourceCardID, 2, 1, "")
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverThisDealsDamageToAnOpponent("Atemsis, All-Seeing — you may reveal your hand; with six different mana values, that player loses the game",
				effects.OldB154AtemsisRevealForSix), "Atemsis, All-Seeing — reveal your hand?"),
		},
	})
}
