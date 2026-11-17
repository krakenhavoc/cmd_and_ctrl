package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Auriok Bladewarden — Creature — Human Soldier {1}{W}, 1/1 (Mirrodin,
// #5606):
//
//	"{T}: Target creature gets +X/+X until end of turn, where X is this
//	 creature's power."
//
// Hemosymbic Mite's pump on an activated ability. X is the Bladewarden's
// power as the ability resolves, or its last known power if it has left
// the battlefield (sourcePowerNow, CR 608.2h), and it is locked in then:
// the bonus does not change if the Bladewarden's power changes later in
// the turn. It may target itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2884e332-707c-4a97-adc1-e1317a513ff4",
		Name:         "Auriok Bladewarden",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature gets +X/+X until end of turn, where X is this creature's power.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0))},
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect: Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				x := sourcePowerNow(ctx)
				return BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Auriok Bladewarden — +X/+X"}.Apply(ctx)
			})),
		}},
	})
}
