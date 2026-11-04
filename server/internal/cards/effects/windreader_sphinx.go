package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Windreader Sphinx — Creature — Sphinx {5}{U}{U}, 3/7 (Jumpstart 2020):
//
//	"Flying
//	 Whenever a creature with flying attacks, you may draw a card."
//
// One trigger per attacking creature that has flying as it is declared,
// whoever controls it: an opponent's flier attacking triggers it too, as
// printed (jmpB9AFlierAttacked). The "you may" is asked as it resolves
// (OptionalDraw).
//
// No simplification. Deck request #4894 (Jumpstart Azorius).
func init() {
	Register(Spec{
		OracleID:        "5e038a3a-fbab-46c2-8773-1327d953da16",
		Name:            "Windreader Sphinx",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OptionalDraw(On(game.EventAttack, jmpB9AFlierAttacked,
				"Windreader Sphinx — a creature with flying attacked: you may draw a card",
				Do(DrawCards{N: 1})), "Windreader Sphinx — draw a card?", 1),
		},
	})
}
