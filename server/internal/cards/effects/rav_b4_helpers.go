package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// rav_b4_helpers.go — shared bodies for Ravnica: City of Guilds, batch 4
// (#5600). Its own file, so a concurrent batch never collides here.
//
// Append-only: add a builder, never change what one means.

// ravB4TargetOpponentCreates is the Hunted cycle's enters trigger: "When
// this creature enters, target opponent creates <n> <token>." (Hunted
// Dragon, Hunted Lammasu, Hunted Phantasm, Hunted Troll). It is Hunted
// Horror's shape: the opponent creates the tokens, so they own and
// control them (CR 111.2), and an opponent who has left by the time the
// trigger resolves gets nothing (CR 608.2b). `token` is a tokens_table.go
// key.
func RavB4TargetOpponentCreates(label, token string, n int) game.TriggeredAbility {
	return Targeting(
		WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					return CreateToken{Controller: t.ID, Template: TokenCard(token), N: n}.Apply(ctx)
				}
			}
			return nil
		}),
		TargetPlayer("target opponent", Opponent()))
}
