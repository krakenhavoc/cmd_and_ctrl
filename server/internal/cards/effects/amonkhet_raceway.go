package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Amonkhet Raceway — Land:
//
//	"Start your engines!
//	 {T}: Add {C}.
//	 Max speed — {T}: Target creature gains haste until end of turn."
//
// ADR 0138 (#2122). The haste ability is an ordinary activated ability
// that can be activated only while its controller has max speed
// (CR 702.178a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fe174586-d36b-40f6-babd-1f98e76eec22",
		Name:            "Amonkhet Raceway",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		ManaAbilities: []ManaAbility{
			{Cost: ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "{T}: Add {C}."},
		},
		Activated: []ActivatedAbility{
			MaxSpeedActivated(ActivatedAbility{
				Label:   "Max speed — {T}: Target creature gains haste until end of turn.",
				Cost:    TapCost(),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"}, Label: "Amonkhet Raceway — haste"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			}),
		},
	})
}
