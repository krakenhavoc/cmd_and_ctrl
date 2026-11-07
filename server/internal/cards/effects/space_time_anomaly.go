package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Space-Time Anomaly — Sorcery {2}{W}{U} (Edge of Eternities, #4461):
//
//	"Target player mills cards equal to your life total."
//
// Glimpse the Unthinkable with the count read as the spell resolves: the
// caster's life total then, not as it was cast. A life total of zero or
// less mills nothing. A target with fewer cards mills the rest and stays
// in the game (CR 701.17b); a player who has left the game is no longer
// a legal target and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "241117a0-8305-48bb-9cb9-9b01e86e08f7",
		Name:         "Space-Time Anomaly",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			victim, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			n := 0
			if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Life > 0 {
				n = p.Life
			}
			if n == 0 {
				return nil
			}
			return MillCards{Player: victim, N: n}.Apply(ctx)
		},
	})
}
