package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Binding Mummy — Creature — Zombie {1}{W}, 2/2:
//
//	"Whenever another Zombie you control enters, you may tap target
//	 artifact or creature."
//
// Corpse Knight's condition narrowed to the Zombie type (effective
// subtypes, so a changeling or an embalmed token counts), and Octopus
// Umbra's optional targeted tap: the target is chosen as the trigger goes
// on the stack and the "you may" is asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5adc9d0d-3393-4d91-be85-1f9dd344698b",
		Name:         "Binding Mummy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, true)
					return ok && c.HasSubtype(source.TextSubtype("Zombie"))
				},
				Targets: TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
				Key:     "Binding Mummy — tap target artifact or creature",
				Effect:  b36TapChosenCreature,
			}, "Binding Mummy — tap target artifact or creature?"),
		},
	})
}
