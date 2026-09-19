package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mahadi, Emporium Master — Legendary Creature — Devil {1}{B}{R}, 3/3
// (EDHREC rank 1262):
//
//	"At the beginning of your end step, create a Treasure token for
//	 each creature that died this turn. (It's an artifact with "{T},
//	 Sacrifice this token: Add one mana of any color.")"
//
// The aristocrats deck's end-of-turn payout. "Each creature that died
// this turn" counts every player's creatures, tokens included, and is
// evaluated when the trigger RESOLVES, so a creature that dies in
// response to it is paid for. b11CreaturesDiedThisTurn is the
// per-turn tally's table-wide death count, the way Bloodchief
// Ascension reads the life-lost cell.
//
// Sandbox gap, weaker than printed: the dead are recognised by where
// they sit now, so a creature card that has since left every tracked
// zone, or a permanent that was a creature only through a layer
// effect when it died, earns no Treasure.
func init() {
	Register(Spec{
		OracleID:     "1b3e841e-0f8f-467d-9983-f9b8d081a67f",
		Name:         "Mahadi, Emporium Master",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"A permanent that was only a creature because of another effect isn't counted when it dies."},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Mahadi, Emporium Master — a Treasure for each creature that died this turn", func(g *game.Game, item *game.StackItem) error {
				n := b11CreaturesDiedThisTurn(g)
				return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: n}.Apply(NewContext(g, item))
			}),
		},
	})
}
