package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blacksnag Buzzard — Creature — Bird {2}{B}, 2/1 (Outlaws of Thunder
// Junction, #4468):
//
//	"Flying
//	 This creature enters with a +1/+1 counter on it if a creature died
//	 this turn.
//	 Plot {1}{B} (You may pay {1}{B} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// The counter is a CR 614.1c self-replacement whose count is read as the
// Buzzard enters (b19EntersWithCountersCounted): one if the table-wide
// tally of creatures that died this turn is above zero
// (b11CreaturesDiedThisTurn, Dreaded Bat Cloud's read), any creature,
// anyone's. Plot is the ordinary special action (CR 702.170), and the
// free cast from exile enters through the same replacement.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "99da512c-e222-4e5c-aa64-a276ed9ebe87",
		Name:            "Blacksnag Buzzard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SpecialActions:  []game.SpecialAction{Plot("{1}{B}")},
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, _ *game.Card) int {
				if b11CreaturesDiedThisTurn(g) > 0 {
					return 1
				}
				return 0
			}, "Blacksnag Buzzard enters with a +1/+1 counter if a creature died this turn"),
		},
	})
}
