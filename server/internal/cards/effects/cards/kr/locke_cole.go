package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Locke Cole — Legendary Creature — Human Rogue {1}{U}{B}, 2/3 (Final
// Fantasy, #4462):
//
//	"Deathtouch, lifelink
//	 Whenever Locke Cole deals combat damage to a player, draw a card,
//	 then discard a card."
//
// Deathtouch and lifelink are printed keywords. The loot is a draw whose
// discard opens after the draw lands (CR 608.2c, b16DrawThenDiscard), the
// discard the controller's choice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bdcb7aed-3595-4c7e-b9da-543e92de919a",
		Name:            "Locke Cole",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Locke Cole — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return effects.B16DrawThenDiscard(g, item, 1) }),
		},
	})
}
