package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tethermage's Advantage — Instant {G} (Reality Fracture, tracker #2795):
//
//	"Target creature gets +2/+2 and gains reach until end of turn.
//	 Untap it."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b4f89885-eeb9-46c7-9820-9758b624ca24",
		Name:         "Tethermage's Advantage",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Tethermage's Advantage — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"reach"}, Label: "Tethermage's Advantage — reach"}).Apply(ctx); err != nil {
					return err
				}
				return UntapTarget{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
