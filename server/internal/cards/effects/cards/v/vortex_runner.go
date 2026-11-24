package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vortex Runner — Creature — Human Wizard {2}{U}, 2/3 (Strixhaven: School of
// Mages, #5502):
//
//	"As long as you control eight or more lands, this creature gets +1/+0 and
//	 can't be blocked."
//
// Gigantoad's layer 7c self-pump and Expedition Lookout's conditional block
// rule, both reading the land count live (CR 604.2), so each comes and goes
// with the eighth land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37cf4ee6-e146-4354-afc8-6b59b0700e87",
		Name:         "Vortex Runner",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "As long as you control eight or more lands, this creature gets +1/+0.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.B10LandsControlled(g, source.Controller) >= 8
			},
			Apply: effects.PumpBy(1, 0),
		}},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnSelf(), func(g *game.Game, you uuid.UUID, _ game.Card) bool {
				return effects.B10LandsControlled(g, you) >= 8
			}),
		},
	})
}
