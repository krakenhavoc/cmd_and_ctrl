package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nighthowl Pursuer — Creature — Wolf {B}, 1/1 (The Hobbit):
//
//	"Menace
//	 Ferocious — Whenever this creature attacks while you control a
//	 creature with power 4 or greater, this creature gets +2/+2 until
//	 end of turn."
//
// "While" is a condition on the trigger event, not an intervening "if":
// it is read once, as the attack is declared, and the pump resolves
// whatever the board looks like by then. The attacker itself counts
// toward the power check.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "89aa97d0-1aef-4ed8-b29c-7943bb3d0e60",
		Name:            "Nighthowl Pursuer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, hobbitFerociousAttack, "Nighthowl Pursuer — ferocious: +2/+2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{Target: item.SourceCardID, Power: 2, Toughness: 2, Label: "Nighthowl Pursuer — +2/+2"}.Apply(NewContext(g, item))
				}),
		},
	})
}
