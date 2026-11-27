package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Broodwarden — Creature — Eldrazi Drone {3}{G}{G}, 4/4 (Rise of the
// Eldrazi, S82 OLD batch 483, #5581):
//
//	"Eldrazi Spawn creatures you control get +2/+1."
//
// A layer 7c anthem over creatures its controller controls that are both
// Eldrazi and Spawn (the Spawn tokens); other Eldrazi, the Broodwarden
// included, are not Spawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14537413-3628-4152-b23f-78b0e9c416ac",
		Name:         "Broodwarden",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.HasSubtype(source.TextSubtype("Eldrazi")) && target.HasSubtype(source.TextSubtype("Spawn"))
			}, 2, 1),
		},
	})
}
