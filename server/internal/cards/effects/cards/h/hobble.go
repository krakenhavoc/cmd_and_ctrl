package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hobble — Enchantment — Aura {2}{W} (Planeshift, #5615):
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature can't attack.
//	 Enchanted creature can't block if it's black."
//
// The draw is the Aura controller's. "Can't attack" is Pacifism's restriction;
// "can't block if it's black" is the same restriction under a colour test read
// afresh on every recompute (Bilbo's Ring's conditional shape), so a creature
// that becomes or stops being black gains or loses the restriction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2398875-0912-4210-90bd-b2d568a580cf",
		Name:         "Hobble",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm, Draws: 1},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttack),
			effects.RestrictAttachedWhile(effects.B43AttachedHasColor("B"), game.CantBlock),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hobble — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
