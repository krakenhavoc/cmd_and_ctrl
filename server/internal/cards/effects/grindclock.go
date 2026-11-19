package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grindclock — Artifact {2} (Scars of Mirrodin, #5577):
//
//	"{T}: Put a charge counter on this artifact.
//	 {T}: Target player mills X cards, where X is the number of charge
//	 counters on this artifact."
//
// X is counted as the second row resolves, off the Grindclock as it is
// then, or as it last existed on the battlefield if it has left
// (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6b7de1c0-4b97-47e1-8ea7-ba55a8d16a9b",
		Name:         "Grindclock",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Put a charge counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    TapCost(),
				Effect:  b33PutChargeCounterOnSelf,
			},
			{
				Label:   "{T}: Target player mills X cards, where X is the number of charge counters on this artifact.",
				Cost:    TapCost(),
				Targets: TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					info, ok := NewContext(g, item).SourcePermanent()
					if !ok || info.Counters[game.CounterCharge] <= 0 {
						return nil
					}
					return TargetPlayerMills(info.Counters[game.CounterCharge])(g, item)
				},
			},
		},
	})
}
