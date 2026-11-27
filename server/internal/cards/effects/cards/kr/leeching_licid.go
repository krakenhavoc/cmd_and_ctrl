package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leeching Licid — Creature — Licid {1}{B}, 1/1 (Tempest, S82 OLD batch
// 758, #5628):
//
//	"{B}, {T}: This creature loses this ability and becomes an Aura
//	 enchantment with enchant creature. Attach it to target creature. You
//	 may pay {B} to end this effect.
//	 At the beginning of the upkeep of enchanted creature's controller, this
//	 creature deals 1 damage to that player."
//
// The first licid on the #5781 seam: LicidForm / LicidAbility / the
// enchant-creature ReturnedAura, and Soul Bleed's upkeep trigger (the
// upkeep of the controller of the creature it is attached to) dealing damage
// in place of a life loss. While it is a creature it enchants nothing, so the
// trigger never fires; paying {B} to end the effect is the engine's
// end_effect special action.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bcb5af89-a304-4de0-86af-b3f4cfabafc4",
		Name:         "Leeching Licid",
		Completeness: effects.CompletenessFull,
		ReturnedAura: effects.EnchantCreature(),
		Static:       effects.LicidForm(0),
		Activated:    []effects.ActivatedAbility{effects.LicidAbility("{B}", "{B}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Leeching Licid — 1 damage to that player",
				func(g *game.Game, item *game.StackItem) error {
					p := item.Trigger.Event.Actor
					if p == uuid.Nil {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: p, Amount: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
