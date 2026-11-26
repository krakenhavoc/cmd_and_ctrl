package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smog Elemental — Creature — Elemental {4}{B}{B}, 3/3 (Gatecrash, S82 OLD
// #5565):
//
//	"Flying
//	 Creatures with flying your opponents control get -1/-1."
//
// Gravitational Shift's layer 7c reading of flying (read after layer 6, so a
// creature granted flying is hit and one that lost it is spared), narrowed to
// creatures an opponent of the Elemental's controller controls. The Elemental's
// own flying never counts against its controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8e51a07d-8618-4908-8798-bd28bc9b5bbb",
		Name:            "Smog Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return effects.CrosswindsCreaturesWithFlying(target, nil, source) && target.Controller != source.Controller
		}, -1, -1)},
	})
}
