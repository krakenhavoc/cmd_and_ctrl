package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thunderclap Drake — Creature — Drake {1}{U}, 2/1 (Secrets of Strixhaven
// Commander, #4834):
//
//	"Flying
//	 Instant and sorcery spells you cast cost {1} less to cast.
//	 {2}{U}, Sacrifice this creature: When you next cast an instant or sorcery
//	 spell this turn, copy it for each time you've cast your commander from the
//	 command zone this game. You may choose new targets for the copies."
//
// The discount is Goblin Electromancer's reducer (CR 601.2f), which spends
// generic mana only and changes no mana value. The activation is
// Doublecast's CR 603.7 delayed trigger, "when you next cast", created as the
// ability resolves, with the sacrifice paid as a cost, so it is owed even
// though the Drake is long gone. It fires on the first instant or sorcery its
// controller casts this turn from any zone (CR 603.7b) and then ends with the
// turn if none comes. As it resolves it copies that spell once for each time
// its controller has cast a commander from the command zone this game (the
// tally sums over several commanders, as the rulings print), with the spell
// as last known if it has been countered (CR 608.2h). None cast, no copies:
// CopySpell reads a count of zero as one, so the body stops short. The
// copies are created on the stack, not cast (CR 707.10), and each may take
// new targets (CR 707.10c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "88d0a394-79e2-42ac-9a37-1a7b78231941",
		Name:            "Thunderclap Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				YourSpell(), InstantOrSorcerySpell()),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{U}, Sacrifice this creature: When you next cast an instant or sorcery spell this turn, copy it for each time you've cast your commander from the command zone this game. You may choose new targets for the copies.",
			Cost:  Plus(ManaCost("{2}{U}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return WhenYouNextCast(
					"Thunderclap Drake — copy that spell for each time you've cast your commander from the command zone",
					game.CastFilter{Types: []string{"Instant", "Sorcery"}},
					socB14CopyPerCommanderCastBody,
				).Apply(NewContext(g, item))
			},
		}},
	})
}
