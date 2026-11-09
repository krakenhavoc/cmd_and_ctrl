package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Omni-Changeling — Creature — Shapeshifter {3}{U}{U}, 0/0 (Lorwyn
// Eclipsed, #4458):
//
//	"Changeling (This card is every creature type.)
//	 Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 You may have this creature enter as a copy of any creature on the
//	 battlefield, except it has changeling."
//
// Clone with Mirrorhall Mimic's exception clause (EntersAsCopyOf, CR
// 707.9a): the copiable values gain the changeling keyword, a printed value
// like any other, so it is every creature type on the battlefield and a
// later copy of it copies that too. Changeling is a characteristic-defining
// ability and works in every zone, so the card itself is declared with it.
// Declined, or with nothing to copy, it is a 0/0 Shapeshifter with
// changeling and dies (CR 704.5f). Convoke is Appeal to Eirdu's TapCost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "66d4370e-19b9-4680-9b1e-a192537bc002",
		Name:            "Omni-Changeling",
		Completeness:    CompletenessFull,
		TapCost:         Convoke(),
		PrintedKeywords: []string{"changeling"},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf("Omni-Changeling", anyCreatureOnBattlefield,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddKeyword("changeling")
				}),
		},
	})
}
