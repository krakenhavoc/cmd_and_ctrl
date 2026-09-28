package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// multi_block.go is the card half of #1706: a creature that can block
// more than one attacker (CR 509.1a/b). The engine half is
// game/multi_block.go; see ADR 0045's 2026-09-28 amendment, Decisions
// 60-63.
//
// Two builders, each a layer-6 static over whatever `applies` selects:
//
//	CanBlockAdditional(selfOnly, 1)              // Two-Headed Giant of Foriys
//	CanBlockAdditional(b16CreaturesYouControl, 1) // High Ground, Brave the Sands
//	CanBlockAdditional(AttachedToSource, 1)      // Echo Circlet
//	CanBlockAnyNumber(selfOnly)                  // Palace Guard
//
// They write Characteristic.AdditionalBlocks / BlocksAnyNumber, which
// only ever grow: several "an additional creature" effects add up, and
// "any number" beats any count. A creature's own line goes with its
// abilities (the static is the creature's, and is not applied once it
// has lost them); High Ground's does not, because it is High Ground's.

// CanBlockAdditional is "can block an additional N creatures each
// combat" (N = 1 for "an additional creature") on every permanent
// `applies` selects.
func CanBlockAdditional(applies func(target *game.Card, g *game.Game, source *game.Card) bool, n int) game.StaticAbility {
	if n <= 0 {
		panic("effects: CanBlockAdditional needs a positive count")
	}
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: applies,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.AdditionalBlocks += n
		},
	}
}

// CanBlockAnyNumber is "can block any number of creatures" on every
// permanent `applies` selects.
func CanBlockAnyNumber(applies func(target *game.Card, g *game.Game, source *game.Card) bool) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: applies,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.BlocksAnyNumber = true
		},
	}
}
