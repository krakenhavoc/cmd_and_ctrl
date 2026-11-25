package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crafty Cutpurse — Creature — Human Pirate {3}{U}, 2/2 (Commander 2021):
//
//	"Flash
//	 When this creature enters, each token that would be created under an
//	 opponent's control this turn is created under your control instead."
//
// A turn-scoped replacement (game.ModRedirectTokens, CR 614.1a, 701.7b)
// registered when the trigger resolves: every create-tokens instruction an
// opponent of the Cutpurse's controller makes until end of turn has its
// tokens minted, and entering, under that controller instead. The record
// belongs to the player, not the creature, so it outlasts the Cutpurse
// leaving the battlefield (CR 611.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2ef53e18-0baf-4eee-b671-4c19b464cbfb",
		Name:            "Crafty Cutpurse",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Crafty Cutpurse — each token an opponent would create this turn is created under your control instead",
				func(g *game.Game, item *game.StackItem) error {
					g.RedirectOpponentTokensThisTurnForEffect(item.SourceCardID, item.Controller, "Crafty Cutpurse — opponents' tokens are created under your control")
					return nil
				}),
		},
	})
}
