package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormcaller's Boon — Enchantment {2}{W}{U} (Alara Reborn, #5586):
//
//	"Sacrifice this enchantment: Creatures you control gain flying until end
//	 of turn.
//	 Cascade"
//
// Cascade is the printed keyword (Bloodbraid Elf's trigger). The grant covers
// the creatures you control as the ability resolves, locked in then
// (CR 611.2c): a creature that arrives later does not fly.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "28ebe75a-cddd-469b-be8a-286776b83388",
		Name:         "Stormcaller's Boon",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Cascade()},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Creatures you control gain flying until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Keywords: []string{"flying"}}, Answers: game.AnswerCombatGrant},
			Cost:    effects.SacrificeThis(),
			Effect: effects.Do(effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"flying"},
				Label:    "Stormcaller's Boon — creatures you control gain flying",
			}),
		}},
	})
}
