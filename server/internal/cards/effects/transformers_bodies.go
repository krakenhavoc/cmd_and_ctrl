package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// transformers_bodies.go — the registered bodies the More Than Meets the
// Eye cards (ADR 0152, #3195) need. A body's key is an on-disk identity
// (testdata/effect_keys.txt), so these live in a file of their own.

// transformersConvertThisBody is the delayed trigger's body for "convert
// <this> at end of combat" (Ultra Magnus). The trigger records the
// source's turn-over count when it is created (CR 701.28e), which
// ConvertThis reads, so a Magnus that turned over in between ignores it.
var transformersConvertThisBody = game.SimpleDelayedBody("transformers/convert-this", func(g *game.Game, item *game.StackItem) error {
	return ConvertThis{}.Apply(NewContext(g, item))
})
