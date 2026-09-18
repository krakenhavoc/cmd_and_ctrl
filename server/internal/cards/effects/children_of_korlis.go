package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Children of Korlis — 1/1 Creature — Human Rebel Cleric for {W}
// (EDHREC rank 4400):
//
//	"Sacrifice this creature: You gain life equal to the life
//	 you've lost this turn. (Damage causes loss of life.)"
//
// A one-mana blocker that undoes a turn. In Commander it is the
// answer to a storm kill that pays its own life and to a single
// enormous hit — anything that took you low in ONE turn can be given
// straight back at instant speed, for free, with the Children already
// on the battlefield. Roadmap batch 42 (#449), "no new machinery".
//
// "The life you've lost this turn" is a per-turn tally the engine
// keeps per player, and the reminder text is the load-bearing part:
// DAMAGE counts. Ten damage from a commander and two life paid to a
// fetchland are both life lost, and both come back. Life you gained
// and then lost counts too; the tally is losses, not net.
//
// The tally is read when the ABILITY RESOLVES, not when it is
// activated. The creature is already gone by then — sacrificing is a
// cost (CR 601.2h) — so a response that drains you between activation
// and resolution is money in the bank.
//
// No cost but the sacrifice: no tap, no mana, so it works the turn it
// arrives and at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d44a8f89-b9c0-4e8a-9179-baa0c448a3c9",
		Name:         "Children of Korlis",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "Sacrifice this creature: You gain life equal to the life you've lost this turn.",
			Cost:  SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{
					Player: item.Controller,
					Amount: b18LifeLostThisTurn(g, item.Controller),
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
