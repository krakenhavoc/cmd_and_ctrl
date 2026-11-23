package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seething Pathblazer — Creature — Elemental Warrior {2}{R}, 2/2
// (Morningtide, #5592):
//
//	"Sacrifice an Elemental: This creature gets +2/+0 and gains first
//	 strike until end of turn."
//
// The sacrifice is a cost (SacrificeN over effective creature types, so a
// changeling counts), and "an Elemental" includes the Pathblazer itself,
// which is then gone before the ability resolves and gets nothing
// (CR 608.2b). The body is the self pump and the self keyword grant, both
// ending at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "574abab1-a722-45f5-a210-c1d4442a278a",
		Name:         "Seething Pathblazer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "Sacrifice an Elemental: This creature gets +2/+0 and gains first strike until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerPump | game.AnswerCombatGrant | game.AnswerSacOutlet,
				Pump:    &game.Pump{Power: 2, Keywords: []string{"first strike"}},
			},
			Cost:   effects.SacrificeN(1, "an Elemental", effects.OfCreatureType("Elemental")),
			Effect: seethingPathblazerPump,
		}},
	})
}

// seethingPathblazerPump is "this creature gets +2/+0 and gains first
// strike until end of turn".
func seethingPathblazerPump(g *game.Game, item *game.StackItem) error {
	if err := effects.ThisGetsUntilEndOfTurn(2, 0, "Seething Pathblazer — +2/+0")(g, item); err != nil {
		return err
	}
	return effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Seething Pathblazer — first strike")(g, item)
}
