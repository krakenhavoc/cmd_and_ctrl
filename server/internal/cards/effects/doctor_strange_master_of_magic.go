package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Doctor Strange, Master of Magic — Legendary Creature — Human Sorcerer
// Hero {2}{W}{U}, 2/3 (Secret Lair Countdown, #5463):
//
//	"Flying, vigilance, prowess
//	 {T}: Draw a card. Activate only if you've cast a noncreature spell
//	 this turn."
//
// The keywords are printed (prowess is the engine's keyword trigger).
// The draw is gated on the per-turn cast tally's noncreature count, taken
// as each spell became cast, so a noncreature spell that was since
// countered still counts. The {T} is the symbol, so summoning sickness
// applies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "554ac5df-b49c-4041-963d-408160f59fc1",
		Name:            "Doctor Strange, Master of Magic",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "prowess"},
		Activated: []ActivatedAbility{{
			Label:     "{T}: Draw a card. Activate only if you've cast a noncreature spell this turn.",
			Cost:      TapCost(),
			Condition: castANoncreatureSpellThisTurn,
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Effect:    Do(DrawCards{N: 1}),
		}},
	})
}

// castANoncreatureSpellThisTurn is "activate only if you've cast a
// noncreature spell this turn".
func castANoncreatureSpellThisTurn(g *game.Game, controller, _ uuid.UUID) bool {
	return g.CastTallyFor(controller).Noncreature > 0
}
