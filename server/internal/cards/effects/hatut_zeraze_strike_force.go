package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hatut Zeraze Strike Force — Creature — Human Spy Warrior {3}{W}, 2/2
// (Marvel Super Heroes Commander, #4455):
//
//	"When you cast this spell, copy it for each time you've cast your
//	 commander from the command zone this game.
//	 When this creature enters, destroy up to one target artifact or
//	 enchantment."
//
// The tally is Player.CommanderCasts summed over every commander the
// player owns (Jeska, Thrice Reborn's read; only a cast from the command
// zone bumps it, and it is never reduced). The copies are spell copies
// (CR 707.10) that become tokens as they resolve, and each enters and
// triggers the destroy on its own. A player who has never cast a
// commander copies nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92ac307e-6531-4465-8870-07988458f5d4",
		Name:         "Hatut Zeraze Strike Force",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Hatut Zeraze Strike Force — copy this spell for each time you've cast your commander from the command zone",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					if p := g.PlayerByIDForEffect(item.Controller); p != nil {
						for _, casts := range p.CommanderCasts {
							n += casts
						}
					}
					if n == 0 {
						return nil
					}
					return CopySpell{StackID: item.SourceCardID, Count: n}.Apply(NewContext(g, item))
				}),
			Targeting(WhenThisEnters("Hatut Zeraze Strike Force — destroy up to one target artifact or enchantment", func(g *game.Game, item *game.StackItem) error {
				return destroyEachLegalTarget(item, NewContext(g, item))
			}),
				TargetPermanent("up to one target artifact or enchantment", Or(Artifact(), Enchantment())).WithCount(0, 1)),
		},
	})
}
