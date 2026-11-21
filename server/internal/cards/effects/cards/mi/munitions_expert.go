package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Munitions Expert — Creature — Goblin {B}{R}, 1/1 (Modern Horizons 3):
//
//	"Flash
//	 When this creature enters, you may have it deal damage to target creature
//	 or planeswalker equal to the number of Goblins you control."
//
// The Goblins are counted as the ability resolves, Munitions Expert included.
func init() {
	const label = "Munitions Expert — you may have it deal damage to target creature or planeswalker equal to the number of Goblins you control"
	enters := effects.Optional(effects.Targeting(effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		n := len(effects.PermanentsControlledByMatching(g, item.Controller, effects.Subtype("Goblin")))
		if n <= 0 {
			return nil
		}
		for _, t := range ctx.LegalTargets() {
			return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
		}
		return nil
	}), effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))),
		"Munitions Expert — have it deal damage equal to the number of Goblins you control?")
	effects.Register(effects.Spec{
		OracleID:        "a0b16b11-69d9-4e94-aae3-298aa44ce3d4",
		Name:            "Munitions Expert",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
