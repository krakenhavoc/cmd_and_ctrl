package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "cda39514-86c9-483e-8d95-fd08138ce33a",
		Name:         "Infinite Guideline Station",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Infinite Guideline Station — create a tapped 2/2 Robot for each multicolored permanent you control",
				func(g *game.Game, item *game.StackItem) error {
					n := len(permanentsControlledByMatching(g, item.Controller, Multicolored()))
					if n == 0 {
						return nil
					}
					return CreateTokenAdvanced{Spec: TokenSpec{Template: TokenCard("2/2 colorless Robot artifact"), Tapped: true}, N: n}.Apply(NewContext(g, item))
				}),
			WheneverThisAttacks("Infinite Guideline Station — draw a card for each multicolored permanent you control",
				func(g *game.Game, item *game.StackItem) error {
					n := len(permanentsControlledByMatching(g, item.Controller, Multicolored()))
					return DrawCards{Player: item.Controller, N: n}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(12, 7, 15),
			ThresholdKeywords(12, "flying"),
		),
	})
}
