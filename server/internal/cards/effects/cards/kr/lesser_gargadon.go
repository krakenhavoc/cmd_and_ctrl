package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lesser Gargadon — Creature — Beast {2}{R}{R}, 6/4 (Eighth Edition, #5607):
//
//	"Whenever this creature attacks or blocks, sacrifice a land."
//
// One trigger for both events (Runaway Carriage's shape): it fires once
// per attack declaration and once per block declaration, and its controller
// chooses the land as it resolves.
//
// No simplification.
func init() {
	const label = "Lesser Gargadon — sacrifice a land"
	effects.Register(effects.Spec{
		OracleID:     "7fe21ade-01f1-45ff-9b70-bc6302c3d284",
		Name:         "Lesser Gargadon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
			}, label, effects.S82OldB614SacrificeALand(label)),
		},
	})
}
