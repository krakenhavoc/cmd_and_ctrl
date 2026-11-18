package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "0ccd1084-64bb-41ee-81c0-666af434b8df",
		Name:            "Storm Skreelix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				YourSpell(), InstantOrSorcerySpell()),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Storm Skreelix — gets +2/+0 until end of turn",
				thisGetsUntilEndOfTurn(2, 0, "Storm Skreelix — +2/+0")),
		},
	})
}
