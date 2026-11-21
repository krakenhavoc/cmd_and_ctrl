package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "39068bf1-cae4-4ac9-82ef-4e81bf59446e",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "create a 1/1 white Human Noble creature token with vanishing 3 and damage prevention"),
				Do(CreateTokenAdvanced{Spec: Token(gapHumanNobleToken()).WithCounters(game.CounterTime, 3), N: 1})),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "create a 2/2 white Horse creature token with \"Doctors you control have horsemanship.\""),
				Do(CreateToken{Template: gapHorseToken(), N: 1})),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "whenever a creature you control deals combat damage to a player this turn, time travel"),
				Do(DelayedOnEvent{
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
			return combatDamageToPlayerBy(ev, dt.Controller, g)
		})

	// "Time travel." On-disk identity.
	girlInTheFireplaceTimeTravelBody = game.SimpleDelayedBody("girl-in-the-fireplace/time-travel",
		func(g *game.Game, item *game.StackItem) error {
			return TimeTravel{}.Apply(NewContext(g, item))
		})
)
