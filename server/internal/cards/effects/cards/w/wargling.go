package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wargling — Creature — Wolf {1}{G}, 2/2 (The Hobbit, HOB):
//
//	"Ferocious — Whenever this creature attacks while you control a
//	 creature with power 4 or greater, until end of turn, this creature gets
//	 +1/+0 and creatures you control gain trample."
//
// The ferocious window is Nighthowl Pursuer's (hobbitFerociousAttack): the
// power check is made once, as the attack is declared. The creatures that
// gain trample are locked as the trigger resolves (CR 611.2c), as Craterclaw
// Colossus's are.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce107656-f222-4d9a-9ba2-b33b4d47bdb5",
		Name:         "Wargling",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.HobbitFerociousAttack, "Wargling — ferocious: +1/+0, and creatures you control gain trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Wargling — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"trample"},
						Label: "Wargling — creatures you control gain trample"}.Apply(ctx)
				}),
		},
	})
}
