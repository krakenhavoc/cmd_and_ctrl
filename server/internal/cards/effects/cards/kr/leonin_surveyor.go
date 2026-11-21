package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leonin Surveyor — Creature — Cat Scout {1}{W}, 2/2:
//
//	"Start your engines!
//	 During your turn, this creature has first strike.
//	 Max speed — {3}, Exile this card from your graveyard: Draw a card."
//
// ADR 0138 (#2711). The first strike is a layer 6 grant on the
// controller's own turn, read at every layer pass (Gideon Blackblade's
// condition); the graveyard ability is the shared Surveyor body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "97c29db0-b3e1-4704-afe8-5218250c044c",
		Name:            "Leonin Surveyor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
			}, "first strike"),
		},
		Activated: []effects.ActivatedAbility{effects.MaxSpeedSurveyorDraw()},
	})
}
