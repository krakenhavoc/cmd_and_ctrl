package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bygone Bishop — Creature — Spirit Cleric {2}{W}, 2/3:
//
//	"Flying
//	 Whenever you cast a creature spell with mana value 3 or less,
//	 investigate."
//
// The trigger is raised on the cast, so it resolves ahead of the spell.
// Mana value is read through ManaValueLE (CR 202.3), so an {X} spell is
// judged by the value chosen for it on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a47d9ea3-95fe-4774-8e72-30e7f56756f4",
		Name:            "Bygone Bishop",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.And(effects.Creature(), effects.ManaValueLE(3)),
				"Bygone Bishop — investigate",
				effects.Do(effects.Investigate{N: 1})),
		},
	})
}
