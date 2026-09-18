package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feather of Flight — Enchantment — Aura {1}{W} (EDHREC rank 4057):
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature gets +1/+0 and has flying."
//
// Angelic Gift with a point of power and, far more importantly,
// FLASH. The flash is what separates it from every other cantrip
// Aura: it comes down after blockers are declared, so the +1/+0 is a
// combat trick and the flying is an evasion grant you can hold up
// mana for. It is the reason a white deck plays a two-mana Aura that
// replaces itself at all.
//
// All four lines are live:
//
//   - Flash rides PrintedKeywords, and game.HasKeyword falls back to
//     the catalog's printed list for a card still in hand — which is
//     the only place a flash gate can be asked.
//   - Enchant creature is the shared EnchantCreature target clause,
//     re-run by the CR 704.5m legality check every time the board
//     changes, so the Aura falls off a creature that stops being one.
//   - The entry draw is a TRIGGER, not an AsEnters hook: "when this
//     Aura enters" uses the stack and can be responded to.
//   - The pump and the keyword are the shared attachment statics.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c61b05b8-00f1-4371-9856-c65852b4ca02",
		Name:            "Feather of Flight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Feather of Flight — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
