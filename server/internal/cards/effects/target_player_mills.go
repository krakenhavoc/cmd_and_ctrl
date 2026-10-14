package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// target_player_mills.go — "target player mills N cards" as an
// ability's whole body (Codex Shredder, Embalmer's Tools). Append-only.

// TargetPlayerMills is the Effect for "Target player mills N cards":
// the first still-legal player target mills, and a target that left
// the game in response leaves nothing to do (CR 608.2b).
func TargetPlayerMills(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetPlayer {
				return MillCards{Player: t.ID, N: n}.Apply(ctx)
			}
		}
		return nil
	}
}
