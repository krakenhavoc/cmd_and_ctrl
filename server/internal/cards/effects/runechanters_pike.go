package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Runechanter's Pike — Artifact — Equipment {2}:
//
//	"Equipped creature has first strike and gets +X/+0, where X is the
//	 number of instant and sorcery cards in your graveyard.
//	 Equip {2}"
//
// X is read live by the layer 7c static, so it moves with the graveyard
// (CR 613.4c); "your" is the Equipment's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2efea460-c3e1-4578-a587-134e13bdf264",
		Name:         "Runechanter's Pike",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("first strike"),
			PumpAttachedPer(1, 0, func(g *game.Game, source *game.Card) int {
				return fandanielInstantsAndSorceries(g, source.Controller)
			}),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
