package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sleep Magic — Enchantment — Aura {U} (Final Fantasy, #4462):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap
//	 step.
//	 When enchanted creature is dealt damage, sacrifice this Aura."
//
// Claustrophobia with a way out. The damage trigger watches damage dealt
// to the creature the Aura enchants (Pain for All's condition) from any
// source, combat or not; it sacrifices the Aura only while the Aura is
// still the permanent the trigger came from (CR 400.7), so a second
// damage event in the same moment does nothing more.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "33d6f4f0-e1ad-4363-a012-46a1bd9daeb7",
		Name:                  "Sleep Magic",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sleep Magic — tap enchanted creature", effects.TapEnchantedCreatureOnEntry),
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && source.AttachedTo.Kind == game.TargetCard && source.AttachedTo.ID == ev.Target
			}, "Sleep Magic — sacrifice this Aura", effects.Do(effects.Instruction(effects.SacrificeThisIfStillHere))),
		},
	})
}
