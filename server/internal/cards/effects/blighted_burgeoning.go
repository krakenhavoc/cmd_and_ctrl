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
// Declared simplification shared with Wild Growth: the auto-tap planner
// does not count the extra mana (ADR 0074 §7). Weaker than printed and
// safe.
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
