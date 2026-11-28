package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Briar Shield — Enchantment — Aura {G} (Weatherlight):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1.
//	 Sacrifice this Aura: Enchanted creature gets +3/+3 until end of turn."
//
// The sacrifice is the cost, so the permanent +1/+1 is gone before the +3/+3
// resolves; the +3/+3 goes to the creature the Aura was attached to as it left
// (CR 608.2h), if that creature is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d437391-d476-4976-8714-6a1d3185ab52",
		Name:         "Briar Shield",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this Aura: Enchanted creature gets +3/+3 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.SacrificeThis(),
			Effect:  effects.OldB774EnchantedCreatureGetsUntilEOT(3, 3, "Briar Shield — +3/+3"),
		}},
	})
}
