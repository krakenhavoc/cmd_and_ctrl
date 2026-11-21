package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bound by Moonsilver — Enchantment — Aura {2}{W} (Innistrad Remastered):
//
//	"Enchant creature
//	 Enchanted creature can't attack, block, or transform.
//	 Sacrifice another permanent: Attach this Aura to target creature.
//	 Activate only as a sorcery and only once each turn."
//
// The static sets CantAttack, CantBlock and Characteristic.CantTransform on
// the host (the can't-transform seam of game/cant_transform.go, CR 101.2).
// The move is the Umbra-style attach (AttachSourceToTarget); the sacrifice is
// a cost of another permanent, so the Aura itself can't pay it.
func init() {
	const label = "Sacrifice another permanent: Attach this Aura to target creature. Activate only as a sorcery and only once each turn."
	Register(Spec{
		OracleID:     "7a98a1d1-7366-4b38-9047-48269790d328",
		Name:         "Bound by Moonsilver",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{{
			Label:     "Enchanted creature can't attack, block, or transform.",
			Layer:     game.Layer6Ability,
			AppliesTo: AttachedToSource,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
				c.CantTransform = true
			},
		}},
		Activated: []ActivatedAbility{{
			Label:        label,
			Cost:         SacrificeAnotherN(1, "another permanent"),
			SorcerySpeed: true,
			Condition:    OncePerTurnActivation(label),
			Targets:      TargetCreature("target creature"),
			Effect:       AttachSourceToTarget,
		}},
	})
}
