package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hidden Herbalists — Creature — Human Druid {1}{G}, 2/2 (Aether Revolt,
// #5536):
//
//	"Revolt — When this creature enters, if a permanent left the
//	 battlefield under your control this turn, add {G}{G}."
//
// Airdrop Aeronauts' intervening-if enters trigger (CR 603.4) with a
// mana body: the trigger is not a mana ability, so it uses the stack and
// the {G}{G} empties from the pool at the end of the step as usual.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3289539e-51eb-4915-bd77-3ac362136f2a",
		Name:         "Hidden Herbalists",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersIfRevolt("Hidden Herbalists — add {G}{G}", Do(AddMana{Produced: "{G}{G}"})),
		},
	})
}
