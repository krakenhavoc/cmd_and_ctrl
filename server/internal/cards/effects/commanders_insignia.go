package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Commander's Insignia — Enchantment {2}{W}{W} (Foundations Commander):
//
//	"Creatures you control get +1/+1 for each time you've cast your commander
//	 from the command zone this game."
//
// A layer 7c static over every creature its controller has. The count is
// Player.CommanderCasts summed over every commander the controller owns
// (Hatut Zeraze Strike Force's read, partner commanders included, the ruling):
// it is bumped when the cast is made, so a commander spell that was countered
// or is still on the stack counts, and it never goes down. It is re-read on
// every layer pass, so the bonus follows the number, and "you" is the
// Insignia's controller (CR 109.5). A player who has never cast a commander
// gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "979a5d92-1c67-4fa4-b932-638b0cc1adcd",
		Name:         "Commander's Insignia",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Creatures you control get +1/+1 for each time you've cast your commander from the command zone this game.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				if p := g.PlayerByIDForEffect(source.Controller); p != nil {
					for _, casts := range p.CommanderCasts {
						n += casts
					}
				}
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
