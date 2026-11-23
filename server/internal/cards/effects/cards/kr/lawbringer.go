package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lawbringer — Creature — Kor Rebel {2}{W}, 2/2 (Nemesis, #5618):
//
//	"{T}, Sacrifice this creature: Exile target red creature."
//
// Hanged Executioner's exile (the shared exileFirstLegalCardTarget body)
// behind a tap-and-sacrifice cost. The tap is a {T} cost, so a Lawbringer
// that came under your control this turn can't pay it (CR 302.6). The
// target is judged again as it resolves (CR 608.2b): a creature that
// stopped being red in response is not exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15cc92c0-76ce-4b97-b1a2-a7d58bbfe857",
		Name:         "Lawbringer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Exile target red creature.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target red creature", effects.OfColor("R")),
			Effect:  effects.ExileFirstLegalCardTarget,
		}},
	})
}
