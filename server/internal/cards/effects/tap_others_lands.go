package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// tap_others_lands.go — Piracy (#6204): "Until end of turn, you may tap
// lands you don't control for mana. Spend this mana only to cast
// spells." The engine half is game/tap_others_lands.go: the permission,
// the activation gate, the auto-tapper's sources and the cast-only
// restriction on the mana.

// TapLandsYouDontControlThisTurn is the whole of Piracy's text.
//
//	Effect: TapLandsYouDontControlThisTurn("Piracy"),
func TapLandsYouDontControlThisTurn(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		g.TapLandsYouDontControlThisTurnForEffect(ctx.Source(), ctx.Controller(),
			name+" — until end of turn, you may tap lands you don't control for mana (spend it only to cast spells)")
		return nil
	}
}
