package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bolt Hound — Creature — Elemental Dog {2}{R}, 2/2 (Core Set 2021, #5511):
//
//	"Haste
//	 Whenever this creature attacks, other creatures you control get +1/+0
//	 until end of turn."
//
// Syr Alin's attack trigger at +1/+0: the creatures are chosen once, as
// it resolves (CR 611.2c), and the Hound itself is left out.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "60555f06-dc85-4a34-ae3c-fc1cdb923be7",
		Name:            "Bolt Hound",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Bolt Hound — other creatures you control get +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match: And(Creature(), YouControl(), OtherThan(item.SourceCardID)), Power: 1,
						Label: "Bolt Hound — other creatures you control get +1/+0",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
