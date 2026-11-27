package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mogg Infestation — Sorcery {3}{R}{R} (Stronghold, S82 OLD b751, #5627):
//
//	"Destroy all creatures target player controls. For each creature that died
//	 this way, that player creates two 1/1 red Goblin creature tokens."
//
// Bane of Progress' sweep with a count: DestroyAllMatching over the target
// player's creatures as the spell resolves, and a keyed tail told how many were
// actually destroyed (CR 701.8a: one a replacement exiled, or kept by
// regeneration, is not counted). That player, not the caster, creates the tokens,
// two for each. A target player who has left the game makes the spell do nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cfd61b53-83c6-4886-8503-ddcb92dc7f85",
		Name:         "Mogg Infestation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := effects.TargetedPlayerOf(ctx)
			if player == uuid.Nil {
				return nil
			}
			return effects.DestroyAllMatching{
				Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCreature() && c.Controller == player },
				ThenTail: ctx.ItemTail(moggInfestationGoblins),
			}.Apply(ctx)
		},
	})
}

// moggInfestationGoblins is "that player creates two 1/1 red Goblin creature
// tokens" for each creature destroyed. Keyed (#4253 slice 10a).
var moggInfestationGoblins = effects.MassTail("mass/mogg-infestation-goblins", func(ctx *effects.Context, _ []game.Card, destroyed int) error {
	player := effects.TargetedPlayerOf(ctx)
	if destroyed <= 0 || player == uuid.Nil {
		return nil
	}
	return effects.CreateToken{Controller: player, Template: effects.TokenCard("1/1 red Goblin"), N: 2 * destroyed}.Apply(ctx)
})
