package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skophos Maze-Warden — Creature — Minotaur Warrior {3}{R}, 3/4 (Theros Beyond
// Death, #5514):
//
//	"{1}: This creature gets +1/-1 until end of turn.
//	 Whenever another creature becomes the target of an ability of a land you control
//	 named Labyrinth of Skophos, you may have this creature fight that creature."
//
// The trigger reads EventBecomesTarget (CR 115.3, announced as the ability goes on
// the stack): the target is a creature other than this one, the targeter is an
// ACTIVATED ability (the event's TargeterKind), and its source is a land named
// Labyrinth of Skophos controlled by this creature's controller. "That creature"
// is the event's target, fought only if it and this creature are both still on the
// battlefield as the trigger resolves (CR 701.14b), after the "you may" is asked
// as it resolves (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d9b7fca-9514-4aea-be51-b6c8debb3de4",
		Name:         "Skophos Maze-Warden",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.SelfPumpRow("Skophos Maze-Warden", "{1}", 1, -1)},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventBecomesTarget, skophosMazeWardenCondition,
				"Skophos Maze-Warden — you may have it fight that creature", skophosMazeWardenFight),
				"Skophos Maze-Warden — have it fight that creature?"),
		},
	})
}

func skophosMazeWardenCondition(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || ev.CardID == source.InstanceID || ev.TargeterKind != game.StackItemActivated {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !target.IsCreature() {
		return false
	}
	land, ok := g.LookupCardForEffect(ev.Source)
	return ok && land.IsLand() && land.Controller == source.Controller && land.Name == "Labyrinth of Skophos"
}

func skophosMazeWardenFight(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return effects.B10Fight(ctx, item.SourceCardID, item.Trigger.Event.CardID)
}
