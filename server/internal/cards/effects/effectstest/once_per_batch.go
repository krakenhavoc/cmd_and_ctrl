package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// once_per_batch_test.go — #587: "whenever one or more …" is an engine
// flag. The engine emits one event per creature; an OncePerBatch
// ability fires for the first event of an event batch and declines
// every later event of that same batch.
//
// #829 replaced the "while its item is pending, on the stack, or
// waiting on its prompt" reading of "the batch" with the real batch
// identity on Event.Batch — see the tests at the bottom of this file
// and server/internal/game/event_batch.go.

const (
	BatchProbeOracle      = "test-once-per-batch-probe"
	BatchMayProbeOracle   = "test-once-per-batch-optional-probe"
	batchProbeAttackLabel = "Batch Probe — you gain 1 life"
)

func init() {
	effects.Register(effects.Spec{
		OracleID: BatchProbeOracle,
		Name:     "Batch Probe",
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, batchProbeAttackLabel, effects.Do(effects.GainLife{Amount: 1}))),
			// A second, unrelated ability on the same source: its own
			// key, so the first one's item must not suppress it.
			effects.WheneverYouDraw("Batch Probe — you gain 5 life", effects.Do(effects.GainLife{Amount: 5})),
		},
	})
	effects.Register(effects.Spec{
		OracleID: BatchMayProbeOracle,
		Name:     "Batch May Probe",
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.Optional(effects.WheneverYouDraw("Batch May Probe — you may gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
				"Batch May Probe — gain 1 life?")),
		},
	})
}

// --- #829: a LATER batch is a second trigger -----------------------
//
// "One or more" collapses one batch, not everything the ability has
// in flight. Dour Port-Mage is the reported card: two bounces at two
// different times are two occurrences (CR 603.2c), even when the
// first one's draw is still sitting on the stack.

const (
	B829UnsummonOracle   = "837182db-1bf3-4a2c-bd01-1af9d9873561"
	B829EvacuationOracle = "fdd94383-b573-439a-8e1c-925af887c5a6"
	B829PortMageOracle   = "cf58e309-00e8-438e-813e-2e1c1002db23"
)
