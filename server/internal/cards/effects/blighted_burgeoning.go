package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blighted Burgeoning — Enchantment — Aura {2}{G} (March of the Machine):
//
//	"Enchant land
//	 When this Aura enters, incubate 2.
//	 Whenever enchanted land is tapped for mana, its controller adds an
//	 additional one mana of any color."
//
// The Wild Growth shape (a mana trigger on the attached land) with an
// enters trigger on top. The incubate is the Aura's controller's, made
// when the trigger resolves.
//
// The auto-tap planner counts the extra mana (ADR 0074, amendment of
// 2026-10-30): a land under it plans as two mana, the second a pick of
// any colour that the plan books like any other slot.
//
// No simplification of the card itself.
func init() {
	Register(Spec{
		OracleID:     "5e10fd35-9768-48e4-8cdb-95224ad33b9b",
		Name:         "Blighted Burgeoning",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		ManaTriggers: []game.ManaTrigger{
			WheneverAttachedTapsForMana("Blighted Burgeoning — add an additional one mana of any color", "{W|U|B|R|G}"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Blighted Burgeoning — incubate 2", Do(Incubate{N: 2})),
		},
	})
}
