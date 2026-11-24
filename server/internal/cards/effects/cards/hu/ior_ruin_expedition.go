package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ior Ruin Expedition — Enchantment {1}{U} (Commander Legends):
//
//	"Landfall — Whenever a land you control enters, you may put a quest counter
//	 on this enchantment.
//	 Remove three quest counters from this enchantment and sacrifice it: Draw
//	 two cards."
//
// Beastmaster Ascension's optional quest-counter trigger over a land entering
// under your control, and an activated ability whose costs are removing three
// counters and sacrificing the enchantment itself, paid as it is announced, so
// it needs the three counters and can be responded to.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc338ba2-c3a4-45d6-895f-af7f72bf7a94",
		Name:         "Ior Ruin Expedition",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Landfall("Ior Ruin Expedition — landfall: put a quest counter on it", func(g *game.Game, item *game.StackItem) error {
				return effects.AddCounter{Target: item.SourceCardID, Kind: "quest", N: 1}.Apply(effects.NewContext(g, item))
			}), "Ior Ruin Expedition — put a quest counter on it?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove three quest counters from this enchantment and sacrifice it: Draw two cards.",
			Purpose: game.Purpose{Draws: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.RemoveCountersFromThis("quest", 3), game.AbilityCost{SacrificeSelf: true}),
			Effect:  effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}
