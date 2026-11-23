package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infinite Guideline Station — Legendary Artifact — Spacecraft
// {W}{U}{B}{R}{G}, 7/15 (Edge of Eternities, #4461):
//
//	"When Infinite Guideline Station enters, create a tapped 2/2
//	 colorless Robot artifact creature token for each multicolored
//	 permanent you control.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 12+.)
//	 12+ | Flying
//	 Whenever Infinite Guideline Station attacks, draw a card for each
//	 multicolored permanent you control."
//
// Both counts are the multicolored permanents you control as the
// ability resolves (Multicolored reads the layered colours), the Station
// itself among them while it is on the battlefield. The Robots are
// Maco Strike Team's tapped tokens. Station, the 12+ creature threshold
// and its keyword are the shared constructors (Galvanizing Sawship); the
// attack trigger only matters at 12+, when it can attack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cda39514-86c9-483e-8d95-fd08138ce33a",
		Name:         "Infinite Guideline Station",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Infinite Guideline Station — create a tapped 2/2 Robot for each multicolored permanent you control",
				func(g *game.Game, item *game.StackItem) error {
					n := len(effects.PermanentsControlledByMatching(g, item.Controller, effects.Multicolored()))
					if n == 0 {
						return nil
					}
					return effects.CreateTokenAdvanced{Spec: effects.TokenSpec{Template: effects.TokenCard("2/2 colorless Robot artifact"), Tapped: true}, N: n}.Apply(effects.NewContext(g, item))
				}),
			effects.WheneverThisAttacks("Infinite Guideline Station — draw a card for each multicolored permanent you control",
				func(g *game.Game, item *game.StackItem) error {
					n := len(effects.PermanentsControlledByMatching(g, item.Controller, effects.Multicolored()))
					return effects.DrawCards{Player: item.Controller, N: n}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(12, 7, 15),
			effects.ThresholdKeywords(12, "flying"),
		),
	})
}
