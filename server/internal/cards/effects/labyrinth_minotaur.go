package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Labyrinth Minotaur — Creature — Minotaur {3}{U}, 1/4 (Homelands, #5753):
//
//	"Whenever this creature blocks a creature, that creature doesn't
//	 untap during its controller's next untap step."
//
// Wall of Frost's text. It triggers once for each attacker it blocks (CR
// 509.3b): EventBlock names the Minotaur in CardID and the attacker in
// Target, one event per pair of the final declaration, and none for a
// creature put onto the battlefield blocking (CR 509.4). The mark is ADR
// 0058's controller-keyed one-shot (UntapSkip.Player nil), which follows
// the creature: the ruling is that the ability "doesn't track the
// creature's controller. If the affected creature changes controllers
// before its old controller's next untap step, this ability will prevent
// it from being untapped during its new controller's next untap step."
// The other ruling, "If the creature Labyrinth Minotaur blocked is
// untapped at the time its controller's next untap step begins, this
// ability has no effect", is the mark being used up by that step either
// way.
//
// "That creature" is the attacker as it was when the Minotaur blocked it.
// The trigger stamps its object (Params.Object, CR 400.7), so an attacker
// that left the battlefield and came back before the trigger resolved is
// a new object and is not marked.
//
// No simplification.
func init() {
	const label = "Labyrinth Minotaur — the creature it blocks doesn't untap during its controller's next untap step"
	t := On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return b28SelfBlocked(ev, source)
	}, label, func(g *game.Game, item *game.StackItem) error {
		attacker := item.Params.Object
		if attacker.ID == uuid.Nil || !onBattlefield(g, attacker.ID) || g.ObjectEpochForEffect(attacker.ID) != attacker.Epoch {
			return nil
		}
		return DoesntUntapNextUntapStep{Targets: []uuid.UUID{attacker.ID}, Label: label}.Apply(NewContext(g, item))
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Object = game.ObjectRef{ID: ev.Target, Epoch: g.ObjectEpochForEffect(ev.Target)}
		return item
	}
	Register(Spec{
		OracleID:     "1e7b075b-1c0f-4c44-a597-5c167a786cb1",
		Name:         "Labyrinth Minotaur",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
