package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cavern Lampad — Enchantment Creature — Nymph {3}{B}, 2/2 (Theros):
//
//	"Bestow {5}{B} (If you cast this card for its bestow cost, it's an Aura
//	 spell with enchant creature. It becomes a creature again if it's not
//	 attached.)
//	 Intimidate
//	 Enchanted creature gets +2/+2 and has intimidate."
//
// Purphoros's Emissary's shape (ADR 0141): the creature, or the same bonus on
// a creature that already has a body, and the keyword on both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "6ead91e1-4062-4080-bbc7-a4b62f8c6111",
		Name:             "Cavern Lampad",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"intimidate"},
		AlternativeCosts: []game.AlternativeCost{effects.Bestow("{5}{B}")},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("intimidate"),
		},
	})
}
