package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Girl in the Fireplace — Enchantment — Saga {2}{W} (Doctor Who):
//
//	"I — Create a 1/1 white Human Noble creature token with vanishing 3 and
//	 'Prevent all damage that would be dealt to this token.'
//	 II — Create a 2/2 white Horse creature token with 'Doctors you control
//	 have horsemanship.'
//	 III — Whenever a creature you control deals combat damage to a player
//	 this turn, time travel."
//
// The Human Noble is a token template whose printed abilities are vanishing's
// two triggers and a damage prevention (token_catalog.go Replacements); it is
// made with three time counters (vanishing 3, CR 702.63a). The Horse's
// ability is a layer-6 horsemanship grant to the Doctors its controller
// controls. III is a delayed trigger for the rest of the turn that fires once
// per creature that deals combat damage to a player (CR 603.7).
//
// No simplification.
func init() {
	const name = "The Girl in the Fireplace"
	effects.Register(effects.Spec{
		OracleID:     "39068bf1-cae4-4ac9-82ef-4e81bf59446e",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create a 1/1 white Human Noble creature token with vanishing 3 and damage prevention"),
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.GapHumanNobleToken()).WithCounters(game.CounterTime, 3), N: 1})),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "create a 2/2 white Horse creature token with \"Doctors you control have horsemanship.\""),
				effects.Do(effects.CreateToken{Template: effects.GapHorseToken(), N: 1})),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "whenever a creature you control deals combat damage to a player this turn, time travel"),
				effects.Do(effects.DelayedOnEvent{
					Label:     "The Girl in the Fireplace — time travel (a creature you control dealt combat damage to a player)",
					On:        []game.EventKind{game.EventDealDamage},
					Condition: girlInTheFireplaceHitsCondition,
					Body:      girlInTheFireplaceTimeTravelBody,
					Repeats:   true,
				})),
		},
	})
}

var (
	// "Whenever a creature you control deals combat damage to a player."
	// On-disk identity.
	girlInTheFireplaceHitsCondition = game.DelayedCondition("girl-in-the-fireplace/creature-combat-damage-to-player",
		func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
			return effects.CombatDamageToPlayerBy(ev, dt.Controller, g)
		})

	// "Time travel." On-disk identity.
	girlInTheFireplaceTimeTravelBody = game.SimpleDelayedBody("girl-in-the-fireplace/time-travel",
		func(g *game.Game, item *game.StackItem) error {
			return effects.TimeTravel{}.Apply(effects.NewContext(g, item))
		})
)
