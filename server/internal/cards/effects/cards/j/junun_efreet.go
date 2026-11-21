package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Junún Efreet — Creature — Efreet {1}{B}{B}, 3/3 (Arabian Nights,
// #5641):
//
//	"Flying
//	 At the beginning of your upkeep, sacrifice this creature unless you
//	 pay {B}{B}."
//
// Krosan Cloudscraper's upkeep pay-or-sacrifice: UpkeepPayUnless holds
// the upkeep while the controller answers (#997), and a decline, or a
// "yes" they can't fund, sacrifices the Efreet.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "afda663e-c5f7-4182-86f7-d95d71793717",
		Name:            "Junún Efreet",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Junún Efreet — sacrifice unless you pay {B}{B}", func(g *game.Game, item *game.StackItem) error {
				return effects.UpkeepPayUnless{
					Chooser:  item.Controller,
					Cost:     "{B}{B}",
					Question: "Junún Efreet — pay {B}{B} or sacrifice Junún Efreet?",
					Else:     effects.UpkeepSacrificeSource,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
