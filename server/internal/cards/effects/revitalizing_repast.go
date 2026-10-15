package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8dd6d060-d023-48a6-85cb-7a5521b6257b",
		Name:         "Revitalizing Repast",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (AddCounter{Target: id, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   id,
				Keywords: []string{"indestructible"},
				Label:    "Revitalizing Repast — indestructible",
			}.Apply(ctx)
		},
	})
}
