package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soltari Champion — Creature — Soltari Soldier {2}{W}, 2/2 (Stronghold, S82
// OLD b750, #5627):
//
//	"Shadow
//	 Whenever this creature attacks, other creatures you control get +1/+1
//	 until end of turn."
//
// Shadow is an engine keyword carried as printed data (CR 702.28). The attack
// trigger is Bolt Hound's at +1/+1: the creatures are chosen once, as it
// resolves (CR 611.2c), and the Champion itself is left out.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "765017e5-b43d-4c77-a9a9-0fcf0c0cd098",
		Name:            "Soltari Champion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Soltari Champion — other creatures you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID)), Power: 1, Toughness: 1,
						Label: "Soltari Champion — other creatures you control get +1/+1",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
