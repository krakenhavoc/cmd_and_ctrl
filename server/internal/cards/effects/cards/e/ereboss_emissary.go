package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Erebos's Emissary — Enchantment Creature — Snake {3}{B}, 3/3 (Theros):
//
//	"Bestow {5}{B} (If you cast this card for its bestow cost, it's an Aura
//	 spell with enchant creature. It becomes a creature again if it's not
//	 attached.)
//	 Discard a creature card: This creature gets +2/+2 until end of turn. If
//	 this permanent is an Aura, enchanted creature gets +2/+2 until end of turn
//	 instead.
//	 Enchanted creature gets +3/+3."
//
// Everflame Eidolon's pump of itself or its host, read as the ability
// resolves (CR 608.2h): a bestowed Aura boosts the creature it enchants and
// anything else boosts the Emissary itself. The discard is a cost.
//
// No simplification.
func init() {
	const label = "Erebos's Emissary — +2/+2 until end of turn"
	effects.Register(effects.Spec{
		OracleID:         "dc39c52b-9bee-4cec-a970-1867314944ef",
		Name:             "Erebos's Emissary",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Bestow("{5}{B}")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a creature card: This creature gets +2/+2 until end of turn. If this permanent is an Aura, enchanted creature gets +2/+2 until end of turn instead.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.DiscardCardsMatching(1, "a creature card", effects.MatchCreature),
			Effect:  effects.S82OldB377PumpSelfOrEnchanted(label, 2, 2),
		}},
		Static: []game.StaticAbility{effects.PumpAttached(3, 3)},
	})
}
