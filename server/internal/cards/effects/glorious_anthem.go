package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glorious Anthem — "Creatures you control get +1/+1."
//
// First S16 catalog card. One static ability at Layer 7c (modify
// power/toughness). Predicate: target is a creature controlled by
// the same player as Anthem itself; self is excluded automatically
// because the Anthem isn't a creature (predicate filters on
// IsCreature() before checking controller).
//
// What the layer engine does with this:
//   - On every recompute pass, walks every battlefield card and
//     calls AppliesTo(target, g, source=Anthem). When true, calls
//     Apply(target.effective, ...) which bumps power and toughness.
//   - Multiple Anthems stack: each contributes its own +1/+1 in
//     timestamp order at Layer 7c.
//   - Removing the Anthem from the battlefield invalidates the
//     layer version (layer_listener.go); the next snapshot
//     re-resolves and the +1/+1 vanishes.
//
// Out of scope for this card / S16:
//   - Anthems are creature-static-abilities-aware in real MTG (they
//     interact with copy effects, type changes, etc.); the layer
//     engine handles those interactions automatically once those
//     effects ship.
func init() {
	Register(Spec{
		OracleID: "e3886fe8-9b76-4613-8891-4ec74657c087",
		Name:     "Glorious Anthem",
		Static: []game.StaticAbility{
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller
				},
				Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
		},
	})
}
