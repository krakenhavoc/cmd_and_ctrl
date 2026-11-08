package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tractor Beam — Enchantment — Aura {2}{U}{U} (Edge of Eternities, #4461):
//
//	"Enchant creature or Spacecraft
//	 When this Aura enters, tap enchanted permanent.
//	 You control enchanted permanent.
//	 Enchanted permanent doesn't untap during its controller's untap step."
//
// Control Magic's layer-2 steal (ControlAttachedBySource) with
// Claustrophobia's enters tap and untap-step restriction. The enchant
// clause is Gravkill's "creature or Spacecraft": "Spacecraft" is the
// artifact subtype, so a Spacecraft below its station threshold (not yet
// a creature) can be enchanted, and the Aura stays on it (CR 704.5m runs
// the same clause). Its controller is the Aura's controller, so the
// stolen permanent stays tapped through that player's untap steps.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "ed99a0b8-d82d-4c80-af97-cabd333620df",
		Name:                  "Tractor Beam",
		Completeness:          CompletenessFull,
		Targets:               TargetPermanent("enchant creature or Spacecraft", Or(Creature(), HasSubtype("Spacecraft"))),
		Static:                []game.StaticAbility{ControlAttachedBySource()},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Triggered:             []game.TriggeredAbility{WhenThisEnters("Tractor Beam — tap enchanted permanent", tapEnchantedCreatureOnEntry)},
	})
}
