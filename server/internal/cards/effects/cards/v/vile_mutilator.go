package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vile Mutilator — Creature — Demon {5}{B}{B}, 6/5 (Duskmourn: House of
// Horror, #4466):
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 enchantment.
//	 Flying, trample
//	 When this creature enters, each opponent sacrifices a nontoken
//	 enchantment of their choice, then sacrifices a nontoken creature of
//	 their choice."
//
// Betrayer's Bargain's "sacrifice a creature or enchantment" as the
// whole additional cost, paid while casting (CR 601.2h), so a Mutilator
// put onto the battlefield by another effect costs nothing and still
// triggers.
//
// The enters trigger is Smallpox's chained edicts: each opponent
// chooses a nontoken enchantment (APNAP, each seeing the earlier
// choices, all sacrificed together), and once those have gone the same
// for a nontoken creature, through a registered continuation
// (vileMutilatorCreatures) so a table waiting on either choice is a
// restore point. An opponent with nothing to sacrifice is not asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1aae80a5-079c-48f3-aad1-563931e1a031",
		Name:            "Vile Mutilator",
		Completeness:    effects.CompletenessFull,
		AdditionalCost:  effects.SacrificeCost("a creature or enchantment", effects.Or(effects.Creature(), effects.Enchantment())),
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.WhenThisEnters("Vile Mutilator — each opponent sacrifices a nontoken enchantment, then a nontoken creature",
				effects.Do(effects.EachPlayerSacrifices{
					ExceptController: true,
					Match:            effects.And(effects.Enchantment(), effects.Not(effects.IsTokenPredicate())),
					Label:            "a nontoken enchantment",
					Keyed:            vileMutilatorCreatures,
				}))),
		},
	})
}

// vileMutilatorCreatures is "then sacrifices a nontoken creature of their
// choice", once every enchantment chosen has gone. An on-disk key.
var vileMutilatorCreatures = effects.RegisterEdictThen("prompt-run/vile-mutilator-sacrifice-creatures", func(ctx *effects.Context, _ game.PromptedSacrifices) error {
	return effects.EachPlayerSacrifices{
		ExceptController: true,
		Match:            effects.B04NontokenCreature,
		Label:            "a nontoken creature",
	}.Apply(ctx)
})
