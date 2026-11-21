package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mistmeadow Witch — Creature — Kithkin Wizard {1}{W/U}, 1/1 (S82 GAP, #5948):
//
//	"{2}{W}{U}: Exile target creature. Return that card to the battlefield
//	 under its owner's control at the beginning of the next end step."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38e274e2-bd04-48de-a1df-44f0ee987ba8",
		Name:         "Mistmeadow Witch",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}{U}: Exile target creature. Return that card to the battlefield under its owner's control at the beginning of the next end step.",
			Cost:    effects.ManaCost("{2}{W}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileTargetsThenScheduleReturn(effects.NewContext(g, item), "Mistmeadow Witch — return the exiled creature")
			},
		}},
	})
}
