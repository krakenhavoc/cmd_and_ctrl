package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hobgoblin Captain — Creature — Goblin Barbarian {1}{R}, 3/1 (Adventures in
// the Forgotten Realms):
//
//	"Pack tactics — Whenever this creature attacks, if you attacked with
//	 creatures with total power 6 or greater this combat, this creature
//	 gains first strike until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d92b3d3-d34b-430f-af6a-db512d204d06",
		Name:         "Hobgoblin Captain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82AfrB09PackTactics("Hobgoblin Captain — pack tactics: first strike until end of turn", func(g *game.Game, item *game.StackItem) error {
				if effects.SourceIsNewObject(g, item) || !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"first strike"},
					Label:    "Hobgoblin Captain — first strike until end of turn",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
