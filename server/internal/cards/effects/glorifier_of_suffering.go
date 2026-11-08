package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glorifier of Suffering — Creature — Vampire Soldier {2}{W}, 3/2 (The
// Lost Caverns of Ixalan, #4470):
//
//	"When this creature enters, you may sacrifice another creature or
//	 artifact. When you do, put a +1/+1 counter on each of up to two
//	 target creatures."
//
// Dahar Master's shape: the enters trigger asks "you may sacrifice
// another creature or artifact" as it resolves (maySacrificeOne, CR
// 608.2c), and only a permanent that really went raises the CR 603.12
// reflexive trigger, which picks up to two target creatures as it goes on
// the stack (CR 603.3d). Each target still legal as it resolves gets one
// +1/+1 counter (putPlusOneCounterOnEachLegalTarget, Rishkar's). The
// Glorifier itself may be one of the targets.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "07a8ca73-9890-4c87-8b39-fd9eaea5ef45",
		Name:         "Glorifier of Suffering",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Glorifier of Suffering — you may sacrifice another creature or artifact; when you do, +1/+1 counters on up to two target creatures",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Glorifier of Suffering — sacrifice another creature or artifact to put a +1/+1 counter on each of up to two target creatures?",
						Or(Creature(), Artifact()), true, glorifierOfSufferingSacrificed)
				}),
		},
	})
}

// glorifierOfSufferingSacrificed raises the reflexive trigger once a
// permanent has really gone, keyed.
var glorifierOfSufferingSacrificed = RegisterSacrificedOne("prompt-run/glorifier-of-suffering-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return ReflexiveTrigger{
		Label: "Glorifier of Suffering — put a +1/+1 counter on each of up to two target creatures",
		Body:  glorifierOfSufferingCountersBody,
	}.Apply(ctx)
})

// glorifierOfSufferingCountersBody is the reflexive trigger's
// registration: up to two target creatures, a counter on each.
var glorifierOfSufferingCountersBody = game.ReflexiveBody("glorifier-of-suffering/counters",
	simpleBody(putPlusOneCounterOnEachLegalTarget),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("up to two target creatures").WithCount(0, 2)
	}))
