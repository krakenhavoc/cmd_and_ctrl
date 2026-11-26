package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Korozda Gorgon — Creature — Gorgon {3}{B}{G}, 2/5 (Dragon's Maze, S82 OLD batch
// 404, #5564):
//
//	"Deathtouch
//	 {2}, Remove a +1/+1 counter from a creature you control: Target creature gets
//	 -1/-1 until end of turn."
//
// Bolrac-Clan Crusher's cost, with the counter taken from any creature you control
// as the ability is activated (CR 602.2b), and a layer-7 -1/-1 on the target until
// cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e9e18a5a-d265-456f-a049-fa4b7e223263",
		Name:            "Korozda Gorgon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Remove a +1/+1 counter from a creature you control: Target creature gets -1/-1 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.RemoveCountersFrom(game.CounterPlusOne, 1, "a creature you control", effects.Creature())),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: -1, Toughness: -1, Label: "Korozda Gorgon — -1/-1 until end of turn"}.Apply(ctx)
			},
		}},
	})
}
