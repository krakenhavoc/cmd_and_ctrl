package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "50779f73-c538-4364-aca4-4ab069a00814",
		Name:         "Life Insurance",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Extort("Life Insurance"),
			On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && !IsToken(dead)
			}, "Life Insurance — lose 1 life, create a Treasure", Do(
				GainLife{Amount: -1},
				CreateToken{Template: TreasureToken(), N: 1},
			)),
		},
	})
}
