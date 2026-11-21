package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Potioner's Trove — Artifact {3} (Secrets of Strixhaven, #4456):
//
//	"{T}: Add one mana of any color.
//	 {T}: You gain 2 life. Activate only if you've cast an instant or
//	 sorcery spell this turn."
//
// The mana is Alloy Myr's ability. The life ability is Sanar's condition
// (sanarCastAnInstantOrSorceryThisTurn): the per-turn cast tally's instant
// and sorcery count, taken as each spell was cast (CR 601.2i), so a spell
// that has since been countered still counts. The {T} is the symbol, so the
// Trove's two abilities can't both be used in one turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d604c75e-9d81-49de-80bc-08f4ef3ee975",
		Name:         "Potioner's Trove",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: You gain 2 life. Activate only if you've cast an instant or sorcery spell this turn.",
			Cost:      effects.TapCost(),
			Condition: effects.SanarCastAnInstantOrSorceryThisTurn,
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Effect:    effects.Do(effects.GainLife{Amount: 2}),
		}},
	})
}
