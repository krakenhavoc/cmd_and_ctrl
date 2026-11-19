package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tribune of Rot — Creature — Elf Shaman {1}{B/G}{B/G}, 3/3 (Ravnica: Clue
// Edition, S82):
//
//	"Whenever this creature attacks, mill two cards. For each creature card
//	 milled this way, create a 1/1 green Saproling creature token."
//
// Old Rutstein's shape: the mill's own tail reads back the cards that were
// really milled, and each creature card among them makes one Saproling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd1b3ef3-73ac-4f06-8cfb-5f2388481379",
		Name:         "Tribune of Rot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Tribune of Rot — mill two cards; for each creature card milled this way, create a 1/1 green Saproling creature token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: 2, ThenTail: ctx.ItemTail(tribuneOfRotMilled)}.Apply(ctx)
				}),
		},
	})
}

// tribuneOfRotMilled is the clause after the mill: one Saproling for each
// creature card milled this way. An on-disk key.
var tribuneOfRotMilled = CardTail("tribune-of-rot/after-mill", func(ctx *Context, o game.TailOutcome) error {
	n := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 green Saproling"), N: n}.Apply(ctx)
})
