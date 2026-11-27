package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cenn's Heir — Creature — Kithkin Soldier {1}{W}, 1/1 (Lorwyn, S82 OLD
// b545, #5593):
//
//	"Whenever this creature attacks, it gets +1/+1 until end of turn for
//	 each other attacking Kithkin."
//
// Counted as the trigger resolves, so a Kithkin declared in the same
// attack counts; every attacking Kithkin counts, whoever controls it, and
// a changeling is a Kithkin (CR 702.73a). A Heir that has left, or come
// back as a new object, gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42adcc63-cfb0-4b62-b8e1-dda01189652a",
		Name:         "Cenn's Heir",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Cenn's Heir — it gets +1/+1 until end of turn for each other attacking Kithkin",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) || g.AbilitySourceIsNewObjectForEffect(item) {
						return nil
					}
					n := effects.S82OldB545OtherAttackingOfSubtype(g, item.SourceCardID, "Kithkin")
					if n <= 0 {
						return nil
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n, Label: "Cenn's Heir — +1/+1 for each other attacking Kithkin"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
