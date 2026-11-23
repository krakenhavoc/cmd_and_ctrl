package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Temmet, Naktamun's Will — Legendary Creature — Zombie Wizard {2}{W}{U}{B},
// 4/4:
//
//	"Vigilance, menace
//	 Whenever you attack, draw a card, then discard a card.
//	 Whenever you draw a card, Zombies you control get +1/+1 until end of
//	 turn."
//
// "Whenever you attack" is one trigger per declaration (trekB40YouAttacked);
// the loot's discard is the draw's continuation, so it is chosen after the
// draw. Each card drawn is its own trigger, the loot's draw and the draw step
// included, and each locks in the Zombies you control as it resolves
// (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3dc43ffe-f6b0-4cdc-86b3-01c0f41df0ae",
		Name:            "Temmet, Naktamun's Will",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace"},
		Triggered: []game.TriggeredAbility{
			effects.TrekB40YouAttacked("Temmet, Naktamun's Will — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DrawThenDiscard(g, item.Controller, item.SourceCardID, 1, 1, "Temmet, Naktamun's Will — discard a card")
				}),
			effects.WheneverYouDraw("Temmet, Naktamun's Will — Zombies you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl(), effects.HasSubtype("Zombie")), Power: 1, Toughness: 1,
						Label: "Temmet, Naktamun's Will — Zombies get +1/+1"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
