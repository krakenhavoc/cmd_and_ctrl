package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Council of Echoes — Creature — Spirit Advisor {4}{U}{U}, 4/4 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Flying
//	 Descend 4 — When this creature enters, if there are four or more
//	 permanent cards in your graveyard, return up to one target nonland
//	 permanent other than this creature to its owner's hand."
//
// Flying is printed. The enters trigger is a descend 4 intervening if
// (WhenThisEntersIfDescend, CR 603.4): with fewer than four permanent
// cards (CR 110.4) in its controller's graveyard as the Council enters it
// does not trigger, and the count is read again as it resolves. "Up to
// one" allows no target; "other than this creature" excludes this object
// (Another). The chosen permanent, if still legal, goes to its owner's
// hand (bounceEachLegalTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7b513bd0-27df-45f3-a85f-1f0aba3cae48",
		Name:            "Council of Echoes",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEntersIfDescend(4, "Council of Echoes — return up to one target nonland permanent to its owner's hand",
				bounceEachLegalTarget),
				Another(TargetPermanent("up to one target nonland permanent other than this creature", Nonland()).WithCount(0, 1))),
		},
	})
}
