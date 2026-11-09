package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Lord Master of Hell — the back face of The Emperor of Palamecia (Final
// Fantasy, #4462): Legendary Creature — Demon Noble Wizard, 3/3.
//
//	"Starfall — Whenever The Lord Master of Hell attacks, it deals X damage
//	 to each opponent, where X is the number of noncreature, nonland cards
//	 in your graveyard."
//
// X is counted once, as the ability resolves (the rulings), and the damage
// is one event from the Lord Master to every opponent. A creature or land
// card is not counted; a token never is, being no card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     theEmperorOfPalameciaOracleID + "#1",
		Name:         "The Lord Master of Hell",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("The Lord Master of Hell — Starfall: it deals X damage to each opponent, where X is the number of noncreature, nonland cards in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					x := 0
					if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Graveyard != nil {
						for _, c := range p.Graveyard.Cards {
							if !c.IsCreature() && !c.IsLand() {
								x++
							}
						}
					}
					return DamageEachOpponent{N: x}.Apply(NewContext(g, item))
				}),
		},
	})
}
