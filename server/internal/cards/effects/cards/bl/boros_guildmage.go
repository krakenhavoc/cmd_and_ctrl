package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boros Guildmage — Creature — Human Wizard {R/W}{R/W}, 2/2 (Ravnica:
// City of Guilds, #5600):
//
//	"{1}{R}: Target creature gains haste until end of turn.
//	 {1}{W}: Target creature gains first strike until end of turn."
//
// Axgard Cavalry's grant twice, each on its own mana cost and on any
// controller's creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "266c2f49-82eb-48e9-a7ea-e0da787fa40e",
		Name:         "Boros Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{R}: Target creature gains haste until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste"))},
				Cost:    effects.ManaCost("{1}{R}"),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.EbTargetCreatureGainsUntilEOT("haste", "Boros Guildmage — haste until end of turn"),
			},
			{
				Label:   "{1}{W}: Target creature gains first strike until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike"))},
				Cost:    effects.ManaCost("{1}{W}"),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.EbTargetCreatureGainsUntilEOT("first strike", "Boros Guildmage — first strike until end of turn"),
			},
		},
	})
}
