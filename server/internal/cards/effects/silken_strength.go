package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silken Strength — Enchantment — Aura {1}{G} (Aetherdrift, #4464):
//
//	"Flash
//	 Enchant creature or Vehicle
//	 When this Aura enters, untap enchanted permanent.
//	 Enchanted permanent gets +1/+2 and has reach."
//
// Roadside Assistance's enchant clause and static shape with Freed from
// the Real's untapAttachedHost as the enters trigger: the host is found
// through the Aura as the trigger resolves, and an Aura that has fallen
// off by then untaps nothing. On an uncrewed Vehicle the +1/+2 and reach
// are there, and matter once it becomes a creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "53a59ba9-f4e9-460b-928a-ce6d7ae883f7",
		Name:            "Silken Strength",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         TargetPermanent("enchant creature or Vehicle", Or(Creature(), HasSubtype("Vehicle"))),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(1, 2),
			GrantToAttached("reach"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Silken Strength — untap enchanted permanent", untapAttachedHost),
		},
	})
}
