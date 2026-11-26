package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// One-Eyed Scarecrow — Artifact Creature — Scarecrow {3}, 2/3 (Innistrad,
// S82 older sets batch 449, #5572):
//
//	"Defender
//	 Creatures with flying your opponents control get -1/-0."
//
// Smog Elemental's flying filter over the opponents' creatures, a layer-7c
// boost that follows each creature's flying as it gains and loses it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "87b61c1f-18e3-446a-a732-cff40adc5e05",
		Name:            "One-Eyed Scarecrow",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return effects.CrosswindsCreaturesWithFlying(target, nil, source) && target.Controller != source.Controller
		}, -1, 0)},
	})
}
