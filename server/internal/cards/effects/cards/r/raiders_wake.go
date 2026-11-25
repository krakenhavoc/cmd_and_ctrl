package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raiders' Wake — Enchantment {3}{B} (Ixalan):
//
//	"Whenever an opponent discards a card, that player loses 2 life.
//	 Raid — At the beginning of your end step, if you attacked this turn,
//	 target opponent discards a card."
//
// Fell Specter's discard trigger (one per card discarded, whatever made the
// opponent discard, and the player who discarded loses the life) and Deadeye
// Tormentor's discard behind Perforating Artist's raid end-step trigger. The
// discard it causes triggers the first ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5fd3060-358c-4a5b-aee9-1afcf6c506ba",
		Name:         "Raiders' Wake",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventDiscardCard},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.B18OpponentDiscarded(ev, source)
				},
				Key: "Raiders' Wake — that player loses 2 life",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Trigger.Event.Actor, -2)
				},
			},
			effects.Targeting(effects.OldB223RaidAtYourEndStep("Raiders' Wake — raid: target opponent discards a card",
				effects.TargetOpponentDiscards(1)),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
