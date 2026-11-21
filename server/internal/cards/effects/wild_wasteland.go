package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wild Wasteland — Enchantment {2}{R} (Fallout):
//
//	"Skip your draw step.
//	 At the beginning of your upkeep, exile the top two cards of your
//	 library. You may play those cards this turn."
//
// Damia's skip-draw-step replacement (CR 614.10) beside Charred Foyer's
// upkeep impulse exile with two cards: the exile grants the controller a
// play permission (land drop included) on exactly those cards, until end of
// turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4297ab35-9f16-4d9e-af77-010469ec4f59",
		Name:         "Wild Wasteland",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SkipYourDrawStep()},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Wild Wasteland — exile the top two cards of your library; you may play those cards this turn",
				func(g *game.Game, item *game.StackItem) error {
					return ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 2}.Apply(NewContext(g, item))
				}),
		},
	})
}
