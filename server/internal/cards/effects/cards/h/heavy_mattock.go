package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heavy Mattock — Artifact — Equipment {3}:
//
//	"Equipped creature gets +1/+1.
//	 As long as equipped creature is a Human, it gets an additional +1/+1.
//	 Equip {2}"
//
// The Human clause is its own static, re-read every layer pass.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1f45310-d665-400d-934f-e48e7858ce8c",
		Name:         "Heavy Mattock",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.OldB442PumpAttachedWhile(func(host *game.Card, _ *game.Game, source *game.Card) bool {
				return host.HasSubtypeWord(source.TextSubtype("Human"))
			}, 1, 1),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
