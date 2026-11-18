package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Markov Retribution — Sorcery {2}{R} (Innistrad: Crimson Vow, #5495):
//
//	"Choose one or both —
//	 • Creatures you control get +1/+0 until end of turn.
//	 • Target Vampire you control deals damage equal to its power to
//	   another target creature."
//
// The bullets resolve in printed order (CR 608.2c), so with both chosen
// the Vampire is already pumped when it bites. The pump covers the
// creatures you control as the spell resolves (CR 611.2c). The bite has two
// target clauses of its own, the second Distinct from the first
// ("another"); it deals nothing unless both are still legal as the spell
// resolves (CR 608.2b), and the damage is the Vampire's power as it is
// then (CR 608.2h), dealt by the Vampire itself. The Vampire is read
// through the layers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a9ca9b2-0085-41de-9c08-60aaa8122da2",
		Name:         "Markov Retribution",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeDoing("Creatures you control get +1/+0 until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Label: "Markov Retribution — +1/+0"}.Apply(ctx)
				}),
			ModeDoing("Target Vampire you control deals damage equal to its power to another target creature.",
				Clauses(
					TargetCreature("target Vampire you control", YouControl(), Subtype("Vampire")),
					Distinct(TargetCreature("another target creature")),
				),
				vowB07VampireBitesAnother),
		),
	})
}
