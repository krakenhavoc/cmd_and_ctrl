package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Showdown of the Skalds — Enchantment — Saga {2}{R}{W} (Kaldheim, #5505):
//
//	"I — Exile the top four cards of your library. Until the end of your
//	 next turn, you may play those cards.
//	 II, III — Whenever you cast a spell this turn, put a +1/+1 counter on
//	 target creature you control."
//
// Chapter I is the impulse exile (Burning Curiosity's), four cards, playable
// until the end of the controller's next turn. Chapters II and III are
// Rediscover the Way's repeating delayed trigger (CR 603.7b) over any spell
// its controller casts, with the target chosen as each instance goes on the
// stack (CR 603.3d) and rechecked as it resolves.
//
// No simplification.
func init() {
	const name = "Showdown of the Skalds"
	schedule := func(g *game.Game, item *game.StackItem) error {
		return effects.DelayedOnEvent{
			Label:     name + " — put a +1/+1 counter on target creature you control",
			On:        []game.EventKind{game.EventCast},
			Condition: effects.YouNextCastCondition,
			Body:      showdownOfTheSkaldsBody,
			Repeats:   true,
		}.Apply(effects.NewContext(g, item))
	}
	effects.Register(effects.Spec{
		OracleID:     "6107f9aa-3373-4166-b4b8-a26fd6d69c72",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "exile the top four cards of your library; until the end of your next turn, you may play those cards"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTopNUntilYourNextTurn(effects.NewContext(g, item), 4)
				}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "whenever you cast a spell this turn, put a +1/+1 counter on target creature you control"), schedule),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "whenever you cast a spell this turn, put a +1/+1 counter on target creature you control"), schedule),
		},
	})
}

// showdownOfTheSkaldsBody is chapters II and III's delayed trigger; its
// target clause is announced as each instance goes on the stack. An on-disk
// key: never renamed, never reused.
var showdownOfTheSkaldsBody = game.TargetedDelayedBody("showdown-of-the-skalds/counter",
	effects.SimpleBody(effects.B36CounterOnChosenAnimal),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature you control", effects.YouControl())
	}))
