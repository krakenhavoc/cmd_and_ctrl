package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crypt Feaster — Creature — Zombie {3}{B}, 3/4 (Foundations, #4465):
//
//	"Menace
//	 Threshold — Whenever this creature attacks, if there are seven or
//	 more cards in your graveyard, this creature gets +2/+0 until end of
//	 turn."
//
// Menace rides PrintedKeywords. The attack trigger carries an
// intervening "if" (CR 603.4), Dawnstrike Vanguard's shape: the seven
// cards are counted as the attack is declared and again as the trigger
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4c22257e-8a00-40b7-a29b-55fcec7ddbd4",
		Name:            "Crypt Feaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "there are seven or more cards in your graveyard",
				Test: func(ctx *effects.Context) bool { return effects.FraThresholdMet(ctx.Game, ctx.Controller()) },
			}, effects.On(game.EventAttack, effects.AllOf(effects.ThisAttacked, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.FraThresholdMet(g, source.Controller)
			}), "Crypt Feaster — threshold: gets +2/+0 until end of turn",
				effects.ThisCreatureUntilEOT("Crypt Feaster — +2/+0", 2, 0))),
		},
	})
}
