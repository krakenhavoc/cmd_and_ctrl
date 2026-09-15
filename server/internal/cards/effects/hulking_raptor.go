package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hulking Raptor — Creature — Dinosaur {2}{G}{G}, 5/3:
//
//	"Ward {2}"
//	"At the beginning of your first main phase, add {G}{G}."
//
// The first ward card in the catalog, and picked because it is the
// most-played one whose OTHER line the engine can also carry out —
// a five-power four-drop that pays for itself every turn, and a body
// an opponent has to pay {2} extra to answer.
//
// Both halves are as printed. "Your first main phase" is the
// precombat main on an ordinary turn, which is the phase
// EventBeginPrecombatMain announces; the mana lands in the pool at
// the start of the phase and empties at the end of the step like any
// other (CR 106.4), which is exactly the printed behaviour — the
// Raptor ramps the turn it is untapped for, not into your opponent's
// turn.
//
// The trigger uses the stack, like every other triggered ability
// here: the mana arrives when the trigger RESOLVES, so an opponent
// gets a response window first. That is correct and occasionally
// relevant (Stifle), even though nothing in a casual game will use
// it.
func init() {
	Register(Spec{
		OracleID: "9f2e7533-cfc1-4dd8-a9b8-09cdc712ef2f",
		Name:     "Hulking Raptor",
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Hulking Raptor — ward {2}"),
			AtYourPrecombatMain("Hulking Raptor — add {G}{G}", func(g *game.Game, item *game.StackItem) error {
				return g.AddManaForEffect(item.Controller, item.SourceCardID, "{G}{G}")
			}),
		},
	})
}
