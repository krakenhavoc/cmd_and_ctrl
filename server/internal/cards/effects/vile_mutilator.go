package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "1aae80a5-079c-48f3-aad1-563931e1a031",
		Name:            "Vile Mutilator",
		Completeness:    CompletenessFull,
		AdditionalCost:  SacrificeCost("a creature or enchantment", Or(Creature(), Enchantment())),
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			TriggerEdict(WhenThisEnters("Vile Mutilator — each opponent sacrifices a nontoken enchantment, then a nontoken creature",
				Do(EachPlayerSacrifices{
					ExceptController: true,
					Match:            And(Enchantment(), Not(IsTokenPredicate())),
					Label:            "a nontoken enchantment",
					Keyed:            vileMutilatorCreatures,
				}))),
		},
	})
}

// vileMutilatorCreatures is "then sacrifices a nontoken creature of their
// choice", once every enchantment chosen has gone. An on-disk key.
var vileMutilatorCreatures = RegisterEdictThen("prompt-run/vile-mutilator-sacrifice-creatures", func(ctx *Context, _ game.PromptedSacrifices) error {
	return EachPlayerSacrifices{
		ExceptController: true,
		Match:            b04NontokenCreature,
		Label:            "a nontoken creature",
	}.Apply(ctx)
})
