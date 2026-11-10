package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wolfwillow Haven — Enchantment — Aura {1}{G}:
//
//	"Enchant land
//	 Whenever enchanted land is tapped for mana, its controller adds an
//	 additional {G}.
//	 {4}{G}, Sacrifice this Aura: Create a 2/2 green Wolf creature token.
//	 Activate only during your turn."
//
// Wild Growth's triggered mana ability (a mana ability, CR 605.1b, so it
// never uses the stack) and a token maker with Castle Ardenvale's shape.
// The Aura is sacrificed as a cost; the token does not depend on it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7579f57d-c76e-4703-a030-34fb7160cb23",
		Name:         "Wolfwillow Haven",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		ManaTriggers: []game.ManaTrigger{
			WheneverAttachedTapsForMana("Wolfwillow Haven — add an additional {G}", "{G}"),
		},
		Activated: []ActivatedAbility{{
			Label:     "{4}{G}, Sacrifice this Aura: Create a 2/2 green Wolf creature token. Activate only during your turn.",
			Purpose:   game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:      Plus(ManaCost("{4}{G}"), SacrificeThis()),
			Condition: DuringYourTurn(),
			Effect:    cmdr2B01MakeToken("2/2 green Wolf"),
		}},
	})
}
