package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Erratic Cyclops — Creature — Cyclops Shaman {3}{R}, 0/8 (Commander 2021,
// S82 OLD b058, #5504):
//
//	"Trample
//	 Whenever you cast an instant or sorcery spell, this creature gets +X/+0
//	 until end of turn, where X is that spell's mana value."
//
// A trigger on the cast. X is the spell's mana value read from last-known
// information when the trigger resolves (CR 608.2h), an {X} in its cost
// counted as the value chosen (CR 202.3e). The boost is on the creature that
// triggered, so a Cyclops that left does nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "19a11c33-269a-45d3-86c7-827965053f26",
		Name:            "Erratic Cyclops",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Erratic Cyclops — this creature gets +X/+0 until end of turn, where X is that spell's mana value",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x, ok := g.LastKnownSpellManaValueForEffect(ctx.Trigger().Event.CardID)
					if !ok || x <= 0 {
						return nil
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: x, Label: "Erratic Cyclops — +X/+0"}.Apply(ctx)
				}),
		},
	})
}
