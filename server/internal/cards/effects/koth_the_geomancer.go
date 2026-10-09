package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Koth, the Geomancer — Legendary Creature — Human Warrior {2}{R},
// 3/2:
//
//	"Reach
//	 Landfall — Whenever a land you control enters, Koth deals 1 damage
//	 to each opponent. If that land is a Mountain, add {R}."
//
// The damage comes from Koth. The Mountain check reads the entering land
// as the ability resolves (it is the same card even if it has left the
// battlefield since), and the {R} is added as the ability resolves, so it
// empties with the step's mana pool (CR 106.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3b17da84-ae4d-4fc1-a327-85e54e6156c6",
		Name:            "Koth, the Geomancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			Landfall("Koth, the Geomancer — 1 damage to each opponent; if that land is a Mountain, add {R}",
				func(g *game.Game, item *game.StackItem) error {
					if err := damageToEachOpponent(g, item, 1); err != nil {
						return err
					}
					if !rfCreatureCEnteredLandHasSubtype(g, item, "Mountain") {
						return nil
					}
					return AddMana{Produced: "{R}"}.Apply(NewContext(g, item))
				}),
		},
	})
}
