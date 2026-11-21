package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Great Ugly-Looking Goblin // Clap! Snap! — an Adventure card (CR 715,
// The Hobbit, #4454), oracle 27e17542…:
//
//	Great Ugly-Looking Goblin — Creature — Goblin Soldier {5}{B}, 4/4
//	  "Each creature you control with a +1/+1 counter on it has menace."
//	Clap! Snap! — Sorcery — Adventure {1}{B}
//	  "Amass Goblins 2. (Then exile this card. You may cast the creature
//	   later from exile.)"
//
// The menace grant is a Layer 6 keyword grant over every creature its
// controller controls (itself included, as printed) that carries a
// +1/+1 counter. It reads the counter map, which the layer pass
// recomputes on every counter change, so a creature gains menace the
// moment its first counter lands and loses it with the last.
//
// No simplification.

// greatUglyLookingGoblinOracleID is shared by both faces of the card.
const GreatUglyLookingGoblinOracleID = "27e17542-549b-4c05-8091-c10a245c916b"

func init() {
	effects.Register(effects.Spec{
		OracleID:     GreatUglyLookingGoblinOracleID,
		Name:         "Great Ugly-Looking Goblin",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.B16CreaturesYouControl(target, g, source) && target.Counters[game.CounterPlusOne] > 0
			}, "menace"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     GreatUglyLookingGoblinOracleID + "#1",
		Name:         "Clap! Snap!",
		Purpose:      game.Purpose{Amass: 2},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Amass{Subtype: "Goblin", N: 2}.Apply(ctx)
		},
	})
}
