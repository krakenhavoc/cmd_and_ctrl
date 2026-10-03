package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soratami Cloud Chariot — Artifact {5}:
//
//	"{2}: Target creature you control gains flying until end of turn.
//	 {2}: Prevent all combat damage that would be dealt to and dealt by
//	 target creature you control this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the second ability is one
// to-and-by record (Mod.AndDealtBy) pinned to the target.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "1a46373a-b48a-4118-92e0-d856ce409690",
		Name:         "Soratami Cloud Chariot",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}: Target creature you control gains flying until end of turn.",
				Cost:    ManaCost("{2}"),
				Targets: TargetCreature("target creature you control", YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					return GrantKeywordUntilEOT{Target: ts[0].ID, Keywords: []string{"flying"}, Label: "Soratami Cloud Chariot — flying"}.Apply(ctx)
				},
			},
			sourceShieldRow("{2}: Prevent all combat damage that would be dealt to and dealt by target creature you control this turn.",
				ManaCost("{2}"),
				TargetCreature("target creature you control", YouControl()),
				toAndByShield(ShieldTheTarget, true)),
		},
	})
}
