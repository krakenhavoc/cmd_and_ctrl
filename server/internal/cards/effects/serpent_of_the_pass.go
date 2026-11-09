package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Serpent of the Pass — Creature — Serpent {5}{U}{U}, 6/5 (Avatar: The
// Last Airbender, #4459):
//
//	"If there are three or more Lesson cards in your graveyard, you may
//	 cast this spell as though it had flash.
//	 This spell costs {1} less to cast for each noncreature, nonland card
//	 in your graveyard."
//
// The flash clause is Bard's Company's shape: the card's own timing
// permission (FlashIfYou, Spec.FlashIf, #5207), read as the cast begins
// (CR 601.3), and not an alternative cost, so a flash cast may still
// claim one (CR 118.9a). The discount is Ravenous Cosmozoan's (CR 601.2f):
// it counts the caster's graveyard as the cost is determined, applies to
// either cast, and reduces only the generic part, so {U}{U} is always
// paid (the ruling). The mana value stays 7 (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "98e36c6e-a868-4679-a5b5-6f02097f560b",
		Name:         "Serpent of the Pass",
		Completeness: CompletenessFull,
		FlashIf:      FlashIfYou("if there are three or more Lesson cards in your graveyard", avatarB6ThreeLessonsInYourGraveyard),
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(trekB17NoncreatureNonlandCardsInYourGraveyard,
				"This spell costs {1} less to cast for each noncreature, nonland card in your graveyard."),
		},
	})
}
