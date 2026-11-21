package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "53501112-bcdb-4bb5-9fd3-674cdba91951",
		Name:            "Crookclaw Transmuter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Crookclaw Transmuter — switch target creature's power and toughness",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SwitchEachTargetPT(effects.NewContext(g, item), "Crookclaw Transmuter — switch power and toughness")
				}), effects.TargetCreature("target creature")),
		},
	})
}
