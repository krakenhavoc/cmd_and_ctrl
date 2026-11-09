package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Onakke Catacomb — Plane — Shandalar (Planechase Anthology, #4870):
//
//	"All creatures are black and have deathtouch.
//	 Whenever chaos ensues, creatures you control get +1/+0 and gain first
//	 strike until end of turn."
//
// A plane (ADR 0216 §6). The first static is a layer-5 colour set (the
// creatures are only black, the ruling) and a layer-6 keyword grant, both on
// battlefield creatures only. The chaos ability's set of creatures is locked
// as it resolves (CR 611.2c): one gained later gets nothing (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69f06edb-dade-4e94-b563-e36a05e567f0",
		Name:         "Onakke Catacomb",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label: "All creatures are black.",
				Layer: game.Layer5Color,
				AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
					return target.IsCreature()
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					c.Colors = []string{source.TextColor("B")}
				},
			},
			b9AllCreaturesHave("deathtouch"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Onakke Catacomb — creatures you control get +1/+0 and gain first strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return untilEndOfTurn(ctx, uuid.Nil, And(Creature(), YouControl()),
						"Onakke Catacomb — +1/+0 and first strike until end of turn",
						game.ModifyPTMod(1, 0), game.AddKeywordsMod("first strike"))
				}),
		},
	})
}
