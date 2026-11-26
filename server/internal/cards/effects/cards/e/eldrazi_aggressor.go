package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eldrazi Aggressor — {2}{R} Creature — Eldrazi Drone 2/3 (Oath of the
// Gatewatch, S82 OLD b296):
//
//	"Devoid (This card has no color.)
//	 This creature has haste as long as you control another colorless
//	 creature."
//
// Bloodghast's conditional haste: a layer-6 keyword grant to the creature
// itself, recomputed on every pass, so it follows the board. The other creature
// must be colorless as the layers have it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aa0c6fea-c90c-4b65-aa1e-18e5b34ea6e3",
		Name:            "Eldrazi Aggressor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && effects.OldB295YouControlAnotherColorlessCreature(g, source)
			}, "haste"),
		},
	})
}
