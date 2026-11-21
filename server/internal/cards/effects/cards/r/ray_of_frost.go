package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ray of Frost — Enchantment — Aura {1}{U}:
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, if enchanted creature is red, tap it.
//	 As long as enchanted creature is red, it loses all abilities.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Three independent clauses. The entry trigger has an intervening "if"
// (CR 603.4), asked as it triggers and again as it resolves: the enchanted
// creature is red then (through the layers). "Loses all abilities" is
// conditional on the colour, read live in layer 6 after colours (layer 5)
// have been applied, so turning the creature to or from red switches it on
// and off while the Aura stays. "Doesn't untap" is unconditional, for any
// colour of creature, and ends with the Aura.
//
// No simplification.
func init() {
	loses := effects.LoseAllAbilities()
	loses.Label = "As long as enchanted creature is red, it loses all abilities."
	loses.AppliesTo = effects.B43AttachedHasColor("R")
	loses.ContinuesAfterRemoval = false
	etb := effects.InterveningIf(
		effects.CondBoard{Name: "enchanted creature is red", Test: rayOfFrostEnchantedIsRed},
		effects.WhenThisEnters("Ray of Frost — if enchanted creature is red, tap it", effects.TapEnchantedCreatureOnEntry),
	)
	effects.Register(effects.Spec{
		OracleID:              "3de32046-454b-40be-b522-a3d1d5302b10",
		Name:                  "Ray of Frost",
		Completeness:          effects.CompletenessFull,
		PrintedKeywords:       []string{"flash"},
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Static:                []game.StaticAbility{loses},
		Triggered:             []game.TriggeredAbility{etb},
	})
}

// rayOfFrostEnchantedIsRed is the entry trigger's condition: the creature
// the Aura enchants is red.
func rayOfFrostEnchantedIsRed(ctx *effects.Context) bool {
	aura, ok := ctx.TriggeringPermanent()
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return false
	}
	c, ok := ctx.Game.LookupCardForEffect(aura.AttachedTo.ID)
	if !ok {
		return false
	}
	red := ctx.TextColor("R")
	for _, col := range c.EffectiveColors() {
		if col == red {
			return true
		}
	}
	return false
}
