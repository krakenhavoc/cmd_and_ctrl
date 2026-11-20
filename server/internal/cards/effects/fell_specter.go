package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fell Specter — Creature — Specter {3}{B}, 1/3 (Core Set 2019, #5527):
//
//	"Flying
//	When this creature enters, target opponent discards a card.
//	Whenever an opponent discards a card, that player loses 2 life."
//
// Flying, Ravenous Rats's enters trigger (targetOpponentDiscards), and
// Liliana's Caress's discard trigger: the player who discarded loses 2
// life, whatever made them discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d655a1c9-b84d-4069-a2ca-3aa0dacab3e1",
		Name:            "Fell Specter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Fell Specter — target opponent discards a card", targetOpponentDiscards(1)),
				TargetPlayer("target opponent", Opponent())),
			{
				Watches: []game.EventKind{game.EventDiscardCard},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return b18OpponentDiscarded(ev, source)
				},
				Key: "Fell Specter — that player loses 2 life",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Trigger.Event.Actor, -2)
				},
			},
		},
	})
}
