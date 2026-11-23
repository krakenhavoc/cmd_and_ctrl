package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exterminatus — Sorcery {5}{W}{B} (Warhammer 40,000 Commander, #5948):
//
//	"Nonland permanents your opponents control lose indestructible until
//	 end of turn.
//	 Destroy all nonland permanents."
//
// Spectacular Pileup's shape: the loss of indestructible is a layer-6
// removal over the opponents' nonland permanents as the spell resolves
// (CR 613.1f), then every nonland permanent is destroyed. Your own
// indestructible permanents survive; regeneration is not barred.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d90ffb5-63da-48d5-a5a2-7b61062556f6",
		Name:         "Exterminatus",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Nonland(), effects.OpponentControls())) {
				if err := effects.G2LosesIndestructibleUntilEOT(ctx, c.InstanceID); err != nil {
					return err
				}
			}
			return effects.DestroyAllMatching{Match: effects.Nonland()}.Apply(ctx)
		},
	})
}
