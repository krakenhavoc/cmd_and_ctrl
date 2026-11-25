package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// God-Pharaoh's Statue — Legendary Artifact {6} (War of the Spark, #5520):
//
//	"Spells your opponents cast cost {2} more to cast.
//	 At the beginning of your end step, each opponent loses 1 life."
//
// Aura of Silence's tax over every spell an opponent casts (a cost increase,
// CR 601.2f), and Throne of the God-Pharaoh's end step drain: life loss, not
// damage, the same amount to each opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "183954d2-7b40-4c7e-adfa-d80191ab5ad6",
		Name:         "God-Pharaoh's Statue",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsMore(2, "Spells your opponents cast cost {2} more to cast.", effects.OpponentsSpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("God-Pharaoh's Statue — each opponent loses 1 life", effects.Do(effects.EachOpponentLosesLife{N: 1})),
		},
	})
}
