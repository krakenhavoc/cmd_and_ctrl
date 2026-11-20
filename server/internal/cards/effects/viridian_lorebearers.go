package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Viridian Lorebearers — Creature — Elf Shaman {3}{G}, 3/3 (Fifth Dawn, #5604):
//
//	"{3}{G}, {T}: Target creature gets +X/+X until end of turn, where X
//	 is the number of artifacts your opponents control."
//
// Timberwatch Elf's +X/+X with Sardian Avenger's count: X is the artifacts
// the opponents control, counted as the ability resolves (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "080c0557-59ee-4c46-afd9-5baabf0dc3d3",
		Name:         "Viridian Lorebearers",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}, {T}: Target creature gets +X/+X until end of turn, where X is the number of artifacts your opponents control.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0)),
			Cost:    Plus(ManaCost("{3}{G}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				x := hobB7ArtifactsByOpponents(g, item.Controller)
				return BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Viridian Lorebearers — +X/+X until end of turn"}.Apply(ctx)
			},
		}},
	})
}
