package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nihil Spellbomb — Artifact {1}:
//
//	"{T}, Sacrifice this artifact: Exile target player's graveyard.
//	 When this artifact is put into a graveyard from the battlefield, you
//	 may pay {B}. If you do, draw a card."
//
// Tormod's Crypt's ability, plus Chromatic Star's graveyard trigger with
// a MayPay in front of the draw (Antagonist Incorporator's shape): the
// trigger fires however the Spellbomb died, and declining, or a pool that
// cannot pay {B}, draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b713e49f-1b13-42d1-91f2-cc7a579e7614",
		Name:         "Nihil Spellbomb",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: Exile target player's graveyard.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.ExileTargetPlayersGraveyard,
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Nihil Spellbomb — you may pay {B}; if you do, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{B}",
						Question: "Pay {B} to draw a card?",
						Then:     effects.RprB15SpellbombPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
