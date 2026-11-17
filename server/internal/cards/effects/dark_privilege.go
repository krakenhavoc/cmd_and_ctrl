package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dark Privilege — Enchantment — Aura {1}{B} (Multiverse Gift Box, #5632):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1.
//	 Sacrifice a creature: Regenerate enchanted creature."
//
// PumpAttached for the bonus, and an activated ability of the Aura's that
// targets nothing and finds the creature through the attachment, as Crab
// Umbra's untap does. The creature sacrificed may be the enchanted one
// itself (it is a creature you control when the Aura's controller controls
// it), which leaves nothing to regenerate. The shield (CR 701.19a) goes on
// whatever the Aura enchants as the ability resolves; an Aura that has
// gone by then enchants nothing and the ability does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "29f4c69a-ca59-4e9f-a808-87ba85b34f2f",
		Name:         "Dark Privilege",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a creature: Regenerate enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    SacrificeACreature(),
			Effect:  regenerateAttachedHost,
		}},
	})
}

// regenerateAttachedHost is "Regenerate enchanted creature": a shield on
// the Aura's host as the ability resolves.
func regenerateAttachedHost(g *game.Game, item *game.StackItem) error {
	host := attachedHostFor(g, item.SourceCardID)
	if host == nil {
		return nil
	}
	return Regenerate{Target: host.InstanceID}.Apply(NewContext(g, item))
}
