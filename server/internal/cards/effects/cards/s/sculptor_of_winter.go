package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sculptor of Winter — Snow Creature — Elf Rogue {1}{G}, 2/2:
//
//	"{T}: Untap target snow land."
//
// Not a mana ability (it targets), so it uses the stack and can be answered.
// The target is a land with the snow supertype, read off its effective type
// line (CR 205.4a); the Sculptor's tap is a cost, so it can't be activated the
// turn it arrives (CR 302.6). A target that has stopped being a snow land by
// resolution is not untapped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "73109353-51fa-43a1-93e0-b74536f372a5",
		Name:         "Sculptor of Winter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Untap target snow land.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target snow land", effects.And(effects.Land(), effects.Snow())),
			Effect:  effects.UntapEachLegalTarget,
		}},
	})
}
