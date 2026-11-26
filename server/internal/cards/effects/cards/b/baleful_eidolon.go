package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baleful Eidolon — Enchantment Creature — Spirit {1}{B}, 1/1 (Theros):
//
//	"Bestow {4}{B} (If you cast this card for its bestow cost, it's an Aura
//	 spell with enchant creature. It becomes a creature again if it's not
//	 attached.)
//	 Deathtouch
//	 Enchanted creature gets +1/+1 and has deathtouch."
//
// Purphoros's Emissary's shape (ADR 0141): the creature, or the same bonus on
// a creature that already has a body, and the keyword on both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "0930a976-f69e-44a8-8b19-72095f20ed13",
		Name:             "Baleful Eidolon",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"deathtouch"},
		AlternativeCosts: []game.AlternativeCost{effects.Bestow("{4}{B}")},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("deathtouch"),
		},
	})
}
