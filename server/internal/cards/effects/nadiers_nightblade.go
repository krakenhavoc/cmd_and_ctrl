package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nadier's Nightblade — Creature — Elf Warrior {2}{B}, 1/3 (EDHREC
// rank 848):
//
//	"Whenever a token you control leaves the battlefield, each
//	 opponent loses 1 life and you gain 1 life."
//
// The token deck's drain — Mirkwood Bats' other half. "Leaves the
// battlefield" is every exit, so a Treasure cracked for mana, a
// Thopter that chump-blocked, and a token bounced or exiled all
// count; EventLTB fires for each of those. The token is read after
// the move (its Controller field survives the move, per
// diedCreature), and it is still findable because tokens only cease
// to exist at the next state check.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "391978f6-0bbc-41e8-9246-f7d0e21c7900",
		Name:         "Nadier's Nightblade",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && IsToken(c) && c.Controller == source.Controller
			}, "Nadier's Nightblade — each opponent loses 1, you gain 1", drainEachOpponent),
		},
	})
}
