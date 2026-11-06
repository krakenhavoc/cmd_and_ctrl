package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sample Collector — Creature — Troll Detective {2}{G}, 2/3 (Murders at
// Karlov Manor, #5193):
//
//	"Whenever this creature attacks, you may collect evidence 3. When you
//	 do, put a +1/+1 counter on target creature you control. (To collect
//	 evidence 3, exile cards with total mana value 3 or greater from your
//	 graveyard.)"
//
// Collect evidence is the effect form (RegisterCollectEvidence), asked as
// the attack trigger resolves; a graveyard holding less than 3 in all asks
// nothing (CR 701.59b). "When you do" is a reflexive trigger (CR 603.12),
// Ant-Man's shape: its target is chosen when it goes on the stack, after
// the cards are exiled, and it may be this creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "566b65bb-fa5d-4da6-a766-b08128ee0049",
		Name:         "Sample Collector",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Sample Collector — you may collect evidence 3; when you do, put a +1/+1 counter on target creature you control",
				func(g *game.Game, item *game.StackItem) error {
					return sampleCollectorEvidence.Ask(NewContext(g, item), "Sample Collector — collect evidence 3 (exile cards with total mana value 3 or more from your graveyard) to put a +1/+1 counter on a creature you control? Choose none to decline.")
				}),
		},
	})
}

// sampleCollectorEvidence is "you may collect evidence 3. When you do, …".
// An on-disk key: never renamed, never reused.
var sampleCollectorEvidence = RegisterCollectEvidence("choose-cards/sample-collector-collect-evidence", 3,
	func(ctx *Context, _ []uuid.UUID) error {
		return ReflexiveTrigger{
			Label: "Sample Collector — put a +1/+1 counter on target creature you control",
			Body:  sampleCollectorCounterBody,
		}.Apply(ctx)
	})

// sampleCollectorCounterBody is "put a +1/+1 counter on target creature you
// control". An on-disk key.
var sampleCollectorCounterBody = game.ReflexiveBody("sample-collector/counter", simpleBody(PlusOneCounterOnTheTarget),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature you control", YouControl()) }))
