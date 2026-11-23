package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smoldering Tar — Enchantment {2}{B}{R} (Invasion, #5616):
//
//	"At the beginning of your upkeep, target player loses 1 life.
//	 Sacrifice this enchantment: It deals 4 damage to target creature.
//	 Activate only as a sorcery."
//
// The upkeep trigger targets any player, its controller included (CR
// 115.1d), and the loss is life loss, not damage. The sacrifice row is
// Seal of Fire's body at 4 on a creature, at sorcery speed (CR 602.5d).
// The enchantment is the damage's source, read as it last existed once
// it has been sacrificed (CR 113.7a). A target gone by resolution gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f268699f-c0a2-43f7-b933-7665599ab5ff",
		Name:         "Smoldering Tar",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourUpkeep("Smoldering Tar — target player loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if p, ok := effects.FirstLegalPlayerTarget(ctx); ok {
						return effects.GainLife{Player: p, Amount: -1}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetPlayer("target player")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "Sacrifice this enchantment: It deals 4 damage to target creature. Activate only as a sorcery.",
			Cost:         effects.SacrificeThis(),
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature"),
			Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
			Effect:       effects.SourceDealsDamageToEachLegalTarget(4),
		}},
	})
}
