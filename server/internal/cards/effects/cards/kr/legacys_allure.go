package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Legacy's Allure — Enchantment {U}{U} (Tempest, S82 OLD batch 759, #5628):
//
//	"At the beginning of your upkeep, you may put a treasure counter on this
//	 enchantment.
//	 Sacrifice this enchantment: Gain control of target creature with power
//	 less than or equal to the number of treasure counters on this
//	 enchantment. (This effect lasts indefinitely.)"
//
// Barrin's Codex's optional upkeep counter. The target's power is compared
// with the Allure's treasure counters (game.SourcePowerVsCounters, #5628) as
// the ability is activated and again as it resolves, when the Allure has been
// sacrificed and its last-known counters are read (CR 608.2b, 608.2h). Control
// is for no stated duration (CR 611.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "106ca7bd-6dad-43fe-af58-8667ebddb2eb",
		Name:         "Legacy's Allure",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Legacy's Allure — put a treasure counter on this enchantment", effects.PutACounterOnThis("treasure")),
				"Legacy's Allure — put a treasure counter on it?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Gain control of target creature with power less than or equal to the number of treasure counters on this enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.SacrificeThis(),
			Targets: effects.RelativeToSource(
				effects.TargetCreature("target creature with power less than or equal to the number of treasure counters on this enchantment"),
				effects.S82OldB759PowerNoGreaterThanCounters("treasure")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GainControl{Target: id, Duration: game.IndefiniteDuration(),
					Label: "Legacy's Allure — gain control"}.Apply(ctx)
			},
		}},
	})
}
