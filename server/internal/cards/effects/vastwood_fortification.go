package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vastwood Fortification // Vastwood Thicket — Modal DFC. Front: Instant
// {G}: "Put a +1/+1 counter on target creature." Back: Land
// (mdfc_lands.go).
//
// One counter through the counter pipeline, so Hardened Scales and
// Doubling Season see it. A target that is no longer legal means the spell
// does nothing. No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce148a0c-6c63-49d5-a156-99efae4e367a",
		Name:         "Vastwood Fortification",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
		},
	})
}
