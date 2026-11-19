package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ecstatic Electromancer — Creature — Goblin Wizard {1}{U/R}{U/R}, 2/3
// (Ravnica: Clue Edition):
//
//	"Prowess
//	 Whenever this creature enters or attacks, you may discard a card. If
//	 you do, draw a card."
//
// One printed ability with two trigger conditions, so one trigger watching
// two event kinds (Flaring Cinder's shape): its own entry or its own attack
// declaration. The rummage is Flaring Cinder's.
//
// No simplification.
func init() {
	const label = "Ecstatic Electromancer — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:        "2ad9f28c-94a9-4884-9e3b-42341adfaa5d",
		Name:            "Ecstatic Electromancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventAttack}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			}, label, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, label, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
	})
}
