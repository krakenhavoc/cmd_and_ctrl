package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giant-Man, Gargantuan Genius — {3}{G}{G} Legendary Creature — Human
// Scientist Hero 6/6 (Marvel Super Heroes Commander, #4455):
//
//	"Reach
//	 At the beginning of your first main phase, add {G} for each creature you
//	 control with power 4 or greater."
//
// Black Market's first-main-phase trigger. The count is read when the trigger
// RESOLVES, Giant-Man himself included; the mana empties at the end of the
// phase like any other (CR 106.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "effa84e6-0aee-422f-81e9-3feeca38eda2",
		Name:            "Giant-Man, Gargantuan Genius",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Giant-Man, Gargantuan Genius — add {G} for each creature you control with power 4 or greater",
				func(g *game.Game, item *game.StackItem) error {
					n := msc12CreaturesWithPowerAtLeast(g, item.Controller, 4)
					if n == 0 {
						return nil
					}
					return AddMana{Amount: ManaN(n, "G")}.Apply(NewContext(g, item))
				}),
		},
	})
}
