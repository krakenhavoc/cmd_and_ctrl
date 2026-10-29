package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const greatUglyLookingGoblinOracleID = "27e17542-549b-4c05-8091-c10a245c916b"

func init() {
	Register(Spec{
		OracleID:     greatUglyLookingGoblinOracleID,
		Name:         "Great Ugly-Looking Goblin",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return b16CreaturesYouControl(target, g, source) && target.Counters[game.CounterPlusOne] > 0
			}, "menace"),
		},
	})
	Register(Spec{
		OracleID:     greatUglyLookingGoblinOracleID + "#1",
		Name:         "Clap! Snap!",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Amass{Subtype: "Goblin", N: 2}.Apply(ctx)
		},
	})
}
