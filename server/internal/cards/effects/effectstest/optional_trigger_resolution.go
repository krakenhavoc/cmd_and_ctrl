package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// optional_trigger_resolution_test.go — #2494, ADR 0197: an optional
// triggered ability goes on the stack like any other and asks its "you
// may" as it resolves (CR 603.5).

const (
	AvengerOfZendikarOracle    = "4ba5b3f6-503b-43e6-b66e-4f8c55cffed7"
	OptionalDrawOnCastProbe    = "test-2494-optional-draw-on-cast"
	optionalDrawOnCastLabel    = "Optional Draw Probe — you may draw a card"
	optionalDrawOnCastQuestion = "Optional Draw Probe — draw a card?"
)

func init() {
	// "Whenever you cast a noncreature spell, you may draw a card": an
	// optional trigger beside the #2884 token probe's footprinted one.
	effects.Register(effects.Spec{
		OracleID: OptionalDrawOnCastProbe,
		Name:     "Optional Draw Probe",
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverYouCast(effects.Noncreature(), optionalDrawOnCastLabel, effects.Do(effects.DrawCards{N: 1})),
				optionalDrawOnCastQuestion),
		},
	})
}
