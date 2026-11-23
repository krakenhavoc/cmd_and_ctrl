package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cloudkill — Sorcery {4}{B}{B} (Commander Legends: Battle for Baldur's Gate,
// S82 GAP #5948):
//
//	"All creatures get -X/-X until end of turn, where X is the greatest mana
//	 value of a commander you own on the battlefield or in the command zone."
//
// X is counted once as the spell resolves, over the commanders the caster owns
// on the battlefield or in the command zone (a commander that is a token copy
// is not one; a card in hand or graveyard is not counted), and every creature
// then on the battlefield, whoever controls it, gets -X/-X for the turn
// (CR 611.2c). With no commander X is 0 and nothing changes.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1941239b-e131-4bde-b493-1c8c805b6a3c",
		Name:         "Cloudkill",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := effects.GreatestOwnedCommanderManaValue(ctx.Game, item.Controller)
			if x <= 0 {
				return nil
			}
			return effects.BoostUntilEOT{Match: effects.Creature(), Power: -x, Toughness: -x, Label: "Cloudkill — -X/-X"}.Apply(ctx)
		},
	})
}
