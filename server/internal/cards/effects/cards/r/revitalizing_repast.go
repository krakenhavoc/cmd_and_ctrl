package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Revitalizing Repast // Old-Growth Grove — Modal double-faced card
// (EDHREC top 1,000, #3654). The front face:
//
//	"Put a +1/+1 counter on target creature. It gains indestructible
//	 until end of turn." {B/G} Instant
//
// A counter and a turn-long keyword grant on the one target, in printed
// order. A target that became illegal in response fizzles the whole spell
// (CR 608.2b). The land back face is the shared mdfc_lands.go cycle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8dd6d060-d023-48a6-85cb-7a5521b6257b",
		Name:         "Revitalizing Repast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			// The keyword waits for a prompt the counter asks (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.AddCounter{Target: id, Kind: "+1/+1", N: 1}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"indestructible"},
					Label:    "Revitalizing Repast — indestructible",
				}.Apply(ctx)
			}),
		),
	})
}
