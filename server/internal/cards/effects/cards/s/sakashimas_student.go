package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sakashima's Student — Creature — Human Ninja {2}{U}{U}, 0/0 (Secret Lair
// Drop):
//
//	"Ninjutsu {1}{U}
//	 You may have this creature enter as a copy of any creature on the
//	 battlefield, except it's a Ninja in addition to its other creature types."
//
// Ninjutsu is the shared ninjutsu ability. The clone clause is the ordinary
// enter-as-a-copy replacement over any creature on the battlefield, with the
// "in addition" subtype added to the copied values (CR 707.9b). Entering by
// ninjutsu or from the stack, it is the same replacement. Declined, or with no
// creature to copy, it is a 0/0 and dies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "afcc68e2-3634-43ce-ae99-dd2ec735d146",
		Name:         "Sakashima's Student",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Ninjutsu("{1}{U}")},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("Sakashima's Student", effects.AnyCreatureOnBattlefield, effects.S82OldB022AlsoANinja),
		},
	})
}
