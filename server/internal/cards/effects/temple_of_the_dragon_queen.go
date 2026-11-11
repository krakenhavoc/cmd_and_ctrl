package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Temple of the Dragon Queen — Land (Tarkir: Dragonstorm Commander):
//
//	"As this land enters, you may reveal a Dragon card from your hand. This
//	 land enters tapped unless you revealed a Dragon card this way or you
//	 control a Dragon.
//	 As this land enters, choose a color.
//	 {T}: Add one mana of the chosen color."
//
// The first line is the reveal-land replacement (a free reveal, nothing moves,
// CR 701.20b) that does not apply while you control a Dragon, so with one out
// the land enters untapped and nobody is asked. A "Dragon card" is a card with
// the creature type Dragon in its type line (the ruling), judged in hand as
// printed. The colour is an as-enters choice (Valgavoth's Lair's), and the
// mana ability reads the stored colour at activation.
//
// No simplification.
func init() {
	const name = "Temple of the Dragon Queen"
	Register(Spec{
		OracleID:     "169a26d2-7bc9-4403-9c92-98d4bd5ca4f3",
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{cmdrB13EntersTappedUnlessRevealOrControlDragon(name)},
		AsEnters:     ChooseColorAsEnters(game.ColorForMana, name),
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedChosenColor(),
			Label:        "Add one mana of the chosen color",
		}},
	})
}
