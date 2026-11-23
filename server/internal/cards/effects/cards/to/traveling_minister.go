package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Traveling Minister — Creature — Human Cleric {W}, 1/1 (Innistrad:
// Crimson Vow):
//
//	"{T}: Target creature gets +1/+0 until end of turn. You gain 1 life.
//	 Activate only as a sorcery."
//
// The pump needs a still-legal target; the life is gained either way, as
// a target that became illegal only skips the part that needs it
// (CR 608.2b). Sorcery timing is the activation gate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d789ccd-8192-41b1-a5bb-1e2174657698",
		Name:         "Traveling Minister",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: Target creature gets +1/+0 until end of turn. You gain 1 life. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)),
			Cost:         effects.TapCost(),
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := effects.BoostTargetUntilEOT("Traveling Minister — +1/+0", 1, 0)(g, item); err != nil {
					return err
				}
				return effects.GainLife{Amount: 1}.Apply(ctx)
			},
		}},
	})
}
