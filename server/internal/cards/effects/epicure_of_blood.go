package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Epicure of Blood — Creature — Vampire {4}{B}, 4/4 (Foundations Jumpstart):
//
//	"Whenever you gain life, each opponent loses 1 life."
//
// Cliffhaven Vampire's trigger: one per life-gain event (lifelink included),
// and the drain is life loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bdb319a6-a1cd-44f5-999d-af1bb55b2fc4",
		Name:         "Epicure of Blood",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventChangeLife, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Target == source.Controller && ev.Amount > 0
			}, "Epicure of Blood — each opponent loses 1 life", Do(EachOpponentLosesLife{N: 1})),
		},
	})
}
