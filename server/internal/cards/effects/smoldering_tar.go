package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f268699f-c0a2-43f7-b933-7665599ab5ff",
		Name:         "Smoldering Tar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourUpkeep("Smoldering Tar — target player loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if p, ok := firstLegalPlayerTarget(ctx); ok {
						return GainLife{Player: p, Amount: -1}.Apply(ctx)
					}
					return nil
				}),
				TargetPlayer("target player")),
		},
		Activated: []ActivatedAbility{{
			Label:        "Sacrifice this enchantment: It deals 4 damage to target creature. Activate only as a sorcery.",
			Cost:         SacrificeThis(),
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature"),
			Purpose:      ForTargets(DamageToTarget(0, 4)),
			Effect:       sourceDealsDamageToEachLegalTarget(4),
		}},
	})
}
