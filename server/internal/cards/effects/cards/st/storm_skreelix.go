package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm Skreelix — Creature — Drake Horror {3}{U}{R}, 2/4 (Midnight Hunt, #5497):
//
//	"Flying
//	 Instant and sorcery spells you cast cost {1} less to cast.
//	 Whenever you cast an instant or sorcery spell, this creature gets +2/+0
//	 until end of turn."
//
// Archmage of Runes' discount and Ghitu Amplifier's pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0ccd1084-64bb-41ee-81c0-666af434b8df",
		Name:            "Storm Skreelix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.InstantOrSorcerySpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Storm Skreelix — gets +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Storm Skreelix — +2/+0")),
		},
	})
}
