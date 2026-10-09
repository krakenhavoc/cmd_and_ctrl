package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Academic Ascent — Instant {1}{W} (Reality Fracture):
//
//	"Target creature gets +2/+2 and gains flying until end of turn.
//	 Empower Jace 2."
//
// ADR 0139: a pump spell with the keyword action after it. The Jace is
// empowered whether or not the creature is still there.
//
// No simplification.
func init() {
	const label = "Academic Ascent — +2/+2 and flying until end of turn"
	Register(Spec{
		OracleID:     "406e853b-713f-4c89-9efb-aca74a11183f",
		Name:         "Academic Ascent",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 {
				id := item.Targets[0].ID
				if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: label}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: label}).Apply(ctx); err != nil {
					return err
				}
			}
			return EmpowerJace{N: 2}.Apply(ctx)
		},
	})
}
