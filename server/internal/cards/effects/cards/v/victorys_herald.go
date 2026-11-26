package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Victory's Herald — Creature — Angel {3}{W}{W}{W}, 4/4 (Commander 2015, S82
// OLD #5546):
//
//	"Flying
//	 Whenever this creature attacks, attacking creatures gain flying and
//	 lifelink until end of turn."
//
// Chieftain en-Dal's attack trigger with two keywords: the attacking
// creatures are read once as it resolves (CR 611.2c), whoever controls them,
// and keep both until the cleanup step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e410f80f-cf67-46ff-aeb3-da0204237d66",
		Name:            "Victory's Herald",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Victory's Herald — attacking creatures gain flying and lifelink until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Match:    effects.AttackingCreature(),
						Keywords: []string{"flying", "lifelink"},
						Label:    "Victory's Herald — flying and lifelink until end of turn",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
