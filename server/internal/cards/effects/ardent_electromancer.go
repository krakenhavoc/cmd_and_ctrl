package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ardent Electromancer — Creature — Human Wizard {2}{R}, 3/2 (Zendikar
// Rising, #5508):
//
//	"When this creature enters, add {R} for each creature in your party.
//	 (Your party consists of up to one each of Cleric, Rogue, Warrior,
//	 and Wizard.)"
//
// The party (CR 700.8) is The Destined cards' count (finB9PartySize: up
// to one creature per type, the best assignment), taken as the trigger
// resolves (CR 700.8a), so the Electromancer itself is the Wizard if
// nothing else fills that slot. The mana is Brightstone Ritual's count
// form (ManaN), and it empties with the pool at the end of the step (CR
// 106.4). The trigger uses the stack: it is not a mana ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43827acb-70cf-44a3-8560-e9f7330c835e",
		Name:         "Ardent Electromancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ardent Electromancer — add {R} for each creature in your party",
				Do(Instruction(func(ctx *Context) error {
					return AddMana{Amount: ManaN(finB9PartySize(ctx.Game, ctx.Controller()), "R")}.Apply(ctx)
				}))),
		},
	})
}
