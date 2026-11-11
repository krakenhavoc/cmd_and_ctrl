package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Temple Altisaur — Creature — Dinosaur {4}{W}, 3/4 (Lost Caverns of Ixalan
// Commander):
//
//	"If a source would deal damage to another Dinosaur you control, prevent
//	 all but 1 of that damage."
//
// Hyperion's prevention (CR 615) over the controller's OTHER Dinosaurs: the
// Altisaur itself is not protected, and an opponent's Dinosaur is not.
func init() {
	Register(Spec{
		OracleID:     "1f499d39-b963-45c0-9604-d3cab98c9e05",
		Name:         "Temple Altisaur",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			preventAllButOneOfTheDamage(
				"Temple Altisaur — prevent all but 1 of that damage",
				"Prevent all but 1 of the damage dealt to another Dinosaur you control.",
				func(target uuid.UUID, g *game.Game, src *game.Card) bool {
					if target == src.InstanceID {
						return false
					}
					c, ok := g.LookupCardForEffect(target)
					return ok && g.Battlefield.Contains(target) && c.Controller == src.Controller &&
						c.HasSubtype(src.TextSubtype("Dinosaur"))
				}),
		},
	})
}
