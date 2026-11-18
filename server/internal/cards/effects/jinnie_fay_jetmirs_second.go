package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jinnie Fay, Jetmir's Second — Legendary Creature — Elf Druid
// {R/G}{G}{G/W}, 3/3:
//
//	"If you would create one or more tokens, you may instead create that
//	 many 2/2 green Cat creature tokens with haste or that many 3/1 green
//	 Dog creature tokens with vigilance."
//
// A "may" replacement (CR 614.1a) on the CR 701.7b creation instruction,
// the event Moonlit Meditation rewrites. The engine's replacement question
// is yes or no, so the one printed effect is two optional replacements, one
// per outcome, that exclude each other: once an instruction's groups are
// Jinnie's tokens neither applies again, so the player picks Cats, Dogs, or
// neither, and each group keeps its count and its entry clause (tapped,
// attacking) while its kind is swapped (ReplaceTokenKinds). Her controller's
// tokens only, whatever makes them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61fbaaf2-4286-4e9a-b9cb-aa31262b596a",
		Name:         "Jinnie Fay, Jetmir's Second",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			s82SNC22JinnieInstead("Jinnie Fay: create 2/2 green Cats with haste instead",
				"Jinnie Fay — create that many 2/2 green Cat tokens with haste instead?", "2/2 green Cat with haste"),
			s82SNC22JinnieInstead("Jinnie Fay: create 3/1 green Dogs with vigilance instead",
				"Jinnie Fay — create that many 3/1 green Dog tokens with vigilance instead?", "3/1 green Dog with vigilance"),
		},
	})
}
