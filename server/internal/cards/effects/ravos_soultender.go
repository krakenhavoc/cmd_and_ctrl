package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ravos, Soultender — Legendary Creature — Human Cleric {3}{W}{B}, 2/2
// (Commander 2016, #5537):
//
//	"Flying
//	 Other creatures you control get +1/+1.
//	 At the beginning of your upkeep, you may return target creature card
//	 from your graveyard to your hand.
//	 Partner (You can have two commanders if both have partner.)"
//
// The anthem is King Darien's (TribalAnthem over every other creature
// you control). The upkeep trigger is Reya Dawnbringer's shape returning
// to hand instead of the battlefield: the target is chosen as it goes on
// the stack (with no creature card in your graveyard it is removed, CR
// 603.3d) and the "you may" is asked as it resolves (CR 603.5). Partner
// is a deck-building rule (CR 702.124) that internal/deck reads off the
// oracle text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7880576b-8ad9-4382-97c5-7a631106a35e",
		Name:            "Ravos, Soultender",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				AtYourUpkeep("Ravos, Soultender — return target creature card from your graveyard to your hand",
					returnFirstLegalGraveyardTargetToHand),
				TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn())),
				"Ravos, Soultender — return the creature card to your hand?"),
		},
	})
}
