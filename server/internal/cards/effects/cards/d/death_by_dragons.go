package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Death by Dragons — Sorcery {4}{R}{R} (Commander 2011, #5574):
//
//	"Each player other than target player creates a 5/5 red Dragon creature
//	 token with flying."
//
// The one target is a player. At resolution every player still in the game
// except that one creates the Dragon, in seat order. If the target player is
// gone the spell has no legal target and does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88e912e3-0548-4a4a-94c1-c804093ca1b0",
		Name:         "Death by Dragons",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 {
				return nil
			}
			spared := legal[0].ID
			for _, p := range effects.LivePlayers(ctx.Game) {
				if p.ID == spared {
					continue
				}
				if err := (effects.CreateToken{Controller: p.ID, Template: effects.TokenCard("5/5 red Dragon with flying"), N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
