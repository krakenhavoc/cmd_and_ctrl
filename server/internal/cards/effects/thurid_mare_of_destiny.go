package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thurid, Mare of Destiny — Legendary Creature — Pegasus {2}{W}{W}, 2/4
// (Foundations Jumpstart, #5929):
//
//	"Flying, lifelink
//	 Whenever you cast a Pegasus, Unicorn, or Horse creature spell, copy it.
//	 (The copy becomes a token.)
//	 Other Pegasi, Unicorns, and Horses you control get +1/+1."
//
// The cast trigger is Double Vision's body without the new-targets choice (a
// creature spell has no target to change): the copy is made under Thurid's
// controller, from last-known information if the original has left the stack
// already (CR 608.2h), and it is not cast, so it triggers nothing (CR 707.10).
// The anthem is a layer 7c static over the other creatures of the three types
// the controller controls; a changeling counts as each (CR 702.73a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0b96cd7a-2132-4ff2-9666-6e2fbd0a2160",
		Name:            "Thurid, Mare of Destiny",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Pegasus", "Unicorn", "Horse"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast,
				YouCast(And(Creature(), AnySubtype("Pegasus", "Unicorn", "Horse"))),
				"Thurid — copy that spell",
				cmdr2S41CopyTheSpellJustCast),
		},
	})
}
