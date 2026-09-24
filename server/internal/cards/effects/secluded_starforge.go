package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Secluded Starforge — Land:
//
//	"{T}: Add {C}.
//	 {2}, {T}, Tap X untapped artifacts you control: Target creature
//	 gets +X/+0 until end of turn. Activate only as a sorcery.
//	 {5}, {T}: Create a 2/2 colorless Robot artifact creature token."
//
// #1421 closes the middle ability's one missing shape: TapXUntapped
// marks the tap clause CountFromX, so the number of artifacts picked
// is the announcement carried on StackItem.XValue. The effect reads
// that same immutable X through Context.X; it never recounts the
// artifacts after they have tapped.
func init() {
	Register(Spec{
		OracleID:      "69f55a7c-6ddf-412e-b63b-b395731a1ff2",
		Name:          "Secluded Starforge",
		Completeness:  CompletenessFull,
		XMatters:      true,
		ManaAbilities: []ManaAbility{painlessColorless()},
		Activated: []ActivatedAbility{
			{
				Label:        "{2}, {T}, Tap X untapped artifacts you control: Target creature gets +X/+0 until end of turn.",
				Cost:         Plus(ManaCost("{2}"), TapCost(), TapXUntapped("X untapped artifacts you control", Artifact())),
				SorcerySpeed: true,
				Targets:      TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 {
						return nil
					}
					ctx := NewContext(g, item)
					return BoostUntilEOT{Target: item.Targets[0].ID, Power: ctx.X(), Label: "Secluded Starforge"}.Apply(ctx)
				},
			},
			{
				Label: "{5}, {T}: Create a 2/2 colorless Robot artifact creature token.",
				Cost:  Plus(ManaCost("{5}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: TokenCard("2/2 colorless Robot artifact"), N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
