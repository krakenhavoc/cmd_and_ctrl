package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mouth of the Storm — Creature — Elemental {6}{U}, 6/6 (Edge of
// Eternities, #4461):
//
//	"Flying
//	 Ward {2}
//	 When this creature enters, creatures your opponents control get -3/-0
//	 until your next turn."
//
// Bewildering Blizzard's power-only shrink over the creatures your
// opponents control, locked as the trigger resolves (CR 611.2c) so a
// creature that enters later keeps its power, lasting until the
// controller's next turn begins (CR 611.2b) instead of the end of this one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "832fbb1f-0ff5-4d6c-b1ee-2231923a6f64",
		Name:            "Mouth of the Storm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Mouth of the Storm — ward {2}"),
			effects.WhenThisEnters("Mouth of the Storm — creatures your opponents control get -3/-0 until your next turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ScopedEffectFor{
						Match:    effects.And(effects.Creature(), effects.OpponentControls()),
						Mods:     []game.Mod{game.ModifyPTMod(-3, 0)},
						Duration: effects.DurationUntilYourNextTurn(ctx, item.Controller),
						Label:    "Mouth of the Storm — creatures your opponents control get -3/-0",
					}.Apply(ctx)
				}),
		},
	})
}
