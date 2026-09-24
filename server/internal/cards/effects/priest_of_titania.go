package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Priest of Titania — Creature — Elf Druid {1}{G}, 1/1:
//
//	"{T}: Add {G} for each Elf on the battlefield."
//
// "On the battlefield" — every player's, not just the controller's —
// so this counts every Elf at the table, itself included. Marwyn,
// the Nurturer's shape (fixed-colour, repeated by a count read at
// activation) with a wider count: strings.Repeat("{G}", n) rather
// than Marwyn's controller-scoped loop. A board with zero Elves
// still exists (the Priest was just cast and hasn't resolved into
// its own count yet an instant later): the ability adds nothing and
// still taps, same as Marwyn at power zero.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a198a16-17b9-481e-b516-5bc945c7e247",
		Name:         "Priest of Titania",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:  ManaAbilityCost{Tap: true},
			Label: "Add {G} for each Elf on the battlefield",
			ProducedFunc: func(g *game.Game, _, _ uuid.UUID) string {
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.HasSubtype("Elf") {
						n++
					}
				}
				return strings.Repeat("{G}", n)
			},
		}},
	})
}
