package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crookclaw Transmuter — Creature — Bird Wizard {3}{U}, 3/1:
//
//	"Flash
//	 Flying
//	 When this creature enters, switch target creature's power and
//	 toughness until end of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "53501112-bcdb-4bb5-9fd3-674cdba91951",
		Name:            "Crookclaw Transmuter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Crookclaw Transmuter — switch target creature's power and toughness",
				func(g *game.Game, item *game.StackItem) error {
					return switchEachTargetPT(NewContext(g, item), "Crookclaw Transmuter — switch power and toughness")
				}), TargetCreature("target creature")),
		},
	})
}
