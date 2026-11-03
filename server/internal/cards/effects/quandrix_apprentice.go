package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quandrix Apprentice — Creature — Human Wizard {G}{U}, 2/2 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell, look
//	 at the top three cards of your library. You may reveal a land card from
//	 among them and put that card into your hand. Put the rest on the bottom
//	 of your library in any order."
//
// Duskwatch Recruiter's look-three-take-one body (socB10LookThreeTakeOne) with
// a land in place of a creature, behind the magecraft trigger. The take is
// optional and the rest go to the bottom in an order the player chooses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92116bbd-c7ad-4033-8d98-3ea09fea8ae7",
		Name:         "Quandrix Apprentice",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Magecraft("Quandrix Apprentice — look at the top three cards; you may reveal a land card and put it into your hand (magecraft)",
				socB10LookThreeTakeOne(Land(), "Quandrix Apprentice — you may reveal a land card from among them and put it into your hand")),
		},
	})
}
