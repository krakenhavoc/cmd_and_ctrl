package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daring Saboteur — Creature — Human Pirate {1}{U}, 2/1:
//
//	"{2}{U}: This creature can't be blocked this turn.
//	 Whenever this creature deals combat damage to a player, you may draw
//	 a card. If you do, discard a card."
//
// The unblockable grant is the until-end-of-turn restriction registry
// (CR 509.1b), pinned to the permanent that activated it (CR 400.7). The
// loot is the shared "you may draw, then discard" prompt.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d36f64a0-53f7-4c56-9f40-a7927a5c2608",
		Name:         "Daring Saboteur",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{2}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Daring Saboteur — can't be blocked this turn"}.Apply(effects.NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Daring Saboteur — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Daring Saboteur — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     effects.CelestusAnswer,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
