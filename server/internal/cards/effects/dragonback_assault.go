package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonback Assault — Enchantment {3}{G}{U}{R} (Tarkir: Dragonstorm,
// #4463):
//
//	"When this enchantment enters, it deals 3 damage to each creature
//	 and each planeswalker.
//	 Landfall — Whenever a land you control enters, create a 4/4 red
//	 Dragon creature token with flying."
//
// The entry trigger is Brotherhood's End's sweep, dealt by the
// enchantment; landfall is the shared trigger, making the table's 4/4
// red flying Dragon.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "413fb2db-f1a1-4d22-ac37-a52821d35ca2",
		Name:         "Dragonback Assault",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dragonback Assault — 3 damage to each creature and each planeswalker",
				Do(Instruction(func(ctx *Context) error {
					return damageEachMatching(ctx, Or(Creature(), Planeswalker()), 3)
				}))),
			Landfall("Dragonback Assault — create a 4/4 red Dragon creature token with flying",
				Do(CreateToken{Template: TokenCard("4/4 red Dragon with flying"), N: 1})),
		},
	})
}
