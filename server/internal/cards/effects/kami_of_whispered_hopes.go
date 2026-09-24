package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kami of Whispered Hopes — Creature — Spirit {2}{G}, 1/1:
//
//	"If one or more +1/+1 counters would be put on a permanent you
//	 control, that many plus one +1/+1 counters are put on that
//	 permanent instead.
//	 {T}: Add X mana of any one color, where X is this creature's
//	 power."
//
// The replacement is Hardened Scales' shape (+1 to CounterDelta) over
// a wider zone: PERMANENT you control, not just a creature — a
// noncreature artifact with a +1/+1 counter (rare, but real) gets the
// bonus too, which is why this doesn't reuse
// plusOneCounterPlacementOnYourCreature. The mana ability is Mona
// Lisa's shape: X read at activation from Kami's own current power,
// so a +1/+1 counter this same replacement doubled up feeds the very
// ability it grew.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f122624f-f30d-444e-a62a-939829241045",
		Name:         "Kami of Whispered Hopes",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			{
				Watches:   []game.EventKind{game.EventCounterPlaced},
				AppliesTo: kamiPlusOneCounterOnYourPermanent,
				Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
					ev.CounterDelta += 1
					return nil
				},
				Controller: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Kami of Whispered Hopes: +1 +1/+1 counter",
			},
		},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedOneColor(SourcePower),
			Label:        "Add X mana of any one color, where X is this creature's power",
		}},
	})
}

// kamiPlusOneCounterOnYourPermanent is Kami of Whispered Hopes' own
// AppliesTo: the Hardened Scales predicate (CR 122.6 / CR 614.1 —
// placement only) widened from "a creature you control" to "a
// permanent you control".
func kamiPlusOneCounterOnYourPermanent(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if ev.Kind != game.RepEventCounter {
		return false
	}
	if ev.CounterDelta <= 0 {
		return false
	}
	if ev.CounterName != "+1/+1" {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.CounterTarget)
	if !ok {
		return false
	}
	return target.Controller == src.Controller
}
