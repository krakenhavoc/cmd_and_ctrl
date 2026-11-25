package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// colorless_damage_sources.go — #6230: "Black and/or red permanents and
// spells are colorless sources of damage." (Ghostly Flame). The engine
// half is game/damage_source_colorless.go.
//
//	Static: []game.StaticAbility{ColorlessDamageSources("BR", "Black and/or red permanents and spells are colorless sources of damage.")}

// ColorlessDamageSources is a world rule the permanent carries on
// itself: while it is on the battlefield, every permanent or spell with
// any of `colors` (WUBRG letters) is a colourless source of damage, so
// protection from those colours, and every prevention shield or source
// choice that names them, doesn't see its damage as coloured (CR 105.2c,
// CR 702.16e). The objects keep their colours for everything else.
func ColorlessDamageSources(colors, label string) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer6Ability,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.ColorlessDamageSourceColors += colors
		},
	}
}
