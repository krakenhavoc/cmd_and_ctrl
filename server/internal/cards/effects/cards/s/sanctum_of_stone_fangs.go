package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanctum of Stone Fangs — Legendary Enchantment — Shrine {1}{B} (Core Set
// 2021, #5511):
//
//	"At the beginning of your first main phase, each opponent loses X life and
//	 you gain X life, where X is the number of Shrines you control."
//
// The first main phase is the precombat main, so it triggers once a turn and
// never in the postcombat main (Black Market Connections). X is counted as the
// trigger resolves, the Sanctum itself included while it is there (CR 608.2h:
// its last-known state once it has gone).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1d8fcb9-26b2-4d09-9db7-e644294749f8",
		Name:         "Sanctum of Stone Fangs",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourPrecombatMain("Sanctum of Stone Fangs — each opponent loses X life and you gain X life, where X is the number of Shrines you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := effects.NeoShrinesYouControl(g, item)
					if x <= 0 {
						return nil
					}
					if err := (effects.EachOpponentLosesLife{N: x}).Apply(ctx); err != nil {
						return err
					}
					return effects.GainLife{Amount: x}.Apply(ctx)
				}),
		},
	})
}
