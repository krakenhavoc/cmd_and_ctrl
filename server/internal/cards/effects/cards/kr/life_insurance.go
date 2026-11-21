package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Life Insurance — Enchantment {3}{W}{B} (Duskmourn Commander, #5368):
//
//	"Extort (Whenever you cast a spell, you may pay {W/B}. If you do,
//	 each opponent loses 1 life and you gain that much life.)
//	 Whenever a nontoken creature dies, you lose 1 life and create a
//	 Treasure token."
//
// Extort is the shared keyword trigger (extort.go). The dies trigger is
// Blight Mound's condition without the controller clause: any nontoken
// creature that left the battlefield for a graveyard, read from its
// last-known information. The life loss is the Enchantment's controller's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50779f73-c538-4364-aca4-4ab069a00814",
		Name:         "Life Insurance",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Extort("Life Insurance"),
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !effects.IsToken(dead)
			}, "Life Insurance — lose 1 life, create a Treasure", effects.Do(
				effects.GainLife{Amount: -1},
				effects.CreateToken{Template: effects.TreasureToken(), N: 1},
			)),
		},
	})
}
