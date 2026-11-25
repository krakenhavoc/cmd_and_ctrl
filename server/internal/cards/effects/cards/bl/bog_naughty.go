package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bog Naughty — Creature — Faerie {3}{B}{B}, 3/3 (Throne of Eldraine,
// #5516):
//
//	"Flying
//	 {2}{B}, Sacrifice a Food: Target creature gets -3/-3 until end of
//	 turn."
//
// Any Food you control pays, a token or not (Bill the Pony's cost).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "74303ddb-d3b6-433d-a299-e0c4a7267461",
		Name:            "Bog Naughty",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice a Food: Target creature gets -3/-3 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -3, -3)),
			Cost:    game.AbilityCost{Mana: "{2}{B}", SacrificeOther: effects.SacrificeSpec("a Food", effects.HasSubtype("Food"))},
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: -3, Toughness: -3, Label: "Bog Naughty — -3/-3 until end of turn"}.Apply(ctx)
			},
		}},
	})
}
