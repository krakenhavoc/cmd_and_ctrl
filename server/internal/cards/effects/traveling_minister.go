package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1d789ccd-8192-41b1-a5bb-1e2174657698",
		Name:         "Traveling Minister",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{T}: Target creature gets +1/+0 until end of turn. You gain 1 life. Activate only as a sorcery.",
			Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0)),
			Cost:         TapCost(),
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := boostTargetUntilEOT("Traveling Minister — +1/+0", 1, 0)(g, item); err != nil {
					return err
				}
				return GainLife{Amount: 1}.Apply(ctx)
			},
		}},
	})
}
