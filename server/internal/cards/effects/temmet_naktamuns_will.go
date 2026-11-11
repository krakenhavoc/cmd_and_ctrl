package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "3dc43ffe-f6b0-4cdc-86b3-01c0f41df0ae",
		Name:            "Temmet, Naktamun's Will",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace"},
		Triggered: []game.TriggeredAbility{
			trekB40YouAttacked("Temmet, Naktamun's Will — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return drawThenDiscard(g, item.Controller, item.SourceCardID, 1, 1, "Temmet, Naktamun's Will — discard a card")
				}),
			WheneverYouDraw("Temmet, Naktamun's Will — Zombies you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{Match: And(Creature(), YouControl(), HasSubtype("Zombie")), Power: 1, Toughness: 1,
						Label: "Temmet, Naktamun's Will — Zombies get +1/+1"}.Apply(NewContext(g, item))
				}),
		},
	})
}
