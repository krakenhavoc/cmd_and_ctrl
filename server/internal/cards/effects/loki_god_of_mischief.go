package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loki, God of Mischief — Legendary Creature — God Sorcerer Villain
// {1}{U}, 2/1 (Marvel Super Heroes, #4455):
//
//	"Whenever a player or permanent becomes the target of an ability you
//	 control, draw a card. This ability triggers only once each turn."
//
// An EventBecomesTarget whose targeter is an activated or triggered
// ability (Event.TargeterKind, #3779) that its controller, you, put on
// the stack: not a spell, and not an opponent's ability. A player or a
// permanent both count. One trigger per batch of targets (an ability
// with two targets draws once), and "only once each turn" is the
// per-turn trigger record (Caretaker's Talent), so a countered draw
// still used the turn's trigger. Loki's own abilities count.
//
// No simplification.
func init() {
	const label = "Loki, God of Mischief — draw a card (one of your abilities targeted)"
	Register(Spec{
		OracleID:     "c112852f-61d5-4b94-b864-c7d0b5333b29",
		Name:         "Loki, God of Mischief",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				if ev.TargeterKind != game.StackItemActivated && ev.TargeterKind != game.StackItemTriggered {
					return false
				}
				return !b11TriggeredThisTurn(g, source.InstanceID, label)
			}, label, Do(DrawCards{N: 1}))),
		},
	})
}
