package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resilient Roadrunner — Creature — Bird {1}{R}, 2/2 (Outlaws of Thunder
// Junction, #4468):
//
//	"Haste, protection from Coyotes
//	 {3}: This creature can't be blocked this turn except by creatures
//	 with haste."
//
// Protection from Coyotes is a subtype quality (CR 702.16, ADR 0072),
// Yawgmoth, Thran Physician's "protection from Humans" grammar: no
// Coyote can block it, and no Coyote, or ability of one, can damage or
// target it. The {3} is Gingerbrute's until-end-of-turn block
// rule (CantBeBlockedThisTurnExceptBy), pinned to this object as it
// resolves; "creatures with haste" reads the blocker's effective
// abilities, so a granted haste counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "78c7c73f-847d-43bf-9e02-cf85357453f4",
		Name:            "Resilient Roadrunner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste", "protection from Coyotes"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: This creature can't be blocked this turn except by creatures with haste.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CantBeBlockedThisTurnExceptBy{
					Target:   item.SourceCardID,
					Keywords: []string{"haste"},
					Text:     "creatures with haste",
					Label:    "Resilient Roadrunner — can't be blocked except by creatures with haste",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
