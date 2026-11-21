package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smaug, the Great Calamity // Spew Flame — an Adventure card (CR 715,
// The Hobbit, #4454), oracle fc75cfec…:
//
//	Smaug, the Great Calamity — Legendary Creature — Dragon {5}{R}{R}, 5/5
//	  "Flying"
//	Spew Flame — Sorcery — Adventure {4}{R}
//	  "Spew Flame deals 5 damage to target creature. (Then exile this
//	   card. You may cast the creature later from exile.)"
//
// A single creature target. A target that left in response counters the
// spell (CR 608.2b).
//
// No simplification.

// smaugTheGreatCalamityOracleID is shared by both faces of the card.
const SmaugTheGreatCalamityOracleID = "fc75cfec-e1f6-4cb5-8a3c-2f7b2c39c03f"

func init() {
	effects.Register(effects.Spec{
		OracleID:        SmaugTheGreatCalamityOracleID,
		Name:            "Smaug, the Great Calamity",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
	effects.Register(effects.Spec{
		OracleID:     SmaugTheGreatCalamityOracleID + "#1",
		Name:         "Spew Flame",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve:    effects.DamageToFirstTargetStep(5),
	})
}
