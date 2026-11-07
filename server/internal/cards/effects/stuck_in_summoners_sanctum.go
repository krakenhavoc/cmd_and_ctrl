package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stuck in Summoner's Sanctum — Enchantment — Aura {2}{U} (Final
// Fantasy, #4462):
//
//	"Flash
//	 Enchant artifact or creature
//	 When this Aura enters, tap enchanted permanent.
//	 Enchanted permanent doesn't untap during its controller's untap
//	 step and its activated abilities can't be activated."
//
// Marooned's enchant clause, Claustrophobia's entry tap and untap-step
// lock, and Arrest's activation lock. The lock covers mana abilities
// too (CantActivateMana), since a mana ability is an activated ability
// (CR 605.1a) and the card spares none. The entry tap reads the
// permanent the Aura enchants as the trigger resolves, or the one it
// last enchanted if the Aura has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "84d308ce-c3c5-4612-accc-6e6cc68a33be",
		Name:                  "Stuck in Summoner's Sanctum",
		Completeness:          CompletenessFull,
		PrintedKeywords:       []string{"flash"},
		Targets:               TargetPermanent("enchant artifact or creature", Or(Artifact(), Creature())),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantActivate | game.CantActivateMana),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Stuck in Summoner's Sanctum — tap enchanted permanent", tapEnchantedCreatureOnEntry),
		},
	})
}
