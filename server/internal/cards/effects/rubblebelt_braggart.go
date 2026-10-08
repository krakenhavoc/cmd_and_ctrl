package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rubblebelt Braggart — Creature — Lizard Warrior {4}{R}:
//
//	"Whenever this creature attacks, if it's not suspected, you may
//	 suspect it. (A suspected creature has menace and can't block.)"
//
// An intervening if (CR 603.4): the trigger is not put on the stack for
// a Braggart that is already suspected, and the condition is checked
// again as it resolves. The "may" is asked as the trigger would go on
// the stack, so the controller is only asked when there is something to
// say yes to. The trigger resolves in the declare-attackers step, before
// blockers, so the menace it grants already applies to this combat's
// block.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ae7f959-9f44-4769-8bd3-49ca18928aba",
		Name:         "Rubblebelt Braggart",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source) && !source.Suspected
			}, "Rubblebelt Braggart — you may suspect it",
				func(g *game.Game, item *game.StackItem) error {
					if g.IsSuspected(item.SourceCardID) { // CR 603.4: re-checked on resolution
						return nil
					}
					return Suspect{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}), "Suspect Rubblebelt Braggart?"),
		},
	})
}
