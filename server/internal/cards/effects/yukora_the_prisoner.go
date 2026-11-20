package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yukora, the Prisoner — Legendary Creature — Demon Spirit {2}{B}{B}, 5/5
// (Betrayers of Kamigawa, #5602):
//
//	"When Yukora leaves the battlefield, sacrifice all non-Ogre creatures
//	 you control."
//
// The leave trigger fires however Yukora leaves (dies, exile, bounce). As
// it resolves, every creature its controller then controls that is not
// an Ogre is sacrificed in one simultaneous batch (b08SacrificeAllMatching),
// so a death watcher sees them all. Indestructible does not save one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "10026343-a8cd-4a99-b4d9-ff7611f52052",
		Name:         "Yukora, the Prisoner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisLeaves("Yukora, the Prisoner — sacrifice all non-Ogre creatures you control",
				func(g *game.Game, item *game.StackItem) error {
					return b08SacrificeAllMatching(NewContext(g, item), And(Creature(), YouControl(), Not(Subtype("Ogre"))))
				}),
		},
	})
}
