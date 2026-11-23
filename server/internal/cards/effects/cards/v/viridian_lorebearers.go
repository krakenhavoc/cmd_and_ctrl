package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "080c0557-59ee-4c46-afd9-5baabf0dc3d3",
		Name:         "Viridian Lorebearers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}, {T}: Target creature gets +X/+X until end of turn, where X is the number of artifacts your opponents control.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.ManaCost("{3}{G}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				x := effects.HobB7ArtifactsByOpponents(g, item.Controller)
				return effects.BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Viridian Lorebearers — +X/+X until end of turn"}.Apply(ctx)
			},
		}},
	})
}
