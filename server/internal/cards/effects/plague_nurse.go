package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plague Nurse — Creature — Phyrexian Cleric {3}{G}, 3/4
// (Phyrexia: All Will Be One, #5479):
//
//	"Toxic 2
//	 {2}{G}: Each other creature you control with toxic gains toxic 1
//	 until end of turn. Activate only once each turn."
//
// Toxic is cumulative (CR 702.164b), so a creature with toxic 2 has a
// total of 3 for the turn. The creatures affected are those with toxic as
// the ability resolves (CR 611.2c); one that enters later gets nothing.
//
// No simplification.
func init() {
	const label = "{2}{G}: Each other creature you control with toxic gains toxic 1 until end of turn. Activate only once each turn."
	Register(Spec{
		OracleID:        "af7abd22-3134-4872-bbe2-0ad84742d9e5",
		Name:            "Plague Nurse",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"toxic 2"},
		Activated: []ActivatedAbility{{
			Label:     label,
			Cost:      ManaCost("{2}{G}"),
			Purpose:   game.Purpose{Answers: game.AnswerCombatGrant},
			Condition: OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				self := ctx.Source()
				return GrantKeywordUntilEOT{
					Match: func(_ *game.Game, caster uuid.UUID, c game.Card) bool {
						return c.IsCreature() && c.Controller == caster && c.InstanceID != self && game.ToxicTotal(&c) > 0
					},
					Keywords: []string{"toxic 1"},
					Label:    "Plague Nurse — toxic 1 until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
